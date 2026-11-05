package user

import (
	"chesscourse/internal/dto"
	"chesscourse/internal/models"
	"chesscourse/internal/storage"
	"context"
	"errors"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAdminUserManagement(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Category{}, &models.Course{}, &models.Enrollment{}); err != nil {
		t.Fatal(err)
	}
	store := storage.NewStorage(db)
	svc := NewUserService(store)
	ctx := context.Background()
	for _, u := range []models.User{{ID: 1, Name: "Admin", Role: models.RoleAdmin}, {ID: 2, Name: "Иван", Role: models.RoleStudent}} {
		if err := db.Create(&u).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Create(&models.Course{ID: 1, Title: "Тактика", Capacity: 1}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.Enrollment{UserID: 2, CourseID: 1, Status: models.StatusApproved}).Error; err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteUser(ctx, 1, 1); !errors.Is(err, ErrSelfDelete) {
		t.Fatalf("self delete: %v", err)
	}
	name := "Пётр"
	last := "Иванов"
	role := models.RoleTeacher
	if err := svc.UpdateUser(ctx, 2, dto.UpdateUserRequest{Name: &name, LastName: &last, Role: &role}); err != nil {
		t.Fatal(err)
	}
	u, err := store.GetUserByID(2)
	if err != nil || u.Role != role || u.LastName != last {
		t.Fatalf("updated user: %v %v", u, err)
	}
	if err := svc.DeleteUser(ctx, 1, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetUserByID(2); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("deleted user: %v", err)
	}
	seats, err := store.AvailableSeats(1)
	if err != nil || seats != 1 {
		t.Fatalf("released seats: %d %v", seats, err)
	}
}
