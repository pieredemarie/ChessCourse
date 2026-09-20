package course

import (
	"chesscourse/internal/dto"
	"chesscourse/internal/models"
	"chesscourse/internal/storage"
)

type CourseService interface {
	GetAllCourses(search, level, categoryID string) ([]models.Course, int64, error)
	GetCourseByID(id int) (*models.Course, error)
	CreateCourse(req dto.CreateCourseRequest) (*models.Course, error)
	UpdateCourse(id int, req dto.UpdateCourseRequest) (*models.Course, error)
	DeleteCourse(id int) error
}

type courseService struct {
	storage storage.CourseStorage
}

func NewCourseService(storage storage.CourseStorage) *courseService {
	return &courseService{storage: storage}
}

func (s *courseService) GetAllCourses(search, level, categoryID string) ([]models.Course, int64, error) {
	return s.storage.GetAllCourses(search, level, categoryID)
}

func (s *courseService) GetCourseByID(id int) (*models.Course, error) {
	return s.storage.GetCourseByID(id)
}

func (s *courseService) CreateCourse(req dto.CreateCourseRequest) (*models.Course, error) {
	course := &models.Course{
		Title:       req.Title,
		Description: req.Description,
		Level:       models.CourseLevel(req.Level),
		CategoryID:  req.CategoryID,
		Price:       req.Price,
		CoverURL:    req.CoverURL,
		Duration:    req.Duration,
		Lessons:     req.Lessons,
	}

	if err := s.storage.CreateCourse(course); err != nil {
		return nil, err
	}

	return s.storage.GetCourseByID(course.ID)
}

func (s *courseService) UpdateCourse(id int, req dto.UpdateCourseRequest) (*models.Course, error) {
	course, err := s.storage.GetCourseByID(id)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		course.Title = req.Title
	}
	if req.Description != "" {
		course.Description = req.Description
	}
	if req.Level != "" {
		course.Level = models.CourseLevel(req.Level)
	}
	if req.CategoryID != 0 {
		course.CategoryID = req.CategoryID
	}
	if req.Price != 0 {
		course.Price = req.Price
	}
	if req.CoverURL != "" {
		course.CoverURL = req.CoverURL
	}
	if req.Duration != 0 {
		course.Duration = req.Duration
	}
	if req.Lessons != 0 {
		course.Lessons = req.Lessons
	}

	if err := s.storage.UpdateCourse(course); err != nil {
		return nil, err
	}

	return s.storage.GetCourseByID(id)
}

func (s *courseService) DeleteCourse(id int) error {
	return s.storage.DeleteCourse(id)
}
