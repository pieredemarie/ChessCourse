package handlers

import (
	"chesscourse/internal/dto"
	"chesscourse/internal/service/enrollment"
	"chesscourse/internal/service/learning"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EnrollmentHandler struct {
	enrollmentService enrollment.EnrollmentService
	learningService   learning.LearningService // Добавить
}

func NewEnrollmentHandler(enrollmentService enrollment.EnrollmentService, learningService learning.LearningService) *EnrollmentHandler {
	return &EnrollmentHandler{
		enrollmentService: enrollmentService,
		learningService:   learningService,
	}
}

// CreateEnrollment - пользователь подает заявку на курс
// POST /api/enrollments
func (h *EnrollmentHandler) CreateEnrollment(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.CreateEnrollmentRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	err := h.enrollmentService.CreateEnrollment(userID.(int), req.CourseID)
	if err != nil {
		//if (errors.Is(...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

// GetMyEnrollments - пользователь получает свои заявки
// GET /api/enrollments
func (h *EnrollmentHandler) GetMyEnrollments(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	enrollments, err := h.enrollmentService.GetUserEnrollments(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get enrollments"})
		return
	}

	response := make([]dto.UserEnrollmentResponse, len(enrollments))
	for i, e := range enrollments {
		response[i] = dto.UserEnrollmentResponse{
			ID:          e.ID,
			CourseID:    e.CourseID,
			CourseName:  e.Course.Title,
			CourseCover: e.Course.CoverURL,
			Status:      string(e.Status),
			CreatedAt:   e.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetAllEnrollments - админ получает все заявки
// GET /api/admin/enrollments
func (h *EnrollmentHandler) GetAllEnrollments(c *gin.Context) {
	status := c.Query("status")
	enrollments, err := h.enrollmentService.GetAllEnrollments(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get enrollments"})
		return
	}

	response := make([]dto.EnrollmentResponse, len(enrollments))
	for i, e := range enrollments {
		response[i] = dto.EnrollmentResponse{
			ID:         e.ID,
			UserID:     e.UserID,
			UserName:   e.User.Name,
			UserEmail:  e.User.Email,
			CourseID:   e.CourseID,
			CourseName: e.Course.Title,
			Status:     string(e.Status),
			AdminNote:  e.AdminNote,
			CreatedAt:  e.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  e.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, response)
}

// UpdateEnrollmentStatus - админ обновляет статус заявки
// PUT /api/admin/enrollments/:id
func (h *EnrollmentHandler) UpdateEnrollmentStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid enrollment id"})
		return
	}

	var req dto.UpdateEnrollmentRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	err = h.enrollmentService.UpdateEnrollmentStatus(id, req.Status, req.AdminNote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update enrollment"})
		return
	}

	c.Status(http.StatusNoContent)
}

// PayForEnrollment - пользователь оплачивает курс
// POST /api/enrollments/:id/pay
func (h *EnrollmentHandler) PayForEnrollment(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	enrollmentID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid enrollment id"})
		return
	}

	// Проверяем, что заявка принадлежит пользователю
	enrollment, err := h.enrollmentService.GetEnrollmentByID(enrollmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "enrollment not found"})
		return
	}

	if enrollment.UserID != userID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	err = h.enrollmentService.PayForEnrollment(enrollmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "paid"})
}

func (h *EnrollmentHandler) GetMyCourses(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	enrollments, err := h.enrollmentService.GetUserEnrollments(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get courses"})
		return
	}

	var paidCourses []dto.UserCourseResponse
	for _, e := range enrollments {
		if e.Status == "paid" {
			// Получаем прогресс по курсу
			_, _, percent, _ := h.learningService.GetCourseProgress(userID.(int), e.Course.ID)

			paidCourses = append(paidCourses, dto.UserCourseResponse{
				ID:          e.Course.ID,
				Title:       e.Course.Title,
				CoverURL:    e.Course.CoverURL,
				Description: e.Course.Description,
				Duration:    e.Course.Duration,
				Lessons:     e.Course.Lessons,
				Progress:    percent,
			})
		}
	}

	c.JSON(http.StatusOK, paidCourses)
}
