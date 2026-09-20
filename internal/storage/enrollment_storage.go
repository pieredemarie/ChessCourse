package storage

import (
	"chesscourse/internal/models"
)

type EnrollmentStorage interface {
	CreateEnrollment(enrollment *models.Enrollment) error
	GetEnrollmentByID(id int) (*models.Enrollment, error)
	GetEnrollmentByUserAndCourse(userID, courseID int) (*models.Enrollment, error)
	GetEnrollmentsByUser(userID int) ([]models.Enrollment, error)
	GetAllEnrollments(status string) ([]models.Enrollment, error)
	UpdateEnrollment(enrollment *models.Enrollment) error
}

func (s *Storage) CreateEnrollment(enrollment *models.Enrollment) error {
	return s.db.Create(enrollment).Error
}

func (s *Storage) GetEnrollmentByID(id int) (*models.Enrollment, error) {
	var enrollment models.Enrollment
	err := s.db.Preload("User").Preload("Course").First(&enrollment, id).Error
	return &enrollment, err
}

func (s *Storage) GetEnrollmentByUserAndCourse(userID, courseID int) (*models.Enrollment, error) {
	var enrollment models.Enrollment
	err := s.db.Where("user_id = ? AND course_id = ?", userID, courseID).
		Preload("Course").
		First(&enrollment).Error

	if err != nil {
		return nil, err
	}
	return &enrollment, nil
}
func (s *Storage) GetEnrollmentsByUser(userID int) ([]models.Enrollment, error) {
	var enrollments []models.Enrollment
	err := s.db.Where("user_id = ?", userID).
		Preload("Course").
		Order("created_at desc").
		Find(&enrollments).Error
	return enrollments, err
}

func (s *Storage) GetAllEnrollments(status string) ([]models.Enrollment, error) {
	var enrollments []models.Enrollment
	query := s.db.Preload("User").Preload("Course")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at desc").Find(&enrollments).Error
	return enrollments, err
}

func (s *Storage) UpdateEnrollment(enrollment *models.Enrollment) error {
	return s.db.Save(enrollment).Error
}
