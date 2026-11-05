package handlers

import (
	"chesscourse/internal/dto"
	"chesscourse/internal/models"
	"chesscourse/internal/service/enrollment"
	"chesscourse/internal/service/learning"
	"chesscourse/internal/storage"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EnrollmentHandler struct {
	enrollmentService enrollment.EnrollmentService
	learningService   learning.LearningService
}

func NewEnrollmentHandler(e enrollment.EnrollmentService, l learning.LearningService) *EnrollmentHandler {
	return &EnrollmentHandler{e, l}
}

func enrollmentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "enrollment or course not found"})
	case errors.Is(err, storage.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
	case errors.Is(err, storage.ErrNoSeats):
		c.JSON(http.StatusConflict, gin.H{"error": "Нет свободных мест"})
	case errors.Is(err, storage.ErrEnrollmentConflict), errors.Is(err, storage.ErrCourseUnavailable), errors.Is(err, storage.ErrPaymentExpired):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, enrollment.ErrInvalidDecision), errors.Is(err, enrollment.ErrInvalidReason):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, enrollment.ErrPaymentFailed):
		c.JSON(http.StatusPaymentRequired, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "enrollment operation failed"})
	}
}
func enrollmentID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid enrollment id"})
		return 0, false
	}
	return id, true
}
func student(c *gin.Context) bool {
	if c.GetString("role") != models.RoleStudent {
		c.JSON(http.StatusForbidden, gin.H{"error": "student role required"})
		return false
	}
	return true
}

// CreateEnrollment godoc
// @Summary Оставить заявку на курс
// @Tags enrollments
// @Security BearerAuth
// @Accept json
// @Param request body dto.CreateEnrollmentRequest true "Курс"
// @Success 201
// @Failure 409 {object} map[string]string
// @Router /api/enrollments [post]
func (h *EnrollmentHandler) CreateEnrollment(c *gin.Context) {
	if !student(c) {
		return
	}
	var req dto.CreateEnrollmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.enrollmentService.CreateEnrollment(c.Request.Context(), c.GetInt("userID"), req.CourseID); err != nil {
		enrollmentError(c, err)
		return
	}
	c.Status(http.StatusCreated)
}

// CancelEnrollment godoc
// @Summary Отменить ожидающую заявку
// @Tags enrollments
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Success 204
// @Router /api/enrollments/{id}/cancel [post]
func (h *EnrollmentHandler) CancelEnrollment(c *gin.Context) {
	if !student(c) {
		return
	}
	id, ok := enrollmentID(c)
	if !ok {
		return
	}
	if err := h.enrollmentService.CancelEnrollment(c.Request.Context(), c.GetInt("userID"), id); err != nil {
		enrollmentError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetMyEnrollments godoc
// @Summary Мои заявки
// @Tags enrollments
// @Security BearerAuth
// @Success 200 {array} dto.UserEnrollmentResponse
// @Router /api/enrollments [get]
func (h *EnrollmentHandler) GetMyEnrollments(c *gin.Context) {
	rows, err := h.enrollmentService.GetUserEnrollments(c.Request.Context(), c.GetInt("userID"))
	if err != nil {
		enrollmentError(c, err)
		return
	}
	out := make([]dto.UserEnrollmentResponse, 0, len(rows))
	for _, e := range rows {
		r := dto.UserEnrollmentResponse{ID: e.ID, CourseID: e.CourseID, CourseName: e.Course.Title, CourseCover: e.Course.CoverURL, Status: string(e.Status), RejectionReason: e.RejectionReason, CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z07:00")}
		if e.PaymentDueAt != nil {
			r.PaymentDueAt = e.PaymentDueAt.Format("2006-01-02T15:04:05Z07:00")
		}
		out = append(out, r)
	}
	c.JSON(http.StatusOK, out)
}

// GetAllEnrollments godoc
// @Summary Заявки на курсы преподавателя
// @Tags enrollments
// @Security BearerAuth
// @Param status query string false "Фильтр статуса"
// @Success 200 {array} dto.EnrollmentResponse
// @Router /api/teacher/enrollments [get]
func (h *EnrollmentHandler) GetAllEnrollments(c *gin.Context) {
	rows, err := h.enrollmentService.GetTeacherEnrollments(c.Request.Context(), c.GetInt("userID"), c.Query("status"))
	if err != nil {
		enrollmentError(c, err)
		return
	}
	out := make([]dto.EnrollmentResponse, 0, len(rows))
	for _, e := range rows {
		r := dto.EnrollmentResponse{ID: e.ID, UserID: e.UserID, UserName: e.User.Name, UserEmail: e.User.Email, CourseID: e.CourseID, CourseName: e.Course.Title, Status: string(e.Status), AdminNote: e.AdminNote, RejectionReason: e.RejectionReason, CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), UpdatedAt: e.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")}
		if e.PaymentDueAt != nil {
			r.PaymentDueAt = e.PaymentDueAt.Format("2006-01-02T15:04:05Z07:00")
		}
		out = append(out, r)
	}
	c.JSON(http.StatusOK, out)
}

// UpdateEnrollmentStatus godoc
// @Summary Одобрить или отклонить заявку
// @Tags enrollments
// @Security BearerAuth
// @Accept json
// @Param id path int true "ID заявки"
// @Param request body dto.UpdateEnrollmentRequest true "Решение"
// @Success 204
// @Router /api/teacher/enrollments/{id} [put]
func (h *EnrollmentHandler) UpdateEnrollmentStatus(c *gin.Context) {
	id, ok := enrollmentID(c)
	if !ok {
		return
	}
	var req dto.UpdateEnrollmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	reason := req.RejectionReason
	if reason == "" {
		reason = req.AdminNote
	}
	if err := h.enrollmentService.DecideEnrollment(c.Request.Context(), c.GetInt("userID"), id, req.Status, reason); err != nil {
		enrollmentError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// PayForEnrollment godoc
// @Summary Подтвердить тестовую оплату
// @Description Тестовый режим без списания денег. success=true оплачивает, success=false возвращает ошибку без смены статуса.
// @Tags enrollments
// @Security BearerAuth
// @Accept json
// @Param id path int true "ID заявки"
// @Param request body dto.TestPaymentRequest true "Результат тестовой оплаты"
// @Success 200 {object} map[string]string
// @Router /api/enrollments/{id}/pay [post]
func (h *EnrollmentHandler) PayForEnrollment(c *gin.Context) {
	if !student(c) {
		return
	}
	id, ok := enrollmentID(c)
	if !ok {
		return
	}
	var req dto.TestPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Success == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "success is required in test payment"})
		return
	}
	if err := h.enrollmentService.PayForEnrollment(c.Request.Context(), c.GetInt("userID"), id, *req.Success); err != nil {
		enrollmentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "paid", "payment_mode": "test"})
}

// GetMyCourses godoc
// @Summary Мои оплаченные курсы
// @Tags enrollments
// @Security BearerAuth
// @Success 200 {array} dto.UserCourseResponse
// @Router /api/my-courses [get]
func (h *EnrollmentHandler) GetMyCourses(c *gin.Context) {
	rows, err := h.enrollmentService.GetUserEnrollments(c.Request.Context(), c.GetInt("userID"))
	if err != nil {
		enrollmentError(c, err)
		return
	}
	out := make([]dto.UserCourseResponse, 0)
	for _, e := range rows {
		if e.Status != models.StatusPaid {
			continue
		}
		_, _, percent, err := h.learningService.GetCourseProgress(c.GetInt("userID"), e.CourseID)
		if err != nil {
			percent = 0
		}
		out = append(out, dto.UserCourseResponse{ID: e.CourseID, Title: e.Course.Title, CoverURL: e.Course.CoverURL, Description: e.Course.Description, Duration: e.Course.Duration, Lessons: e.Course.Lessons, Progress: percent})
	}
	c.JSON(http.StatusOK, out)
}
