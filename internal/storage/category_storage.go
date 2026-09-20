package storage

import (
	"chesscourse/internal/models"
)

type CategoryStorage interface {
	GetAllCategories() ([]models.Category, error)
	GetCategoryByID(id int) (*models.Category, error)
	GetCategoryByName(name string) (*models.Category, error)
	CreateCategory(category *models.Category) error
	UpdateCategory(category *models.Category) error
	DeleteCategory(id int) error
	CountCoursesInCategory(categoryID int) (int64, error)
}

func (s *Storage) GetCategoryByName(name string) (*models.Category, error) {
	var category models.Category
	err := s.db.Where("name = ?", name).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (s *Storage) CountCoursesInCategory(categoryID int) (int64, error) {
	var count int64
	err := s.db.Model(&models.Course{}).Where("category_id = ?", categoryID).Count(&count).Error
	return count, err
}

func (s *Storage) GetAllCategories() ([]models.Category, error) {
	var categories []models.Category
	err := s.db.Preload("Courses").Find(&categories).Error
	return categories, err
}

func (s *Storage) GetCategoryByID(id int) (*models.Category, error) {
	var category models.Category
	err := s.db.Preload("Courses").First(&category, id).Error
	return &category, err
}

func (s *Storage) CreateCategory(category *models.Category) error {
	return s.db.Create(category).Error
}

func (s *Storage) UpdateCategory(category *models.Category) error {
	return s.db.Save(category).Error
}

func (s *Storage) DeleteCategory(id int) error {
	return s.db.Delete(&models.Category{}, id).Error
}
