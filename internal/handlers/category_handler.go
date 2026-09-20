package handlers

import (
	"chesscourse/internal/apperror"
	"chesscourse/internal/dto"
	"chesscourse/internal/service/category"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryService category.CategoryService
}

func NewCategoryHandler(categoryService category.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// GetCategories godoc
//
// @Summary Получить все категории
// @Description Возвращает список всех категорий курсов. Доступно без авторизации.
// @Tags categories
// @Produce json
// @Success 200 {array} dto.CategoryResponse "Список категорий"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/categories [get]
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	categories, err := h.categoryService.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GetCategoryByID godoc
//
// @Summary Получить категорию по ID
// @Description Возвращает информацию о категории по её идентификатору.
// @Tags categories
// @Produce json
// @Param id path int true "ID категории"
// @Success 200 {object} dto.CategoryResponse "Категория найдена"
// @Failure 400 {object} map[string]string "Неверный ID категории"
// @Failure 404 {object} map[string]string "Категория не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/categories/{id} [get]
func (h *CategoryHandler) GetCategoryByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	category, err := h.categoryService.GetCategoryByID(id)
	if err != nil {
		if errors.Is(err, apperror.ErrCategoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get category"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// CreateCategory godoc
//
// @Summary Создать новую категорию
// @Description Создаёт новую категорию курсов. Требует права администратора.
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateCategoryRequest true "Данные для создания категории"
// @Success 201 "Категория успешно создана"
// @Failure 400 {object} map[string]string "Некорректный запрос (проверьте поля)"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Доступ запрещен (требуется роль admin)"
// @Failure 409 {object} map[string]string "Категория с таким названием уже существует"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/admin/categories [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req dto.CreateCategoryRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	err := h.categoryService.CreateCategory(req.Name, req.Description)
	if err != nil {
		if errors.Is(err, apperror.ErrCategoryExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "category already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create category"})
		return
	}

	c.Status(http.StatusCreated)
}

// UpdateCategory godoc
//
// @Summary Обновить категорию
// @Description Обновляет название и/или описание категории. Требует права администратора.
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID категории"
// @Param request body dto.UpdateCategoryRequest true "Данные для обновления категории"
// @Success 204 "Категория успешно обновлена"
// @Failure 400 {object} map[string]string "Некорректный запрос (неверный ID или данные)"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Доступ запрещен (требуется роль admin)"
// @Failure 404 {object} map[string]string "Категория не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/admin/categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	var req dto.UpdateCategoryRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	err = h.categoryService.UpdateCategory(id, req.Name, req.Description)
	if errors.Is(err, apperror.ErrCategoryNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update category"})
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteCategory godoc
//
// @Summary Удалить категорию
// @Description Удаляет категорию по ID. Требует права администратора.
// @Tags categories
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID категории"
// @Success 204 "Категория успешно удалена"
// @Failure 400 {object} map[string]string "Неверный ID категории"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Доступ запрещен (требуется роль admin)"
// @Failure 404 {object} map[string]string "Категория не найдена"
// @Failure 409 {object} map[string]string "Невозможно удалить: категория содержит курсы"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/admin/categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	err = h.categoryService.DeleteCategory(id)
	if errors.Is(err, apperror.ErrCategoryNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}
	if errors.Is(err, apperror.ErrCategoryHasCourses) {
		c.JSON(http.StatusConflict, gin.H{"error": "cannot delete category with existing courses"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete category"})
		return
	}

	c.Status(http.StatusNoContent)
}
