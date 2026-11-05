package handlers

import (
	"chesscourse/internal/apperror"
	"chesscourse/internal/dto"
	"chesscourse/internal/models"
	coursesvc "chesscourse/internal/service/course"
	"errors"
	"gorm.io/gorm"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CourseHandler struct {
	courseService coursesvc.CourseService
}

func (h *CourseHandler) response(course models.Course) (dto.CourseResponse, error) {
	seats, err := h.courseService.AvailableSeats(course.ID)
	if err != nil {
		return dto.CourseResponse{}, err
	}
	return dto.CourseResponse{ID: course.ID, Title: course.Title, Description: course.Description, Level: string(course.Level), Price: course.Price, Format: string(course.Format), Capacity: course.Capacity, StartDate: course.StartDate.Format("2006-01-02"), Location: course.Location, Archived: course.Archived, AvailableSeats: seats, CoverURL: course.CoverURL, Duration: course.Duration, Lessons: course.Lessons, CategoryID: course.CategoryID, Category: course.Category.Name}, nil
}

// GetTeacherCourses godoc
// @Summary Курсы текущего преподавателя
// @Tags courses
// @Security BearerAuth
// @Success 200 {object} dto.CoursesListResponse
// @Router /api/teacher/courses [get]
func (h *CourseHandler) GetTeacherCourses(c *gin.Context) {
	id := c.GetInt("userID")
	courses, err := h.courseService.GetTeacherCourses(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get courses"})
		return
	}
	response := make([]dto.CourseResponse, len(courses))
	for i, course := range courses {
		response[i], err = h.response(course)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get available seats"})
			return
		}
	}
	c.JSON(http.StatusOK, dto.CoursesListResponse{Courses: response, Total: int64(len(response))})
}

func courseError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, coursesvc.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
	case errors.Is(err, coursesvc.ErrInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, coursesvc.ErrCapacity):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	}
}

// ArchiveCourse godoc
// @Summary Архивировать свой курс
// @Tags courses
// @Security BearerAuth
// @Param id path int true "ID курса"
// @Success 204
// @Router /api/teacher/courses/{id}/archive [post]
func (h *CourseHandler) ArchiveCourse(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}
	if err := h.courseService.ArchiveCourse(c.GetInt("userID"), id); err != nil {
		courseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func NewCourseHandler(courseService coursesvc.CourseService) *CourseHandler {
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
		seats, seatErr := h.courseService.AvailableSeats(course.ID)
		if seatErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get available seats"})
			return
		}
		response[i] = dto.CourseResponse{
			AvailableSeats: seats,
			ID:             course.ID,
			Title:          course.Title,
			Description:    course.Description,
			Level:          string(course.Level),
			Price:          course.Price,
			Format:         string(course.Format), Capacity: course.Capacity, StartDate: course.StartDate.Format("2006-01-02"), Location: course.Location, Archived: course.Archived,
			CoverURL:   course.CoverURL,
			Duration:   course.Duration,
			Lessons:    course.Lessons,
			CategoryID: course.CategoryID,
			Category:   course.Category.Name,
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

	seats, err := h.courseService.AvailableSeats(course.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get available seats"})
		return
	}
	c.JSON(http.StatusOK, dto.CourseResponse{
		AvailableSeats: seats,
		ID:             course.ID,
		Title:          course.Title,
		Description:    course.Description,
		Level:          string(course.Level),
		Price:          course.Price,
		Format:         string(course.Format), Capacity: course.Capacity, StartDate: course.StartDate.Format("2006-01-02"), Location: course.Location, Archived: course.Archived,
		CoverURL:   course.CoverURL,
		Duration:   course.Duration,
		Lessons:    course.Lessons,
		CategoryID: course.CategoryID,
		Category:   course.Category.Name,
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
// @Router /api/teacher/courses [post]
func (h *CourseHandler) CreateCourse(c *gin.Context) {
	var req dto.CreateCourseRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	course, err := h.courseService.CreateCourse(c.GetInt("userID"), req)
	switch err {

	}
	if err != nil {
		if errors.Is(err, coursesvc.ErrInvalid) {
			courseError(c, err)
			return
		}
		switch {
		case errors.Is(err, apperror.ErrInvalidTitleCourse):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid title"})
			return
		case errors.Is(err, apperror.ErrInvalidDescCourse):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid description"})
			return
		case errors.Is(err, apperror.ErrInvalidPriceCourse):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid price"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create course"})
		return
	}

	seats, err := h.courseService.AvailableSeats(course.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get available seats"})
		return
	}
	c.JSON(http.StatusCreated, dto.CourseResponse{
		AvailableSeats: seats,
		ID:             course.ID,
		Title:          course.Title,
		Description:    course.Description,
		Level:          string(course.Level),
		Price:          course.Price,
		Format:         string(course.Format), Capacity: course.Capacity, StartDate: course.StartDate.Format("2006-01-02"), Location: course.Location, Archived: course.Archived,
		CoverURL:   course.CoverURL,
		Duration:   course.Duration,
		Lessons:    course.Lessons,
		CategoryID: course.CategoryID,
		Category:   course.Category.Name,
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
// @Router /api/teacher/courses/{id} [put]
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

	course, err := h.courseService.UpdateCourse(c.GetInt("userID"), id, req)
	if err != nil {
		courseError(c, err)
		return
	}

	seats, err := h.courseService.AvailableSeats(course.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get available seats"})
		return
	}
	c.JSON(http.StatusOK, dto.CourseResponse{
		AvailableSeats: seats,
		ID:             course.ID,
		Title:          course.Title,
		Description:    course.Description,
		Level:          string(course.Level),
		Price:          course.Price,
		Format:         string(course.Format), Capacity: course.Capacity, StartDate: course.StartDate.Format("2006-01-02"), Location: course.Location, Archived: course.Archived,
		CoverURL:   course.CoverURL,
		Duration:   course.Duration,
		Lessons:    course.Lessons,
		CategoryID: course.CategoryID,
		Category:   course.Category.Name,
	})
}

// DeleteCourse godoc
// @Summary Удалить курс
// @Tags courses
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID курса"
// @Success 204
// @Router /api/teacher/courses/{id} [delete]
func (h *CourseHandler) DeleteCourse(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	if err := h.courseService.DeleteCourse(c.GetInt("userID"), id, c.Query("confirm") == "true"); err != nil {
		courseError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
