// db/todo_postgres.go
package db

import (
	"database/sql"
	"todo-api/models"
)

func (s *PostgresStore) GetTodos() ([]models.Todo, error) {
	var todos []models.Todo
	err := s.DB.Select(&todos, "SELECT * FROM todos ORDER BY id")
	return todos, err
}

func (s *PostgresStore) CreateTodo(todo models.Todo) (models.Todo, error) {
	query := `INSERT INTO todos (title, done) VALUES ($1, $2) RETURNING id`
	err := s.DB.QueryRow(query, todo.Title, todo.Done).Scan(&todo.ID)
	return todo, err
}

func (s *PostgresStore) UpdateTodo(id int, updated models.Todo) (models.Todo, error) {
	res, err := s.DB.Exec(
		"UPDATE todos SET title=$1, done=$2 WHERE id=$3",
		updated.Title, updated.Done, id,
	)
	if err != nil {
		return models.Todo{}, err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return models.Todo{}, sql.ErrNoRows
	}
	updated.ID = id
	return updated, nil
}

func (s *PostgresStore) DeleteTodo(id int) error {
	res, err := s.DB.Exec("DELETE FROM todos WHERE id=$1", id)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *PostgresStore) GetTodoByID(id int) (models.Todo, error) {
	var todo models.Todo
	query := "SELECT id, title, done FROM todos WHERE id = $1"
	err := s.DB.QueryRow(query, id).Scan(&todo.ID, &todo.Title, &todo.Done)
	return todo, err
}
