package handlers

import (
	"bytes"
	"chesscourse/internal/service/learning"
	"net/http"
	"net/http/httptest"
	"testing"

	"chesscourse/internal/models"
	"chesscourse/internal/service/enrollment"
	"chesscourse/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestEnrollments(t *testing.T) {
	gin.SetMode(gin.TestMode)

	setupDB := func() *gorm.DB {
		db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		db.AutoMigrate(&models.User{}, &models.Course{}, &models.Enrollment{})
		return db
	}

	t.Run("create enrollment requires auth", func(t *testing.T) {
		db := setupDB()
		store := storage.NewStorage(db)

		// Создаем пользователя и курс
		user := &models.User{Email: "test@test.com", Name: "Test"}
		store.CreateUser(user)
		course := &models.Course{Title: "Test Course"}
		store.CreateCourse(course)

		learningService := learning.NewLearningService(store)
		enrollmentService := enrollment.NewEnrollmentService(store)
		enrollmentHandler := NewEnrollmentHandler(enrollmentService, learningService)

		router := gin.Default()
		router.POST("/api/enrollments", func(c *gin.Context) {
			c.Set("userID", user.ID)
			enrollmentHandler.CreateEnrollment(c)
		})

		body := `{"course_id":1}`
		req := httptest.NewRequest(http.MethodPost, "/api/enrollments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("duplicate enrollment returns error", func(t *testing.T) {
		db := setupDB()
		store := storage.NewStorage(db)

		user := &models.User{Email: "test2@test.com", Name: "Test"}
		store.CreateUser(user)
		course := &models.Course{Title: "Test Course"}
		store.CreateCourse(course)

		// Создаем существующую заявку
		existing := &models.Enrollment{UserID: user.ID, CourseID: course.ID, Status: models.StatusPending}
		store.CreateEnrollment(existing)

		learningService := learning.NewLearningService(store)
		enrollmentService := enrollment.NewEnrollmentService(store)
		enrollmentHandler := NewEnrollmentHandler(enrollmentService, learningService)

		router := gin.Default()
		router.POST("/api/enrollments", func(c *gin.Context) {
			c.Set("userID", user.ID)
			enrollmentHandler.CreateEnrollment(c)
		})

		body := `{"course_id":1}`
		req := httptest.NewRequest(http.MethodPost, "/api/enrollments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
