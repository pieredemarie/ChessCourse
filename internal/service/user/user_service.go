package user

import (
	"chesscourse/internal/apperror"
	"chesscourse/internal/models"
	"chesscourse/internal/storage"
)

type UserService interface {
	GetAllUsers() ([]models.User, error)
	UpdateUserRole(id int, role string) error
}

type userService struct {
	storage storage.UserStorage
}

func NewUserService(storage storage.UserStorage) *userService {
	return &userService{storage: storage}
}

func (s *userService) GetAllUsers() ([]models.User, error) {
	return s.storage.GetAllUsers()
}

func (s *userService) UpdateUserRole(id int, role string) error {
	// Проверяем, существует ли пользователь
	_, err := s.storage.GetUserByID(id)
	if err != nil {
		return apperror.ErrUserNotFound
	}

	// Валидируем роль
	if role != models.RoleStudent && role != models.RoleAdmin {
		return apperror.ErrInvalidRole
	}

	return s.storage.UpdateUserRole(id, role)
}
