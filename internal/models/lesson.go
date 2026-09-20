package models

import "time"

type Lesson struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	ModuleID    int       `json:"module_id" gorm:"not null;index"`
	Module      Module    `json:"module,omitempty" gorm:"foreignKey:ModuleID"`
	Title       string    `json:"title" gorm:"size:200;not null"`
	Description string    `json:"description" gorm:"type:text"`
	VideoURL    string    `json:"video_url"`
	Duration    int       `json:"duration"`
	SortOrder   int       `json:"order" gorm:"column:sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserProgress struct {
	ID          int        `json:"id" gorm:"primaryKey"`
	UserID      int        `json:"user_id" gorm:"not null;index"`
	LessonID    int        `json:"lesson_id" gorm:"not null;index"`
	Lesson      Lesson     `json:"lesson,omitempty" gorm:"foreignKey:LessonID"`
	Completed   bool       `json:"completed" gorm:"default:false"`
	CompletedAt *time.Time `json:"completed_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
