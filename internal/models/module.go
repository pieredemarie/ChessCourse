package models

import "time"

type Module struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	CourseID    int       `json:"course_id" gorm:"not null;index"`
	Course      Course    `json:"course,omitempty" gorm:"foreignKey:CourseID"`
	Title       string    `json:"title" gorm:"size:200;not null"`
	Description string    `json:"description" gorm:"type:text"`
	SortOrder   int       `json:"order" gorm:"column:sort_order"` // порядковый номер модуля
	CreatedAt   time.Time `json:"created_at"`

	Lessons []Lesson `json:"lessons,omitempty" gorm:"foreignKey:ModuleID"`
}
