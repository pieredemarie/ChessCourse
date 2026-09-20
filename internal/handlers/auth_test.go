package handlers

import (
	"bytes"
	authservice "chesscourse/internal/service/auth"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"chesscourse/internal/models"
	"chesscourse/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&models.User{}, &models.Category{}, &models.Course{}, &models.Enrollment{}, &models.Module{}, &models.Lesson{}, &models.UserProgress{})
	return db
}

func TestAuthRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful registration", func(t *testing.T) {
		db := setupTestDB()
		store := storage.NewStorage(db)
		authService := authservice.NewAuthService(store)
		authHandler := NewAuthHandler(authService)

		router := gin.Default()
		router.POST("/api/auth/register", authHandler.Register)

		body := `{"email":"test@example.com","name":"Test User","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("duplicate email returns 409", func(t *testing.T) {
		db := setupTestDB()
		store := storage.NewStorage(db)

		// Создаем пользователя
		user := &models.User{Email: "duplicate@test.com", Name: "Test", Role: models.RoleStudent}
		store.CreateUser(user)

		authService := authservice.NewAuthService(store)
		authHandler := NewAuthHandler(authService)

		router := gin.Default()
		router.POST("/api/auth/register", authHandler.Register)

		body := `{"email":"duplicate@test.com","name":"Test User","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("invalid request body returns 400", func(t *testing.T) {
		db := setupTestDB()
		store := storage.NewStorage(db)
		authService := authservice.NewAuthService(store)
		authHandler := NewAuthHandler(authService)

		router := gin.Default()
		router.POST("/api/auth/register", authHandler.Register)

		body := `{"invalid":"json"`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful login", func(t *testing.T) {
		db := setupTestDB()
		store := storage.NewStorage(db)

		// Создаем пользователя с правильным хешированным паролем
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		user := &models.User{
			Email:        "login@test.com",
			Name:         "Test",
			Role:         models.RoleStudent,
			PasswordHash: string(hashedPassword),
		}
		store.CreateUser(user)

		authService := authservice.NewAuthService(store)
		authHandler := NewAuthHandler(authService)

		router := gin.Default()
		router.POST("/api/auth/login", authHandler.Login)

		body := `{"email":"login@test.com","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "token")
	})

	t.Run("invalid credentials returns 401", func(t *testing.T) {
		db := setupTestDB()
		store := storage.NewStorage(db)
		authService := authservice.NewAuthService(store)
		authHandler := NewAuthHandler(authService)

		router := gin.Default()
		router.POST("/api/auth/login", authHandler.Login)

		body := `{"email":"nonexistent@test.com","password":"wrong"}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
