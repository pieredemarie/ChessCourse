package user

import (
	"chesscourse/internal/dto"
	"chesscourse/internal/models"
	"chesscourse/internal/storage"
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"
)

var ErrSelfDelete = errors.New("administrator cannot delete self")
var ErrInvalidProfile = errors.New("invalid user profile")

type UserService interface {
	GetAllUsers() ([]models.User, error)
	UpdateUserRole(id int, role string) error
	UpdateUser(ctx context.Context, id int, req dto.UpdateUserRequest) error
	DeleteUser(ctx context.Context, adminID, id int) error
}
type userService struct{ storage storage.UserStorage }

func NewUserService(s storage.UserStorage) *userService    { return &userService{storage: s} }
func (s *userService) GetAllUsers() ([]models.User, error) { return s.storage.GetAllUsers() }
func validRole(role string) bool {
	return role == models.RoleStudent || role == models.RoleTeacher || role == models.RoleAdmin
}
func validName(name string) bool {
	n := []rune(strings.TrimSpace(name))
	if len(n) < 2 || len(n) > 30 {
		return false
	}
	for _, r := range n {
		if !unicode.IsLetter(r) && r != '-' && r != ' ' {
			return false
		}
	}
	return true
}
func (s *userService) UpdateUserRole(id int, role string) error {
	if !validRole(role) {
		return ErrInvalidProfile
	}
	return s.storage.UpdateUserRole(id, role)
}
func (s *userService) UpdateUser(ctx context.Context, id int, req dto.UpdateUserRequest) error {
	if req.Name != nil && !validName(*req.Name) {
		return ErrInvalidProfile
	}
	if req.LastName != nil && !validName(*req.LastName) {
		return ErrInvalidProfile
	}
	if req.Role != nil && !validRole(*req.Role) {
		return ErrInvalidProfile
	}
	if err := s.storage.UpdateUser(ctx, id, req.Name, req.LastName, req.Role); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}
func (s *userService) DeleteUser(ctx context.Context, adminID, id int) error {
	if adminID == id {
		return ErrSelfDelete
	}
	if err := s.storage.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
