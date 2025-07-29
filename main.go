// @title           ToDo API
// @version         1.0
// @description     Simple ToDo API with PostgreSQL and Go.

// @host      localhost:8080
// @BasePath  /api
package main

import (
	"fmt"
	"log"
	"net/http"

	"todo-api/db"
	"todo-api/handlers"

	_ "todo-api/docs"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	connStr := "host=localhost port=5432 user=postgres password=psw dbname=todo_db sslmode=disable"
	store := db.NewPostgresStore(connStr)

	// Разделяем хранилища
	todoHandler := handlers.NewTodoHandler(store)
	userHandler := handlers.NewUserHandler(store)

	// Роутер
	r := mux.NewRouter()

	// Регистрируем маршруты
	todoHandler.RegisterRoutes(r)
	userHandler.RegisterRoutes(r)

	// Swagger
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	fmt.Println("🚀 Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
