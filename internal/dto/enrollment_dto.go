package dto

type CreateEnrollmentRequest struct {
	CourseID int `json:"course_id" binding:"required"`
}

type UpdateEnrollmentRequest struct {
	Status    string `json:"status" binding:"oneof=approved rejected"`
	AdminNote string `json:"admin_note"`
}

type EnrollmentResponse struct {
	ID         int    `json:"id"`
	UserID     int    `json:"user_id"`
	UserName   string `json:"user_name"`
	UserEmail  string `json:"user_email"`
	CourseID   int    `json:"course_id"`
	CourseName string `json:"course_name"`
	Status     string `json:"status"`
	AdminNote  string `json:"admin_note"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type UserEnrollmentResponse struct {
	ID          int    `json:"id"`
	CourseID    int    `json:"course_id"`
	CourseName  string `json:"course_name"`
	CourseCover string `json:"course_cover"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

type UserCourseResponse struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	CoverURL    string  `json:"cover_url"`
	Description string  `json:"description"`
	Duration    int     `json:"duration"`
	Lessons     int     `json:"lessons"`
	Progress    float64 `json:"progress"`
}
