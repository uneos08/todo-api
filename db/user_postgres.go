package db

import (
	"database/sql"
	"todo-api/models"
)

func (s *PostgresStore) GetUsers() ([]models.User, error) {
	var users []models.User
	err := s.DB.Select(&users, "SELECT id, username, password_hash FROM users ORDER BY id")
	return users, err
}

func (s *PostgresStore) CreateUser(user models.User) (models.User, error) {
	query := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id`
	err := s.DB.QueryRow(query, user.Username, user.PasswordHash).Scan(&user.ID)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (s *PostgresStore) UpdateUser(id int, updated models.User) (models.User, error) {
	query := `UPDATE users SET username=$1, password_hash=$2 WHERE id=$3`
	res, err := s.DB.Exec(query, updated.Username, updated.PasswordHash, id)
	if err != nil {
		return models.User{}, err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return models.User{}, sql.ErrNoRows
	}
	updated.ID = id
	return updated, nil
}

func (s *PostgresStore) DeleteUser(id int) error {
	res, err := s.DB.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *PostgresStore) GetUserByID(id int) (models.User, error) {
	var user models.User
	query := `SELECT id, username, password_hash FROM users WHERE id = $1`
	err := s.DB.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (s *PostgresStore) GetByUsername(username string) (models.User, error) {
	var user models.User
	query := `SELECT id, username, password_hash FROM users WHERE username = $1`
	err := s.DB.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}
