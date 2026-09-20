package storage

import (
	"chesscourse/internal/models"
)

type CourseStorage interface {
	GetAllCourses(search, level, categoryID string) ([]models.Course, int64, error)
	GetCourseByID(id int) (*models.Course, error)
	CreateCourse(course *models.Course) error
	UpdateCourse(course *models.Course) error
	DeleteCourse(id int) error
}

func (s *Storage) GetAllCourses(search, level, categoryID string) ([]models.Course, int64, error) {
	var courses []models.Course
	var total int64

	query := s.db.Model(&models.Course{}).Preload("Category")

	// Поиск по названию
	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}

	// Фильтр по уровню
	if level != "" {
		query = query.Where("level = ?", level)
	}

	// Фильтр по категории
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	// Считаем общее количество
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Получаем курсы
	err := query.Order("id desc").Find(&courses).Error

	return courses, total, err
}

func (s *Storage) GetCourseByID(id int) (*models.Course, error) {
	var course models.Course
	err := s.db.Preload("Category").First(&course, id).Error
	return &course, err
}

func (s *Storage) CreateCourse(course *models.Course) error {
	return s.db.Create(course).Error
}

func (s *Storage) UpdateCourse(course *models.Course) error {
	return s.db.Save(course).Error
}

func (s *Storage) DeleteCourse(id int) error {
	return s.db.Delete(&models.Course{}, id).Error
}
