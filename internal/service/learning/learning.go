package learning

import (
	"chesscourse/internal/models"
	"chesscourse/internal/storage"
)

type LearningService interface {
	GetCourseModules(courseID int) ([]models.Module, error)
	GetUserProgressMap(userID, courseID int) (map[int]bool, error)
	GetCourseProgress(userID, courseID int) (totalLessons, completedLessons int, percent float64, err error)
	UpdateLessonProgress(userID, lessonID int, completed bool) error
}

type learningService struct {
	storage storage.LearningStorage
}

func NewLearningService(storage storage.LearningStorage) *learningService {
	return &learningService{storage: storage}
}

func (s *learningService) GetCourseModules(courseID int) ([]models.Module, error) {
	return s.storage.GetModulesByCourse(courseID)
}

func (s *learningService) GetUserProgressMap(userID, courseID int) (map[int]bool, error) {
	return s.storage.GetUserProgress(userID, courseID)
}

func (s *learningService) GetCourseProgress(userID, courseID int) (totalLessons, completedLessons int, percent float64, err error) {
	modules, err := s.storage.GetModulesByCourse(courseID)
	if err != nil {
		return 0, 0, 0, err
	}

	for _, module := range modules {
		totalLessons += len(module.Lessons)
	}

	progress, err := s.storage.GetUserProgress(userID, courseID)
	if err != nil {
		return totalLessons, 0, 0, err
	}

	for _, completed := range progress {
		if completed {
			completedLessons++
		}
	}

	if totalLessons > 0 {
		percent = float64(completedLessons) / float64(totalLessons) * 100
	}

	return totalLessons, completedLessons, percent, nil
}

func (s *learningService) UpdateLessonProgress(userID, lessonID int, completed bool) error {
	return s.storage.UpdateUserProgress(userID, lessonID, completed)
}
