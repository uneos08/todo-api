package store

import "todo-api/models"

type UserStore interface {
	GetUsers() ([]models.User, error)
	CreateUser(models.User) (models.User, error)
	UpdateUser(id int, updated models.User) (models.User, error)
	DeleteUser(id int) error
	GetUserByID(id int) (models.User, error)
	GetByUsername(username string) (models.User, error)
}
