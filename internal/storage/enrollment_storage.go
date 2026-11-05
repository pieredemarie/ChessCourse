package storage

import (
	"chesscourse/internal/models"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrEnrollmentConflict = errors.New("enrollment state conflict")
var ErrNoSeats = errors.New("нет свободных мест")
var ErrForbidden = errors.New("access denied")
var ErrCourseUnavailable = errors.New("course is unavailable")
var ErrPaymentExpired = errors.New("payment period expired")

type EnrollmentStorage interface {
	CreatePending(ctx context.Context, userID, courseID int, now time.Time) error
	Decide(ctx context.Context, teacherID, id int, status models.EnrollmentStatus, reason string, now time.Time) error
	CancelPending(ctx context.Context, userID, id int) error
	PayApproved(ctx context.Context, userID, id int, now time.Time) error
	ExpireApproved(ctx context.Context, now time.Time) (int64, error)
	GetEnrollmentByID(id int) (*models.Enrollment, error)
	GetEnrollmentByUserAndCourse(userID, courseID int) (*models.Enrollment, error)
	GetEnrollmentsByUser(userID int) ([]models.Enrollment, error)
	GetTeacherEnrollments(teacherID int, status string) ([]models.Enrollment, error)
}

func (s *Storage) CreatePending(ctx context.Context, userID, courseID int, now time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var c models.Course
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&c, courseID).Error; err != nil {
			return fmt.Errorf("load course: %w", err)
		}
		if c.Archived || c.StartDate.Format("2006-01-02") < now.Format("2006-01-02") {
			return ErrCourseUnavailable
		}
		var reserved int64
		if err := tx.Model(&models.Enrollment{}).Where("course_id = ? AND status IN ?", courseID, []models.EnrollmentStatus{models.StatusApproved, models.StatusPaid}).Count(&reserved).Error; err != nil {
			return err
		}
		if reserved >= int64(c.Capacity) {
			return ErrNoSeats
		}
		var existing int64
		if err := tx.Model(&models.Enrollment{}).Where("user_id = ? AND course_id = ? AND status IN ?", userID, courseID, []models.EnrollmentStatus{models.StatusPending, models.StatusApproved, models.StatusPaid}).Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return ErrEnrollmentConflict
		}
		return tx.Create(&models.Enrollment{UserID: userID, CourseID: courseID, Status: models.StatusPending}).Error
	})
}

func (s *Storage) Decide(ctx context.Context, teacherID, id int, status models.EnrollmentStatus, reason string, now time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var e models.Enrollment
		if err := tx.First(&e, id).Error; err != nil {
			return err
		}
		var c models.Course
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&c, e.CourseID).Error; err != nil {
			return err
		}
		if c.AuthorID != teacherID {
			return ErrForbidden
		}
		if e.Status != models.StatusPending {
			return ErrEnrollmentConflict
		}
		if status == models.StatusApproved {
			var reserved int64
			if err := tx.Model(&models.Enrollment{}).Where("course_id = ? AND status IN ?", c.ID, []models.EnrollmentStatus{models.StatusApproved, models.StatusPaid}).Count(&reserved).Error; err != nil {
				return err
			}
			if reserved >= int64(c.Capacity) {
				return ErrNoSeats
			}
			due := now.Add(72 * time.Hour)
			e.PaymentDueAt = &due
		} else {
			e.RejectionReason = strings.TrimSpace(reason)
			e.AdminNote = e.RejectionReason
		}
		e.Status = status
		return tx.Save(&e).Error
	})
}

func (s *Storage) CancelPending(ctx context.Context, userID, id int) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var e models.Enrollment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&e, id).Error; err != nil {
			return err
		}
		if e.UserID != userID {
			return ErrForbidden
		}
		if e.Status != models.StatusPending {
			return ErrEnrollmentConflict
		}
		return tx.Model(&e).Update("status", models.StatusCancelled).Error
	})
}

func (s *Storage) PayApproved(ctx context.Context, userID, id int, now time.Time) error {
	expired := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var e models.Enrollment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&e, id).Error; err != nil {
			return err
		}
		if e.UserID != userID {
			return ErrForbidden
		}
		if e.Status != models.StatusApproved {
			return ErrEnrollmentConflict
		}
		if e.PaymentDueAt == nil || !now.Before(*e.PaymentDueAt) {
			if err := tx.Model(&e).Update("status", models.StatusExpired).Error; err != nil {
				return err
			}
			expired = true
			return nil
		}
		return tx.Model(&e).Update("status", models.StatusPaid).Error
	})
	if err != nil {
		return err
	}
	if expired {
		return ErrPaymentExpired
	}
	return nil
}

func (s *Storage) ExpireApproved(ctx context.Context, now time.Time) (int64, error) {
	r := s.db.WithContext(ctx).Model(&models.Enrollment{}).Where("status = ? AND payment_due_at <= ?", models.StatusApproved, now).Update("status", models.StatusExpired)
	return r.RowsAffected, r.Error
}

func (s *Storage) GetEnrollmentByID(id int) (*models.Enrollment, error) {
	var e models.Enrollment
	err := s.db.Preload("User").Preload("Course", func(tx *gorm.DB) *gorm.DB { return tx.Unscoped() }).First(&e, id).Error
	return &e, err
}
func (s *Storage) GetEnrollmentByUserAndCourse(userID, courseID int) (*models.Enrollment, error) {
	var e models.Enrollment
	err := s.db.Where("user_id = ? AND course_id = ?", userID, courseID).Preload("Course", func(tx *gorm.DB) *gorm.DB { return tx.Unscoped() }).Order("id desc").First(&e).Error
	return &e, err
}
func (s *Storage) GetEnrollmentsByUser(userID int) ([]models.Enrollment, error) {
	var rows []models.Enrollment
	err := s.db.Where("user_id = ?", userID).Preload("Course", func(tx *gorm.DB) *gorm.DB { return tx.Unscoped() }).Order("created_at desc").Find(&rows).Error
	return rows, err
}
func (s *Storage) GetTeacherEnrollments(teacherID int, status string) ([]models.Enrollment, error) {
	var rows []models.Enrollment
	q := s.db.Model(&models.Enrollment{}).Joins("JOIN courses ON courses.id = enrollments.course_id").Where("courses.author_id = ?", teacherID).Preload("User").Preload("Course", func(tx *gorm.DB) *gorm.DB { return tx.Unscoped() })
	if status != "" {
		q = q.Where("enrollments.status = ?", status)
	}
	err := q.Order("enrollments.created_at desc").Find(&rows).Error
	return rows, err
}
