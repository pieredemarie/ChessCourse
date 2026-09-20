package auth

import (
	"chesscourse/internal/apperror"
	"chesscourse/internal/models"
	"chesscourse/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(email, name, password string) error
	Login(email, password string) (string, error)
	GetUserByID(id int) (*models.User, error)
}

type authService struct {
	Storage storage.AuthStorage
}

func NewAuthService(storage storage.AuthStorage) *authService {
	return &authService{Storage: storage}
}

func (s *authService) Register(email, name, password string) error {
	_, err := s.Storage.GetUserByEmail(email)
	if err == nil {
		return apperror.ErrEmailExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		Role:         models.RoleStudent, // по умолчанию — ученик
	}

	return s.Storage.CreateUser(user)
}

func (s *authService) Login(email, password string) (string, error) {
	user, err := s.Storage.GetUserByEmail(email)
	if err != nil {
		return "", apperror.ErrBadCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", apperror.ErrBadCredentials
	}

	token, err := GenerateToken(user.ID, user.Role) // передаём роль
	if err != nil {
		return "", apperror.ErrToken
	}

	return token, nil
}

func (s *authService) GetUserByID(id int) (*models.User, error) {
	return s.Storage.GetUserByID(id)
}
