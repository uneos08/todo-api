package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"todo-api/models"
	"todo-api/store"

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
		h.getTodos(w)
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
func (h *TodoHandler) getTodos(w http.ResponseWriter) {
	todos, err := h.Store.GetTodos()
	if err != nil {
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
	var todo models.Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		writeGeneralResponse(w, "error", "Invalid JSON", nil, http.StatusBadRequest)
		log.Printf("❌ Failed to create todo: %v", err)
		return
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
	var updated models.Todo
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		writeGeneralResponse(w, "error", "Invalid JSON", nil, http.StatusBadRequest)
		return
	}
	todo, err := h.Store.UpdateTodo(id, updated)
	if err != nil {
		writeGeneralResponse(w, "error", "Todo not found", nil, http.StatusNotFound)
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
