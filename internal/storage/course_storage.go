package storage

import (
	"chesscourse/internal/models"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CourseStorage interface {
	GetAllCourses(search, level, categoryID string) ([]models.Course, int64, error)
	GetCourseByID(id int) (*models.Course, error)
	GetTeacherCourses(id int) ([]models.Course, error)
	ReservedSeats(id int) (int64, error)
	AvailableSeats(id int) (int64, error)
	CreateCourse(course *models.Course) error
	UpdateCourse(course *models.Course) error
	DeleteCourse(id int, confirm bool) error
	ArchiveCourse(id int) error
}

func (s *Storage) GetAllCourses(search, level, categoryID string) ([]models.Course, int64, error) {
	var courses []models.Course
	var total int64

	query := s.db.Model(&models.Course{}).Where("archived = ?", false).Preload("Category")

	// Поиск по названию
	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}

	// Фильтр по уровню
	if level != "" {
		query = query.Where("level = ?", level)
	}

	// Фильтр по категории
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	// Считаем общее количество
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Получаем курсы
	err := query.Order("id desc").Find(&courses).Error

	return courses, total, err
}

func (s *Storage) GetCourseByID(id int) (*models.Course, error) {
	var course models.Course
	err := s.db.Preload("Category").First(&course, id).Error
	return &course, err
}

func (s *Storage) CreateCourse(course *models.Course) error {
	return s.db.Create(course).Error
}

func (s *Storage) UpdateCourse(course *models.Course) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var current models.Course
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, course.ID).Error; err != nil {
			return err
		}
		var reserved int64
		if err := tx.Model(&models.Enrollment{}).Where("course_id = ? AND status IN ?", course.ID, []models.EnrollmentStatus{models.StatusApproved, models.StatusPaid}).Count(&reserved).Error; err != nil {
			return err
		}
		if int64(course.Capacity) < reserved {
			return errors.New("capacity is below reserved seats")
		}
		return tx.Save(course).Error
	})
}

func (s *Storage) GetTeacherCourses(id int) ([]models.Course, error) {
	var courses []models.Course
	err := s.db.Where("author_id = ?", id).Preload("Category").Order("id desc").Find(&courses).Error
	return courses, err
}

func (s *Storage) ReservedSeats(id int) (int64, error) {
	var count int64
	err := s.db.Model(&models.Enrollment{}).Where("course_id = ? AND status IN ?", id, []models.EnrollmentStatus{models.StatusApproved, models.StatusPaid}).Count(&count).Error
	return count, err
}

func (s *Storage) AvailableSeats(id int) (int64, error) {
	c, err := s.GetCourseByID(id)
	if err != nil {
		return 0, err
	}
	used, err := s.ReservedSeats(id)
	if err != nil {
		return 0, err
	}
	return int64(c.Capacity) - used, nil
}

func (s *Storage) DeleteCourse(id int, confirm bool) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var current models.Course
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, id).Error; err != nil {
			return err
		}
		var protected int64
		if err := tx.Model(&models.Enrollment{}).Where("course_id = ? AND status IN ?", id, []models.EnrollmentStatus{models.StatusApproved, models.StatusPaid}).Count(&protected).Error; err != nil {
			return err
		}
		if protected > 0 {
			return errors.New("course has approved or paid enrollments; archive it")
		}
		var pending int64
		if err := tx.Model(&models.Enrollment{}).Where("course_id = ? AND status = ?", id, models.StatusPending).Count(&pending).Error; err != nil {
			return err
		}
		if pending > 0 && !confirm {
			return errors.New("confirm deletion of pending enrollments")
		}
		if err := tx.Model(&models.Enrollment{}).Where("course_id = ? AND status = ?", id, models.StatusPending).Update("status", models.StatusCancelled).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Course{}, id).Error
	})
}

func (s *Storage) ArchiveCourse(id int) error {
	return s.db.Model(&models.Course{}).Where("id = ?", id).Update("archived", true).Error
}
