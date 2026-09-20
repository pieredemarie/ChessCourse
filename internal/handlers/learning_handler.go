package handlers

import (
	"chesscourse/internal/dto"
	"chesscourse/internal/service/learning"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LearningHandler struct {
	learningService learning.LearningService
}

func NewLearningHandler(learningService learning.LearningService) *LearningHandler {
	return &LearningHandler{
		learningService: learningService,
	}
}

// GetCourseModules возвращает модули и уроки курса с прогрессом пользователя
func (h *LearningHandler) GetCourseModules(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	courseID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	// Получаем модули с уроками
	modules, err := h.learningService.GetCourseModules(courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get modules"})
		return
	}

	// Получаем прогресс пользователя
	progress, err := h.learningService.GetUserProgressMap(userID.(int), courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get progress"})
		return
	}

	// Формируем ответ
	response := make([]dto.ModuleResponse, len(modules))
	for i, module := range modules {
		lessons := make([]dto.LessonResponse, len(module.Lessons))
		for j, lesson := range module.Lessons {
			completed := false
			if val, ok := progress[lesson.ID]; ok {
				completed = val
			}
			lessons[j] = dto.LessonResponse{
				ID:          lesson.ID,
				Title:       lesson.Title,
				Description: lesson.Description,
				VideoURL:    lesson.VideoURL,
				Duration:    lesson.Duration,
				Order:       lesson.SortOrder,
				Completed:   completed,
			}
		}
		response[i] = dto.ModuleResponse{
			ID:          module.ID,
			Title:       module.Title,
			Description: module.Description,
			Order:       module.SortOrder,
			Lessons:     lessons,
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetCourseProgress возвращает прогресс пользователя по курсу
func (h *LearningHandler) GetCourseProgress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	courseID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	total, completed, percent, err := h.learningService.GetCourseProgress(userID.(int), courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get progress"})
		return
	}

	c.JSON(http.StatusOK, dto.CourseProgressResponse{
		CourseID:         courseID,
		TotalLessons:     total,
		CompletedLessons: completed,
		ProgressPercent:  percent,
	})
}

// UpdateLessonProgress обновляет прогресс по уроку
func (h *LearningHandler) UpdateLessonProgress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	lessonID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}

	var req dto.UpdateProgressRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	err = h.learningService.UpdateLessonProgress(userID.(int), lessonID, req.Completed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update progress"})
		return
	}

	c.Status(http.StatusNoContent)
}
