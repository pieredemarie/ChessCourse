package enrollment

import (
	"chesscourse/internal/models"
	"chesscourse/internal/storage"
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setup(t *testing.T) (*enrollmentService, *storage.Storage) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Category{}, &models.Course{}, &models.Enrollment{}); err != nil {
		t.Fatal(err)
	}
	store := storage.NewStorage(db)
	if err := db.Create(&models.Course{ID: 1, Title: "Тактика", Description: "Занятия по тактике", AuthorID: 10, Capacity: 1, Format: models.FormatOnlineGroup, StartDate: time.Now().AddDate(0, 0, 10)}).Error; err != nil {
		t.Fatal(err)
	}
	return NewEnrollmentService(store), store
}
func TestEnrollmentLifecycle(t *testing.T) {
	s, store := setup(t)
	ctx := context.Background()
	now := time.Now()
	s.now = func() time.Time { return now }
	if err := s.CreateEnrollment(ctx, 1, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateEnrollment(ctx, 1, 1); !errors.Is(err, storage.ErrEnrollmentConflict) {
		t.Fatalf("duplicate: %v", err)
	}
	rows, err := s.GetUserEnrollments(ctx, 1)
	if err != nil || len(rows) != 1 {
		t.Fatalf("rows: %v %v", rows, err)
	}
	id := rows[0].ID
	if err := s.DecideEnrollment(ctx, 20, id, "approved", ""); !errors.Is(err, storage.ErrForbidden) {
		t.Fatalf("foreign teacher: %v", err)
	}
	if err := s.DecideEnrollment(ctx, 10, id, "approved", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.DecideEnrollment(ctx, 10, id, "rejected", "wrong group"); !errors.Is(err, storage.ErrEnrollmentConflict) {
		t.Fatalf("repeat decision: %v", err)
	}
	if err := s.CreateEnrollment(ctx, 2, 1); !errors.Is(err, storage.ErrNoSeats) {
		t.Fatalf("last seat: %v", err)
	}
	if err := s.PayForEnrollment(ctx, 1, id, false); !errors.Is(err, ErrPaymentFailed) {
		t.Fatalf("failed payment: %v", err)
	}
	s.now = func() time.Time { return now.Add(73 * time.Hour) }
	if err := s.PayForEnrollment(ctx, 1, id, true); !errors.Is(err, storage.ErrPaymentExpired) {
		t.Fatalf("late payment: %v", err)
	}
	e, err := store.GetEnrollmentByID(id)
	if err != nil || e.Status != models.StatusExpired {
		t.Fatalf("expired: %v %v", e, err)
	}
	if err := s.CreateEnrollment(ctx, 2, 1); err != nil {
		t.Fatalf("seat released: %v", err)
	}
}
func TestCancelAndReject(t *testing.T) {
	s, store := setup(t)
	ctx := context.Background()
	if err := s.CreateEnrollment(ctx, 1, 1); err != nil {
		t.Fatal(err)
	}
	rows, _ := store.GetEnrollmentsByUser(1)
	id := rows[0].ID
	if err := s.DecideEnrollment(ctx, 10, id, "rejected", "bad"); !errors.Is(err, ErrInvalidReason) {
		t.Fatalf("reason: %v", err)
	}
	if err := s.CancelEnrollment(ctx, 1, id); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateEnrollment(ctx, 1, 1); err != nil {
		t.Fatalf("retry: %v", err)
	}
}

func TestSuccessfulTestPayment(t *testing.T) {
	s, store := setup(t)
	ctx := context.Background()
	if err := s.CreateEnrollment(ctx, 1, 1); err != nil {
		t.Fatal(err)
	}
	rows, err := store.GetEnrollmentsByUser(1)
	if err != nil || len(rows) != 1 {
		t.Fatalf("rows: %v %v", rows, err)
	}
	id := rows[0].ID
	if err := s.DecideEnrollment(ctx, 10, id, "approved", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.PayForEnrollment(ctx, 2, id, true); !errors.Is(err, storage.ErrForbidden) {
		t.Fatalf("foreign payment: %v", err)
	}
	if err := s.PayForEnrollment(ctx, 1, id, true); err != nil {
		t.Fatal(err)
	}
	e, err := store.GetEnrollmentByID(id)
	if err != nil || e.Status != models.StatusPaid {
		t.Fatalf("paid: %v %v", e, err)
	}
	if err := s.CreateEnrollment(ctx, 1, 1); !errors.Is(err, storage.ErrNoSeats) && !errors.Is(err, storage.ErrEnrollmentConflict) {
		t.Fatalf("paid repeat: %v", err)
	}
}
