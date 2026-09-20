package handlers

import (
	"chesscourse/internal/dto"
	"chesscourse/internal/service/course"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CourseHandler struct {
	courseService course.CourseService
}

func NewCourseHandler(courseService course.CourseService) *CourseHandler {
	return &CourseHandler{
		courseService: courseService,
	}
}

// GetCourses godoc
// @Summary Получить список курсов с фильтрацией
// @Tags courses
// @Produce json
// @Param search query string false "Поиск по названию"
// @Param level query string false "Уровень (beginner/intermediate/advanced)"
// @Param category_id query string false "ID категории"
// @Success 200 {object} dto.CoursesListResponse
// @Router /api/courses [get]
func (h *CourseHandler) GetCourses(c *gin.Context) {
	search := c.Query("search")
	level := c.Query("level")
	categoryID := c.Query("category_id")

	courses, total, err := h.courseService.GetAllCourses(search, level, categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get courses"})
		return
	}

	response := make([]dto.CourseResponse, len(courses))
	for i, course := range courses {
		response[i] = dto.CourseResponse{
			ID:          course.ID,
			Title:       course.Title,
			Description: course.Description,
			Level:       string(course.Level),
			Price:       course.Price,
			CoverURL:    course.CoverURL,
			Duration:    course.Duration,
			Lessons:     course.Lessons,
			CategoryID:  course.CategoryID,
			Category:    course.Category.Name,
		}
	}

	c.JSON(http.StatusOK, dto.CoursesListResponse{
		Courses: response,
		Total:   total,
	})
}

// GetCourseByID godoc
// @Summary Получить курс по ID
// @Tags courses
// @Produce json
// @Param id path int true "ID курса"
// @Success 200 {object} dto.CourseResponse
// @Router /api/courses/{id} [get]
func (h *CourseHandler) GetCourseByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	course, err := h.courseService.GetCourseByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
		return
	}

	c.JSON(http.StatusOK, dto.CourseResponse{
		ID:          course.ID,
		Title:       course.Title,
		Description: course.Description,
		Level:       string(course.Level),
		Price:       course.Price,
		CoverURL:    course.CoverURL,
		Duration:    course.Duration,
		Lessons:     course.Lessons,
		CategoryID:  course.CategoryID,
		Category:    course.Category.Name,
	})
}

// CreateCourse godoc
// @Summary Создать новый курс
// @Tags courses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateCourseRequest true "Данные курса"
// @Success 201 {object} dto.CourseResponse
// @Router /api/admin/courses [post]
func (h *CourseHandler) CreateCourse(c *gin.Context) {
	var req dto.CreateCourseRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	course, err := h.courseService.CreateCourse(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create course"})
		return
	}

	c.JSON(http.StatusCreated, dto.CourseResponse{
		ID:          course.ID,
		Title:       course.Title,
		Description: course.Description,
		Level:       string(course.Level),
		Price:       course.Price,
		CoverURL:    course.CoverURL,
		Duration:    course.Duration,
		Lessons:     course.Lessons,
		CategoryID:  course.CategoryID,
		Category:    course.Category.Name,
	})
}

// UpdateCourse godoc
// @Summary Обновить курс
// @Tags courses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID курса"
// @Param request body dto.UpdateCourseRequest true "Данные курса"
// @Success 200 {object} dto.CourseResponse
// @Router /api/admin/courses/{id} [put]
func (h *CourseHandler) UpdateCourse(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	var req dto.UpdateCourseRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	course, err := h.courseService.UpdateCourse(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update course"})
		return
	}

	c.JSON(http.StatusOK, dto.CourseResponse{
		ID:          course.ID,
		Title:       course.Title,
		Description: course.Description,
		Level:       string(course.Level),
		Price:       course.Price,
		CoverURL:    course.CoverURL,
		Duration:    course.Duration,
		Lessons:     course.Lessons,
		CategoryID:  course.CategoryID,
		Category:    course.Category.Name,
	})
}

// DeleteCourse godoc
// @Summary Удалить курс
// @Tags courses
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID курса"
// @Success 204
// @Router /api/admin/courses/{id} [delete]
func (h *CourseHandler) DeleteCourse(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	if err := h.courseService.DeleteCourse(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete course"})
		return
	}

	c.Status(http.StatusNoContent)
}
