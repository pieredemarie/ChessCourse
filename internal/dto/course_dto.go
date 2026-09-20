package dto

type CourseResponse struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Level       string  `json:"level"`
	Price       float64 `json:"price"`
	CoverURL    string  `json:"cover_url"`
	Duration    int     `json:"duration"` // часов
	Lessons     int     `json:"lessons"`  // уроков
	CategoryID  int     `json:"category_id"`
	Category    string  `json:"category"`
}

type CoursesListRequest struct {
	Search     string `form:"search"`
	Level      string `form:"level"`
	CategoryID string `form:"category_id"`
}

type CoursesListResponse struct {
	Courses []CourseResponse `json:"courses"`
	Total   int64            `json:"total"`
}

type CreateCourseRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description"`
	Level       string  `json:"level" binding:"required,oneof=beginner intermediate advanced"`
	CategoryID  int     `json:"category_id" binding:"required"`
	Price       float64 `json:"price"`
	CoverURL    string  `json:"cover_url"`
	Duration    int     `json:"duration"` // часы
	Lessons     int     `json:"lessons"`  // уроки
}

type UpdateCourseRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Level       string  `json:"level" binding:"omitempty,oneof=beginner intermediate advanced"`
	CategoryID  int     `json:"category_id"`
	Price       float64 `json:"price"`
	CoverURL    string  `json:"cover_url"`
	Duration    int     `json:"duration"`
	Lessons     int     `json:"lessons"`
}
