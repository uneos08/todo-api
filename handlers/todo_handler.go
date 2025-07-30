package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"todo-api/auth"
	"todo-api/models"
	"todo-api/store"

	"path/filepath"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type TodoHandler struct {
	Store store.TodoStore
}

func NewTodoHandler(store store.TodoStore) *TodoHandler {
	return &TodoHandler{Store: store}
}

func (h *TodoHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/todos", h.handleTodos).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/api/todos/{id}", h.handleTodoByID).Methods(http.MethodPut, http.MethodDelete)
}

func (h *TodoHandler) handleTodos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getTodos(w, r)
	case http.MethodPost:
		h.createTodo(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *TodoHandler) handleTodoByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/todos/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeGeneralResponse(w, "error", "Invalid ID", nil, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getTodoByID(w, r, id)
	case http.MethodPut:
		h.updateTodo(w, r, id)
	case http.MethodDelete:
		h.deleteTodo(w, id)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// writeGeneralResponse — универсальный ответ
func writeGeneralResponse(w http.ResponseWriter, status, message string, data any, httpStatus int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	// Если data == nil, и это слайс — вернуть пустой массив
	if slice, ok := data.([]models.Todo); ok && slice == nil {
		data = []models.Todo{}
	}

	resp := models.GeneralResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(resp)
}

// @Summary      Get all todos
// @Description  Получить список всех задач
// @Tags         todos
// @Produce      json
// @Success      200  {object}  models.GeneralResponse{data=[]models.Todo}
// @Failure      500  {object}  models.GeneralResponse
// @Router       /todos [get]
func (h *TodoHandler) getTodos(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ExtractClaimsFromRequest(r)
	if err != nil {
		writeGeneralResponse(w, "error", "Unauthorized", nil, http.StatusUnauthorized)
		return
	}

	todos, err := h.Store.GetTodos(claims.UserID)
	if err != nil {
		log.Printf("❌ Failed to fetch todos from DB: %v", err)
		writeGeneralResponse(w, "error", "Failed to fetch todos", nil, http.StatusInternalServerError)
		return
	}

	writeGeneralResponse(w, "success", "Todos fetched", todos, http.StatusOK)
}

// @Summary      Create a new todo
// @Description  Создать новую задачу
// @Tags         todos
// @Accept       json
// @Produce      json
// @Param        todo  body      models.Todo        true  "Todo data"
// @Success      201   {object}  models.GeneralResponse{data=models.Todo}
// @Failure      400   {object}  models.GeneralResponse
// @Failure      500   {object}  models.GeneralResponse
// @Router       /todos [post]
func (h *TodoHandler) createTodo(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // максимум 10MB
	if err != nil {
		log.Fatalf("❌ Failed to create uploads folder: %v", err)
		writeGeneralResponse(w, "error", "Failed to parse form", nil, http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	doneStr := r.FormValue("done")
	done := doneStr == "true" || doneStr == "1"

	// Получаем файл
	file, handler, err := r.FormFile("photo")
	var photoURL *string
	if err == nil {
		defer file.Close()

		ext := filepath.Ext(handler.Filename)
		filename := uuid.New().String() + ext
		filePath := "./uploads/" + filename

		outFile, err := os.Create(filePath)
		if err != nil {
			writeGeneralResponse(w, "error", "Failed to save file", nil, http.StatusInternalServerError)
			return
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, file)
		if err != nil {
			writeGeneralResponse(w, "error", "Failed to save file content", nil, http.StatusInternalServerError)
			return
		}

		url := "http://localhost:8080/uploads/" + filename
		photoURL = &url
	}

	// Получить userID из токена (пример)
	userID, err := auth.ExtractUserIDFromRequest(r)
	if err != nil {
		writeGeneralResponse(w, "error", "Unauthorized", nil, http.StatusUnauthorized)
		return
	}

	todo := models.Todo{
		Title:    title,
		Done:     done,
		UserID:   userID,
		PhotoURL: photoURL,
	}

	created, err := h.Store.CreateTodo(todo)
	if err != nil {
		writeGeneralResponse(w, "error", "Failed to create todo", nil, http.StatusInternalServerError)
		return
	}
	writeGeneralResponse(w, "success", "Todo created", created, http.StatusCreated)
}

// @Summary      Update a todo by ID
// @Description  Обновить задачу по ID
// @Tags         todos
// @Accept       json
// @Produce      json
// @Param        id    path      int           true  "Todo ID"
// @Param        todo  body      models.Todo   true  "Updated todo data"
// @Success      200   {object}  models.GeneralResponse{data=models.Todo}
// @Failure      400   {object}  models.GeneralResponse
// @Failure      404   {object}  models.GeneralResponse
// @Router       /todos/{id} [put]
func (h *TodoHandler) updateTodo(w http.ResponseWriter, r *http.Request, id int) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		writeGeneralResponse(w, "error", "Failed to parse form", nil, http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	doneStr := r.FormValue("done")
	done := doneStr == "true" || doneStr == "1"

	existingTodo, err := h.Store.GetTodoByID(id)
	if err != nil {
		writeGeneralResponse(w, "error", "Todo not found", nil, http.StatusNotFound)
		return
	}

	photoURL := existingTodo.PhotoURL

	file, handler, err := r.FormFile("photo")
	if err == nil {
		defer file.Close()
		log.Printf("⚙️ Existing photoURL: %v", photoURL)

		// Удаляем старый файл, если он есть
		if photoURL != nil && *photoURL != "" {
			parsedURL, err := url.Parse(*photoURL)
			if err != nil {
				log.Printf("❌ Invalid photo URL: %v", err)
			} else {
				// parsedURL.Path, например: /uploads/filename.jpg
				filePath := "." + filepath.Clean(parsedURL.Path)

				log.Printf("🗑️ Removing old photo file: %s", filePath)

				err = os.Remove(filePath)
				if err != nil && !os.IsNotExist(err) {
					log.Printf("❌ Failed to delete old photo: %v", err)
				} else {
					log.Printf("✅ Old photo deleted")
				}
			}
		}

		ext := filepath.Ext(handler.Filename)
		filename := uuid.New().String() + ext
		filePath := "./uploads/" + filename

		outFile, err := os.Create(filePath)
		if err != nil {
			writeGeneralResponse(w, "error", "Failed to save file", nil, http.StatusInternalServerError)
			return
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, file)
		if err != nil {
			writeGeneralResponse(w, "error", "Failed to save file content", nil, http.StatusInternalServerError)
			return
		}

		url := "http://localhost:8080/uploads/" + filename
		photoURL = &url
	}

	updated := models.Todo{
		Title:    title,
		Done:     done,
		PhotoURL: photoURL,
		UserID:   existingTodo.UserID,
	}

	todo, err := h.Store.UpdateTodo(id, updated)
	if err != nil {
		writeGeneralResponse(w, "error", "Failed to update todo", nil, http.StatusInternalServerError)
		return
	}

	writeGeneralResponse(w, "success", "Todo updated", todo, http.StatusOK)
}

// @Summary      Delete a todo by ID
// @Description  Удалить задачу по ID
// @Tags         todos
// @Produce      json
// @Param        id   path      int  true  "Todo ID"
// @Success      204  {object}  models.GeneralResponse "No Content"
// @Failure      404  {object}  models.GeneralResponse
// @Router       /todos/{id} [delete]
func (h *TodoHandler) deleteTodo(w http.ResponseWriter, id int) {
	err := h.Store.DeleteTodo(id)
	if err != nil {
		writeGeneralResponse(w, "error", "Todo not found", nil, http.StatusNotFound)
		return
	}
	writeGeneralResponse(w, "success", "Todo deleted", nil, http.StatusNoContent)
}

// @Summary      Get a todo by ID
// @Description  Получить задачу по ID
// @Tags         todos
// @Produce      json
// @Param        id   path      int  true  "Todo ID"
// @Success      200  {object}  models.GeneralResponse{data=models.Todo}
// @Failure      404  {object}  models.GeneralResponse
// @Router       /todos/{id} [get]
func (h *TodoHandler) getTodoByID(w http.ResponseWriter, _ *http.Request, id int) {
	todo, err := h.Store.GetTodoByID(id)
	if err != nil {
		writeGeneralResponse(w, "error", "Todo not found", nil, http.StatusNotFound)
		return
	}
	writeGeneralResponse(w, "success", "Todo fetched", todo, http.StatusOK)
}
