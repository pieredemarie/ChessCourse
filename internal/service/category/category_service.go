package category

import (
	"chesscourse/internal/apperror"
	"chesscourse/internal/models"
	"chesscourse/internal/storage"
	"errors"
)

type CategoryService interface {
	GetAllCategories() ([]models.Category, error)
	GetCategoryByID(id int) (*models.Category, error)
	CreateCategory(name, description string) error
	UpdateCategory(id int, name, description string) error
	DeleteCategory(id int) error
}

type categoryService struct {
	Storage storage.CategoryStorage
}

func NewCategoryService(storage storage.CategoryStorage) *categoryService {
	return &categoryService{Storage: storage}
}

func (s *categoryService) GetAllCategories() ([]models.Category, error) {
	return s.Storage.GetAllCategories()
}

func (s *categoryService) GetCategoryByID(id int) (*models.Category, error) {
	category, err := s.Storage.GetCategoryByID(id)
	if err != nil {
		return nil, apperror.ErrCategoryNotFound
	}
	return category, nil
}

func (s *categoryService) CreateCategory(name, description string) error {
	if name == "" {
		return errors.New("name is required")
	}

	existing, _ := s.Storage.GetCategoryByName(name)
	if existing != nil {
		return apperror.ErrCategoryExists
	}

	category := &models.Category{
		Name:        name,
		Description: description,
	}

	return s.Storage.CreateCategory(category)
}

func (s *categoryService) UpdateCategory(id int, name, description string) error {
	category, err := s.Storage.GetCategoryByID(id)
	if err != nil {
		return apperror.ErrCategoryNotFound
	}

	if name != "" {
		category.Name = name
	}
	if description != "" {
		category.Description = description
	}

	return s.Storage.UpdateCategory(category)
}

func (s *categoryService) DeleteCategory(id int) error {
	// Проверяем существование категории
	_, err := s.Storage.GetCategoryByID(id)
	if err != nil {
		return apperror.ErrCategoryNotFound
	}

	// Проверяем, есть ли связанные курсы
	count, err := s.Storage.CountCoursesInCategory(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return apperror.ErrCategoryHasCourses
	}

	return s.Storage.DeleteCategory(id)
}
