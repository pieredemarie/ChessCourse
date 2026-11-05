package storage

import (
	"chesscourse/internal/models"
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserStorage interface {
	GetAllUsers() ([]models.User, error)
	GetUserByID(id int) (*models.User, error)
	UpdateUserRole(id int, role string) error
	UpdateUser(ctx context.Context, id int, name, lastName, role *string) error
	DeleteUser(ctx context.Context, id int) error
}

func (s *Storage) GetAllUsers() ([]models.User, error) {
	var users []models.User
	err := s.db.Order("id asc").Find(&users).Error
	return users, err
}

func (s *Storage) UpdateUserRole(id int, role string) error {
	r := s.db.Model(&models.User{}).Where("id = ?", id).Update("role", role)
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *Storage) UpdateUser(ctx context.Context, id int, name, lastName, role *string) error {
	updates := map[string]any{}
	if name != nil {
		updates["name"] = *name
	}
	if lastName != nil {
		updates["last_name"] = *lastName
	}
	if role != nil {
		updates["role"] = *role
	}
	if len(updates) == 0 {
		return nil
	}
	r := s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Updates(updates)
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *Storage) DeleteUser(ctx context.Context, id int) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var u models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&u, id).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Enrollment{}).Where("user_id = ? AND status IN ?", id, []models.EnrollmentStatus{models.StatusPending, models.StatusApproved, models.StatusPaid}).Update("status", models.StatusCancelled).Error; err != nil {
			return err
		}
		return tx.Delete(&u).Error
	})
}
