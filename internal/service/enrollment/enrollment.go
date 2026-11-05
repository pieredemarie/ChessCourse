package enrollment

import (
	"chesscourse/internal/models"
	"chesscourse/internal/storage"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

var ErrInvalidDecision = errors.New("invalid enrollment decision")
var ErrInvalidReason = errors.New("rejection reason must be 5-200 characters")
var ErrPaymentFailed = errors.New("test payment failed")

type EnrollmentService interface {
	CreateEnrollment(ctx context.Context, userID, courseID int) error
	GetUserEnrollments(ctx context.Context, userID int) ([]models.Enrollment, error)
	GetTeacherEnrollments(ctx context.Context, teacherID int, status string) ([]models.Enrollment, error)
	DecideEnrollment(ctx context.Context, teacherID, id int, status, reason string) error
	CancelEnrollment(ctx context.Context, userID, id int) error
	CheckUserCanAccessCourse(ctx context.Context, userID, courseID int) (bool, error)
	PayForEnrollment(ctx context.Context, userID, id int, success bool) error
	ExpireApproved(ctx context.Context) (int64, error)
	GetEnrollmentByID(ctx context.Context, id int) (*models.Enrollment, error)
}

type enrollmentService struct {
	storage storage.EnrollmentStorage
	now     func() time.Time
}

func NewEnrollmentService(store storage.EnrollmentStorage) *enrollmentService {
	return &enrollmentService{storage: store, now: time.Now}
}

func (s *enrollmentService) CreateEnrollment(ctx context.Context, userID, courseID int) error {
	if courseID <= 0 {
		return storage.ErrCourseUnavailable
	}
	if err := s.storage.CreatePending(ctx, userID, courseID, s.now()); err != nil {
		return fmt.Errorf("create enrollment: %w", err)
	}
	return nil
}
func (s *enrollmentService) GetUserEnrollments(ctx context.Context, userID int) ([]models.Enrollment, error) {
	if _, err := s.ExpireApproved(ctx); err != nil {
		return nil, err
	}
	return s.storage.GetEnrollmentsByUser(userID)
}
func (s *enrollmentService) GetTeacherEnrollments(ctx context.Context, teacherID int, status string) ([]models.Enrollment, error) {
	if _, err := s.ExpireApproved(ctx); err != nil {
		return nil, err
	}
	return s.storage.GetTeacherEnrollments(teacherID, status)
}
func (s *enrollmentService) DecideEnrollment(ctx context.Context, teacherID, id int, status, reason string) error {
	state := models.EnrollmentStatus(status)
	if state != models.StatusApproved && state != models.StatusRejected {
		return ErrInvalidDecision
	}
	if state == models.StatusRejected {
		n := len([]rune(strings.TrimSpace(reason)))
		if n < 5 || n > 200 {
			return ErrInvalidReason
		}
	}
	if err := s.storage.Decide(ctx, teacherID, id, state, reason, s.now()); err != nil {
		return fmt.Errorf("decide enrollment: %w", err)
	}
	return nil
}
func (s *enrollmentService) CancelEnrollment(ctx context.Context, userID, id int) error {
	if err := s.storage.CancelPending(ctx, userID, id); err != nil {
		return fmt.Errorf("cancel enrollment: %w", err)
	}
	return nil
}
func (s *enrollmentService) PayForEnrollment(ctx context.Context, userID, id int, success bool) error {
	if !success {
		e, err := s.storage.GetEnrollmentByID(id)
		if err != nil {
			return err
		}
		if e.UserID != userID {
			return storage.ErrForbidden
		}
		if e.Status != models.StatusApproved {
			return storage.ErrEnrollmentConflict
		}
		if e.PaymentDueAt == nil || !s.now().Before(*e.PaymentDueAt) {
			return s.storage.PayApproved(ctx, userID, id, s.now())
		}
		return ErrPaymentFailed
	}
	if err := s.storage.PayApproved(ctx, userID, id, s.now()); err != nil {
		return fmt.Errorf("pay enrollment: %w", err)
	}
	return nil
}
func (s *enrollmentService) ExpireApproved(ctx context.Context) (int64, error) {
	return s.storage.ExpireApproved(ctx, s.now())
}
func (s *enrollmentService) GetEnrollmentByID(ctx context.Context, id int) (*models.Enrollment, error) {
	return s.storage.GetEnrollmentByID(id)
}
func (s *enrollmentService) CheckUserCanAccessCourse(ctx context.Context, userID, courseID int) (bool, error) {
	e, err := s.storage.GetEnrollmentByUserAndCourse(userID, courseID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return e.Status == models.StatusPaid, nil
}
