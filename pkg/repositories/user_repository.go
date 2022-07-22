package repositories

import (
	"errors"

	"github.com/johannaojeling/go-rest-api/pkg/models"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	CreateUser(user *models.User) error
	GetAllUsers() ([]*models.User, error)
	GetUserById(id string) (*models.User, error)
	UpdateUserById(id string, updates *models.User) (*models.User, error)
	DeleteUserById(id string) error
}
