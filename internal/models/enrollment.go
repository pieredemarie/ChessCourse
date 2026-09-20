package models

import "time"

type EnrollmentStatus string

const (
	StatusPending  EnrollmentStatus = "pending"  // ожидает рассмотрения
	StatusApproved EnrollmentStatus = "approved" // одобрено, ожидает оплаты
	StatusPaid     EnrollmentStatus = "paid"     // оплачено, доступ к курсу
	StatusRejected EnrollmentStatus = "rejected" // отклонено
)

type Enrollment struct {
	ID        int              `json:"id" gorm:"primaryKey"`
	UserID    int              `json:"user_id" gorm:"not null;index"`
	User      User             `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CourseID  int              `json:"course_id" gorm:"not null;index"`
	Course    Course           `json:"course,omitempty" gorm:"foreignKey:CourseID"`
	Status    EnrollmentStatus `json:"status" gorm:"default:pending"`
	AdminNote string           `json:"admin_note"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}
