package storage

import (
	"chesscourse/internal/models"

	"gorm.io/gorm"
)

type LearningStorage interface {
	GetModulesByCourse(courseID int) ([]models.Module, error)
	GetUserProgress(userID, courseID int) (map[int]bool, error)
	UpdateUserProgress(userID, lessonID int, completed bool) error
}

func (s *Storage) GetModulesByCourse(courseID int) ([]models.Module, error) {
	var modules []models.Module
	err := s.db.Where("course_id = ?", courseID).
		Preload("Lessons", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order asc")
		}).
		Order("sort_order asc").
		Find(&modules).Error
	return modules, err
}

func (s *Storage) GetUserProgress(userID, courseID int) (map[int]bool, error) {
	type ProgressResult struct {
		LessonID  int  `json:"lesson_id"`
		Completed bool `json:"completed"`
	}

	var results []ProgressResult
	err := s.db.Table("user_progresses").
		Select("user_progresses.lesson_id, user_progresses.completed").
		Joins("JOIN lessons ON lessons.id = user_progresses.lesson_id").
		Joins("JOIN modules ON modules.id = lessons.module_id").
		Where("user_progresses.user_id = ? AND modules.course_id = ?", userID, courseID).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	progress := make(map[int]bool)
	for _, r := range results {
		progress[r.LessonID] = r.Completed
	}
	return progress, nil
}

func (s *Storage) UpdateUserProgress(userID, lessonID int, completed bool) error {
	var progress models.UserProgress
	result := s.db.Where("user_id = ? AND lesson_id = ?", userID, lessonID).First(&progress)

	if result.Error != nil {
		progress = models.UserProgress{
			UserID:    userID,
			LessonID:  lessonID,
			Completed: completed,
		}
		return s.db.Create(&progress).Error
	}

	progress.Completed = completed
	return s.db.Save(&progress).Error
}
