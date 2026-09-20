package storage

import (
	"chesscourse/internal/models"
)

type UserStorage interface {
	GetAllUsers() ([]models.User, error)
	GetUserByID(id int) (*models.User, error)
	UpdateUserRole(id int, role string) error
}

func (s *Storage) GetAllUsers() ([]models.User, error) {
	var users []models.User
	err := s.db.Order("id asc").Find(&users).Error
	return users, err
}

func (s *Storage) UpdateUserRole(id int, role string) error {
	return s.db.Model(&models.User{}).Where("id = ?", id).Update("role", role).Error
}
