// db/todo_postgres.go
package db

import (
	"database/sql"
	"todo-api/models"
)

func (s *PostgresStore) GetTodos(userID int) ([]models.Todo, error) {
	var todos []models.Todo
	query := `SELECT id, title, done, user_id, COALESCE(photo_url, '') as photo_url
          FROM todos WHERE user_id = $1 ORDER BY id`
	err := s.DB.Select(&todos, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return []models.Todo{}, nil
		}
		return nil, err
	}
	return todos, nil
}

func (s *PostgresStore) CreateTodo(todo models.Todo) (models.Todo, error) {
	query := `INSERT INTO todos (title, done, user_id, photo_url) VALUES ($1, $2, $3, $4) RETURNING id`
	err := s.DB.QueryRow(query, todo.Title, todo.Done, todo.UserID, todo.PhotoURL).Scan(&todo.ID)
	return todo, err
}

func (s *PostgresStore) UpdateTodo(id int, updated models.Todo) (models.Todo, error) {
	res, err := s.DB.Exec(
		"UPDATE todos SET title=$1, done=$2, photo_url=$3 WHERE id=$4",
		updated.Title, updated.Done, updated.PhotoURL, id,
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
	query := "SELECT id, title, done, user_id, photo_url FROM todos WHERE id = $1"
	err := s.DB.QueryRow(query, id).Scan(&todo.ID, &todo.Title, &todo.Done, &todo.UserID, &todo.PhotoURL)
	return todo, err
}
