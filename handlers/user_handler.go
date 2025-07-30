package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"todo-api/auth"
	"todo-api/models"
	"todo-api/store"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	Store store.UserStore
}

func NewUserHandler(store store.UserStore) *UserHandler {
	return &UserHandler{Store: store}
}

func (h *UserHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/register", h.Register).Methods("POST")
	r.HandleFunc("/api/login", h.Login).Methods("POST")
	r.HandleFunc("/api/refresh", h.RefreshToken).Methods("POST")
	r.HandleFunc("/api/users", h.GetAllUsers).Methods("GET")
	r.HandleFunc("/api/users/{id}", h.GetUserByID).Methods("GET")
	r.HandleFunc("/api/users/{id}", h.UpdateUser).Methods("PUT")
	r.HandleFunc("/api/users/{id}", h.DeleteUser).Methods("DELETE")
}

type userCredentials struct {
	Username string `json:"username" example:"user1"`
	Password string `json:"password" example:"password123"`
}

// Register godoc
// @Summary      Register a new user
// @Description  Register a new user with username and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      userCredentials  true  "User registration info"
// @Success      201   {object}  models.GeneralResponse{data=models.User}
// @Failure      400   {object}  models.GeneralResponse
// @Failure      500   {object}  models.GeneralResponse
// @Router       /register [post]
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input userCredentials
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeGeneralResponse(w, "error", "Invalid JSON", nil, http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		writeGeneralResponse(w, "error", "Error hashing password", nil, http.StatusInternalServerError)
		return
	}

	user := models.User{
		Username:     input.Username,
		PasswordHash: string(hashedPassword),
	}

	createdUser, err := h.Store.CreateUser(user)
	if err != nil {
		writeGeneralResponse(w, "error", "Failed to create user", nil, http.StatusInternalServerError)
		return
	}

	createdUser.PasswordHash = "" // не показываем хэш
	writeGeneralResponse(w, "success", "User registered", createdUser, http.StatusCreated)
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user and return JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      userCredentials  true  "Login credentials"
// @Success      200   {object}  models.GeneralResponse{data=object{token=string}}
// @Failure      400   {object}  models.GeneralResponse
// @Failure      401   {object}  models.GeneralResponse
// @Router       /login [post]
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds userCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		writeGeneralResponse(w, "error", "Invalid JSON", nil, http.StatusBadRequest)
		return
	}

	user, err := h.Store.GetByUsername(creds.Username)
	if err != nil {
		writeGeneralResponse(w, "error", "User not found", nil, http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)); err != nil {
		writeGeneralResponse(w, "error", "Invalid password", nil, http.StatusUnauthorized)
		return
	}

	token, err := auth.CreateJWTToken(user.ID)
	refreshToken, err1 := auth.CreateRefreshToken(user.ID)
	if err != nil || err1 != nil {
		writeGeneralResponse(w, "error", "Failed to create refresh token", nil, http.StatusInternalServerError)

		return
	}

	writeGeneralResponse(w, "success", "Login successful", map[string]string{
		"access_token":  token,
		"refresh_token": refreshToken,
	}, http.StatusOK)

}

// GetAllUsers godoc
// @Summary      Get all users
// @Description  Retrieve list of all users (passwords omitted)
// @Tags         users
// @Produce      json
// @Success      200  {object}  models.GeneralResponse{data=[]models.User}
// @Failure      500  {object}  models.GeneralResponse
// @Router       /users [get]
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.Store.GetUsers()
	if err != nil {
		writeGeneralResponse(w, "error", "Failed to fetch users", nil, http.StatusInternalServerError)
		return
	}
	for i := range users {
		users[i].PasswordHash = ""
	}
	writeGeneralResponse(w, "success", "Users fetched", users, http.StatusOK)
}

// GetUserByID godoc
// @Summary      Get user by ID
// @Description  Get user details by user ID (password omitted)
// @Tags         users
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  models.GeneralResponse{data=models.User}
// @Failure      400  {object}  models.GeneralResponse
// @Failure      404  {object}  models.GeneralResponse
// @Router       /users/{id} [get]
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeGeneralResponse(w, "error", "Invalid ID", nil, http.StatusBadRequest)
		return
	}

	user, err := h.Store.GetUserByID(id)
	if err != nil {
		writeGeneralResponse(w, "error", "User not found", nil, http.StatusNotFound)
		return
	}
	user.PasswordHash = ""
	writeGeneralResponse(w, "success", "User found", user, http.StatusOK)
}

type updateUserInput struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
}

// UpdateUser godoc
// @Summary      Update user by ID
// @Description  Update user data (password will be hashed if provided)
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      int          true  "User ID"
// @Param        user  body      updateUserInput  true  "Updated user info"
// @Success      200   {object}  models.GeneralResponse{data=models.User}
// @Failure      400   {object}  models.GeneralResponse
// @Failure      500   {object}  models.GeneralResponse
// @Router       /users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeGeneralResponse(w, "error", "Invalid ID", nil, http.StatusBadRequest)
		return
	}

	var input updateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeGeneralResponse(w, "error", "Invalid JSON", nil, http.StatusBadRequest)
		return
	}

	user := models.User{
		Username: input.Username,
	}
	if input.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			writeGeneralResponse(w, "error", "Error hashing password", nil, http.StatusInternalServerError)
			return
		}
		user.PasswordHash = string(hashedPassword)
	}

	updatedUser, err := h.Store.UpdateUser(id, user)
	if err != nil {
		writeGeneralResponse(w, "error", "Failed to update user", nil, http.StatusInternalServerError)
		return
	}

	updatedUser.PasswordHash = ""
	writeGeneralResponse(w, "success", "User updated", updatedUser, http.StatusOK)
}

// DeleteUser godoc
// @Summary      Delete user by ID
// @Description  Delete user by given ID
// @Tags         users
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  models.GeneralResponse
// @Failure      400  {object}  models.GeneralResponse
// @Failure      500  {object}  models.GeneralResponse
// @Router       /users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeGeneralResponse(w, "error", "Invalid ID", nil, http.StatusBadRequest)
		return
	}

	err = h.Store.DeleteUser(id)
	if err != nil {
		writeGeneralResponse(w, "error", "Failed to delete user", nil, http.StatusInternalServerError)
		return
	}

	writeGeneralResponse(w, "success", "User deleted", nil, http.StatusOK)
}

// RefreshToken godoc
// @Summary      Refresh JWT tokens
// @Description  Use refresh token to get new access and refresh tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        token  body      object{refresh_token=string}  true  "Refresh token"
// @Success      200    {object}  models.GeneralResponse{data=object{access_token=string, refresh_token=string}}
// @Failure      400    {object}  models.GeneralResponse
// @Failure      401    {object}  models.GeneralResponse
// @Router       /refresh [post]
func (h *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		writeGeneralResponse(w, "error", "Invalid JSON or missing refresh_token", nil, http.StatusBadRequest)
		return
	}

	claims, err := auth.ParseJWTToken(req.RefreshToken)
	if err != nil {
		writeGeneralResponse(w, "error", "Invalid refresh token", nil, http.StatusUnauthorized)
		return
	}

	// Здесь можно проверить, что пользователь существует
	user, err := h.Store.GetUserByID(claims.UserID)
	if err != nil {
		writeGeneralResponse(w, "error", "User not found", nil, http.StatusUnauthorized)
		return
	}

	// Создаем новые токены
	accessToken, err := auth.CreateJWTToken(user.ID)
	if err != nil {
		writeGeneralResponse(w, "error", "Failed to create access token", nil, http.StatusInternalServerError)
		return
	}

	refreshToken, err := auth.CreateRefreshToken(user.ID)
	if err != nil {
		writeGeneralResponse(w, "error", "Failed to create refresh token", nil, http.StatusInternalServerError)
		return
	}

	writeGeneralResponse(w, "success", "Tokens refreshed", map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, http.StatusOK)
}
