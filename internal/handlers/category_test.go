package handlers

import (
	"bytes"
	"chesscourse/internal/models"
	"chesscourse/internal/service/category"
	"chesscourse/internal/storage"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCategoryHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("create category success", func(t *testing.T) {
		db := setupTestDB() // Твоя функция из прошлых тестов
		store := storage.NewStorage(db)
		service := category.NewCategoryService(store)
		handler := NewCategoryHandler(service)

		router := gin.Default()
		router.POST("/api/admin/categories", handler.CreateCategory)

		body := `{"name":"Стратегия","description":"Основы позиционной игры"}`
		req := httptest.NewRequest(http.MethodPost, "/api/admin/categories", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		// Проверяем, что в базе реально появилась запись
		var cat models.Category
		db.First(&cat, "name = ?", "Стратегия")
		assert.Equal(t, "Основы позиционной игры", cat.Description)
	})

	t.Run("delete category fails if has courses", func(t *testing.T) {
		db := setupTestDB()
		store := storage.NewStorage(db)
		service := category.NewCategoryService(store)
		handler := NewCategoryHandler(service)

		// 1. Создаем категорию
		cat := models.Category{Name: "Эндшпиль"}
		db.Create(&cat)
		// 2. Создаем курс, привязанный к этой категории
		db.Create(&models.Course{Title: "Ладейники", CategoryID: cat.ID})

		router := gin.Default()
		router.DELETE("/api/admin/categories/:id", handler.DeleteCategory)

		req := httptest.NewRequest(http.MethodDelete, "/api/admin/categories/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "cannot delete category with existing courses")
	})

	t.Run("get category by id success", func(t *testing.T) {
		db := setupTestDB()
		db.Create(&models.Category{Name: "Дебюты", Description: "Теория"})

		store := storage.NewStorage(db)
		service := category.NewCategoryService(store)
		handler := NewCategoryHandler(service)

		router := gin.Default()
		router.GET("/api/categories/:id", handler.GetCategoryByID)

		req := httptest.NewRequest(http.MethodGet, "/api/categories/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Дебюты")
	})
}
