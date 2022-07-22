package repositories

import (
	"errors"

	"gorm.io/gorm"

	"github.com/johannaojeling/go-rest-api/pkg/models"
)

type UserSQLRepository struct {
	gormDB *gorm.DB
}

func NewSQLUserRepository(DB *gorm.DB) *UserSQLRepository {
	return &UserSQLRepository{gormDB: DB}
}

func (repo *UserSQLRepository) CreateUser(user *models.User) error {
	return repo.gormDB.Create(user).Error
}

func (repo *UserSQLRepository) GetUserById(id string) (*models.User, error) {
	user := &models.User{}
	err := repo.gormDB.Where("id = ?", id).First(user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *UserSQLRepository) GetAllUsers() ([]*models.User, error) {
	var users []*models.User
	err := repo.gormDB.Find(&users).Error
	return users, err
}

func (repo *UserSQLRepository) UpdateUserById(
	id string,
	updates *models.User,
) (*models.User, error) {
	user := &models.User{}
	err := repo.gormDB.Where("id = ?", id).First(user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	err = repo.gormDB.Model(user).Updates(updates).Error
	return user, err
}

func (repo *UserSQLRepository) DeleteUserById(id string) error {
	user := &models.User{}
	err := repo.gormDB.Where("id = ?", id).First(user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrUserNotFound
	}
	if err != nil {
		return err
	}
	return repo.gormDB.Delete(&user).Error
}
