package course

import (
	"chesscourse/internal/dto"
	"chesscourse/internal/models"
	"chesscourse/internal/storage"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCourseRules(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Category{}, &models.Course{}, &models.Enrollment{}); err != nil {
		t.Fatal(err)
	}
	store := storage.NewStorage(db)
	if err := db.Create(&models.Category{ID: 1, Name: "Тактика"}).Error; err != nil {
		t.Fatal(err)
	}
	svc := NewCourseService(store)
	base := dto.CreateCourseRequest{Title: "Шахматная тактика", Description: "Групповые занятия шахматами", Level: "beginner", CategoryID: 1, Price: 1500, Format: "online_group", Capacity: 10, StartDate: time.Now().AddDate(0, 0, 2).Format("2006-01-02")}
	tests := []struct {
		name    string
		change  func(*dto.CreateCourseRequest)
		wantErr bool
	}{
		{"valid", func(*dto.CreateCourseRequest) {}, false},
		{"short description", func(r *dto.CreateCourseRequest) { r.Description = "коротко" }, true},
		{"fractional price", func(r *dto.CreateCourseRequest) { r.Price = 100.001 }, true},
		{"zero capacity", func(r *dto.CreateCourseRequest) { r.Capacity = 0 }, true},
		{"in person location", func(r *dto.CreateCourseRequest) { r.Format = "in_person" }, true},
		{"past date", func(r *dto.CreateCourseRequest) { r.StartDate = "2020-01-01" }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := base
			tt.change(&r)
			_, err := svc.CreateCourse(42, r)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, want error %v", err, tt.wantErr)
			}
		})
	}
	courses, err := store.GetTeacherCourses(42)
	if err != nil || len(courses) != 1 {
		t.Fatalf("teacher courses = %d, %v", len(courses), err)
	}
	if _, err := svc.UpdateCourse(99, courses[0].ID, dto.UpdateCourseRequest{}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("foreign update = %v", err)
	}
	if err := svc.ArchiveCourse(99, courses[0].ID); !errors.Is(err, ErrForbidden) {
		t.Fatalf("foreign archive = %v", err)
	}
}
