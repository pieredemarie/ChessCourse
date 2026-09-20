package enrollment

import (
	"chesscourse/internal/apperror"
	"chesscourse/internal/models"
	"chesscourse/internal/storage"
	"errors"

	"gorm.io/gorm"
)

type EnrollmentService interface {
	CreateEnrollment(userID, courseID int) error
	GetUserEnrollments(userID int) ([]models.Enrollment, error)
	GetAllEnrollments(status string) ([]models.Enrollment, error)
	UpdateEnrollmentStatus(id int, status string, adminNote string) error
	CheckUserCanAccessCourse(userID, courseID int) (bool, error)
	PayForEnrollment(enrollmentID int) error
	GetEnrollmentByID(id int) (*models.Enrollment, error)
	GetUserCourses(userID int) ([]models.Enrollment, error)
}

type enrollmentService struct {
	storage storage.EnrollmentStorage
}

func NewEnrollmentService(storage storage.EnrollmentStorage) *enrollmentService {
	return &enrollmentService{storage: storage}
}

func (s *enrollmentService) CreateEnrollment(userID, courseID int) error {
	existing, err := s.storage.GetEnrollmentByUserAndCourse(userID, courseID)

	if err == nil {
		if existing.Status == models.StatusRejected {
			return s.storage.CreateEnrollment(&models.Enrollment{
				UserID:   userID,
				CourseID: courseID,
				Status:   models.StatusPending,
			})
		}
		return errors.New("enrollment already exists")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return s.storage.CreateEnrollment(&models.Enrollment{
		UserID:   userID,
		CourseID: courseID,
		Status:   models.StatusPending,
	})
}

func (s *enrollmentService) GetUserEnrollments(userID int) ([]models.Enrollment, error) {
	return s.storage.GetEnrollmentsByUser(userID)
}

func (s *enrollmentService) GetAllEnrollments(status string) ([]models.Enrollment, error) {
	return s.storage.GetAllEnrollments(status)
}

func (s *enrollmentService) UpdateEnrollmentStatus(id int, status string, adminNote string) error {
	enrollment, err := s.storage.GetEnrollmentByID(id)
	if err != nil {
		return apperror.ErrUserNotFound
	}

	enrollment.Status = models.EnrollmentStatus(status)
	enrollment.AdminNote = adminNote

	return s.storage.UpdateEnrollment(enrollment)
}

func (s *enrollmentService) CheckUserCanAccessCourse(userID, courseID int) (bool, error) {
	enrollment, err := s.storage.GetEnrollmentByUserAndCourse(userID, courseID)
	if err != nil {
		return false, nil
	}
	return enrollment.Status == models.StatusPaid, nil
}

func (s *enrollmentService) PayForEnrollment(enrollmentID int) error {
	enrollment, err := s.storage.GetEnrollmentByID(enrollmentID)
	if err != nil {
		return apperror.ErrUserNotFound
	}

	if enrollment.Status != models.StatusApproved {
		return errors.New("enrollment is not approved")
	}

	enrollment.Status = models.StatusPaid
	return s.storage.UpdateEnrollment(enrollment)
}

func (s *enrollmentService) GetEnrollmentByID(id int) (*models.Enrollment, error) {
	return s.storage.GetEnrollmentByID(id)
}
func (s *enrollmentService) GetUserCourses(userID int) ([]models.Enrollment, error) {
	return s.storage.GetEnrollmentsByUser(userID)
}
