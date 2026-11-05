package course

import (
	"chesscourse/internal/dto"
	"chesscourse/internal/models"
	"chesscourse/internal/storage"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

var ErrForbidden = errors.New("course does not belong to teacher")
var ErrInvalid = errors.New("invalid course")
var ErrCapacity = errors.New("capacity is below reserved seats")
var ErrHasEnrollments = errors.New("course has approved or paid enrollments; archive it")

type CourseService interface {
	GetAllCourses(search, level, categoryID string) ([]models.Course, int64, error)
	GetTeacherCourses(teacherID int) ([]models.Course, error)
	GetCourseByID(id int) (*models.Course, error)
	CreateCourse(teacherID int, req dto.CreateCourseRequest) (*models.Course, error)
	UpdateCourse(teacherID, id int, req dto.UpdateCourseRequest) (*models.Course, error)
	DeleteCourse(teacherID, id int, confirm bool) error
	ArchiveCourse(teacherID, id int) error
	AvailableSeats(id int) (int64, error)
}

type courseService struct{ storage storage.CourseStorage }

func NewCourseService(store storage.CourseStorage) *courseService {
	return &courseService{storage: store}
}

func (s *courseService) GetAllCourses(search, level, categoryID string) ([]models.Course, int64, error) {
	return s.storage.GetAllCourses(search, level, categoryID)
}
func (s *courseService) GetTeacherCourses(id int) ([]models.Course, error) {
	return s.storage.GetTeacherCourses(id)
}
func (s *courseService) GetCourseByID(id int) (*models.Course, error) {
	return s.storage.GetCourseByID(id)
}
func (s *courseService) AvailableSeats(id int) (int64, error) { return s.storage.AvailableSeats(id) }

func parseDate(value string) (time.Time, error) {
	d, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid start_date: %w", ErrInvalid)
	}
	return d, nil
}

func validate(c *models.Course) error {
	if n := len([]rune(strings.TrimSpace(c.Title))); n < 3 || n > 100 {
		return fmt.Errorf("title must be 3-100 characters: %w", ErrInvalid)
	}
	if n := len([]rune(strings.TrimSpace(c.Description))); n < 10 || n > 300 {
		return fmt.Errorf("description must be 10-300 characters: %w", ErrInvalid)
	}
	if c.Level != models.LevelBeginner && c.Level != models.LevelIntermediate && c.Level != models.LevelAdvanced {
		return fmt.Errorf("invalid level: %w", ErrInvalid)
	}
	if c.CategoryID <= 0 {
		return fmt.Errorf("category is required: %w", ErrInvalid)
	}
	if c.Format != models.FormatInPerson && c.Format != models.FormatOnlineGroup {
		return fmt.Errorf("invalid format: %w", ErrInvalid)
	}
	if c.Format == models.FormatInPerson && strings.TrimSpace(c.Location) == "" {
		return fmt.Errorf("location is required: %w", ErrInvalid)
	}
	if c.Capacity < 1 || c.Capacity > 20 {
		return fmt.Errorf("capacity must be 1-20: %w", ErrInvalid)
	}
	if c.StartDate.IsZero() || c.StartDate.Format("2006-01-02") < time.Now().Format("2006-01-02") {
		return fmt.Errorf("start_date is in the past: %w", ErrInvalid)
	}
	if math.IsNaN(c.Price) || math.IsInf(c.Price, 0) || c.Price < 100 || c.Price > 100000 || math.Abs(c.Price*100-math.Round(c.Price*100)) > 1e-7 {
		return fmt.Errorf("price must be 100.00-100000.00 with two decimals: %w", ErrInvalid)
	}
	return nil
}

func (s *courseService) CreateCourse(teacherID int, req dto.CreateCourseRequest) (*models.Course, error) {
	d, err := parseDate(req.StartDate)
	if err != nil {
		return nil, err
	}
	c := &models.Course{Title: req.Title, Description: req.Description, Level: models.CourseLevel(req.Level), CategoryID: req.CategoryID, Price: req.Price, Format: models.CourseFormat(req.Format), Capacity: req.Capacity, StartDate: d, Location: req.Location, CoverURL: req.CoverURL, Duration: req.Duration, Lessons: req.Lessons, AuthorID: teacherID}
	if err := validate(c); err != nil {
		return nil, err
	}
	if err := s.storage.CreateCourse(c); err != nil {
		return nil, fmt.Errorf("create course: %w", err)
	}
	return s.storage.GetCourseByID(c.ID)
}

func (s *courseService) UpdateCourse(teacherID, id int, req dto.UpdateCourseRequest) (*models.Course, error) {
	c, err := s.storage.GetCourseByID(id)
	if err != nil {
		return nil, err
	}
	if c.AuthorID != teacherID {
		return nil, ErrForbidden
	}
	if req.Title != nil {
		c.Title = *req.Title
	}
	if req.Description != nil {
		c.Description = *req.Description
	}
	if req.Level != nil {
		c.Level = models.CourseLevel(*req.Level)
	}
	if req.CategoryID != nil {
		c.CategoryID = *req.CategoryID
	}
	if req.Price != nil {
		c.Price = *req.Price
	}
	if req.Format != nil {
		c.Format = models.CourseFormat(*req.Format)
	}
	if req.Capacity != nil {
		c.Capacity = *req.Capacity
	}
	if req.Location != nil {
		c.Location = *req.Location
	}
	if req.StartDate != nil {
		c.StartDate, err = parseDate(*req.StartDate)
		if err != nil {
			return nil, err
		}
	}
	if req.CoverURL != "" {
		c.CoverURL = req.CoverURL
	}
	if req.Duration != 0 {
		c.Duration = req.Duration
	}
	if req.Lessons != 0 {
		c.Lessons = req.Lessons
	}
	if err := validate(c); err != nil {
		return nil, err
	}
	used, err := s.storage.ReservedSeats(id)
	if err != nil {
		return nil, err
	}
	if int64(c.Capacity) < used {
		return nil, ErrCapacity
	}
	if err := s.storage.UpdateCourse(c); err != nil {
		return nil, fmt.Errorf("update course: %w", err)
	}
	return s.storage.GetCourseByID(id)
}

func (s *courseService) DeleteCourse(teacherID, id int, confirm bool) error {
	c, err := s.storage.GetCourseByID(id)
	if err != nil {
		return err
	}
	if c.AuthorID != teacherID {
		return ErrForbidden
	}
	return s.storage.DeleteCourse(id, confirm)
}
func (s *courseService) ArchiveCourse(teacherID, id int) error {
	c, err := s.storage.GetCourseByID(id)
	if err != nil {
		return err
	}
	if c.AuthorID != teacherID {
		return ErrForbidden
	}
	return s.storage.ArchiveCourse(id)
}
