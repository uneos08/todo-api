package store

import "todo-api/models"

type TodoStore interface {
	GetTodos(userID int) ([]models.Todo, error)
	CreateTodo(models.Todo) (models.Todo, error)
	UpdateTodo(id int, updated models.Todo) (models.Todo, error)
	DeleteTodo(id int) error
	GetTodoByID(id int) (models.Todo, error)
}
