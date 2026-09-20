package dto

type LessonResponse struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	VideoURL    string `json:"video_url"`
	Duration    int    `json:"duration"`
	Order       int    `json:"order"`
	Completed   bool   `json:"completed"`
}

type ModuleResponse struct {
	ID          int              `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Order       int              `json:"order"`
	Lessons     []LessonResponse `json:"lessons"`
}

type CourseProgressResponse struct {
	CourseID         int     `json:"course_id"`
	TotalLessons     int     `json:"total_lessons"`
	CompletedLessons int     `json:"completed_lessons"`
	ProgressPercent  float64 `json:"progress_percent"`
}

type UpdateProgressRequest struct {
	Completed bool `json:"completed"`
}
