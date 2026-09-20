package models

import "time"

type CourseLevel string
type Category struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;size:100"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`

	Courses []Course `gorm:"foreignKey:CategoryID" json:"courses,omitempty"`
}

const (
	LevelBeginner     CourseLevel = "beginner"     // Новичок
	LevelIntermediate CourseLevel = "intermediate" // Средний
	LevelAdvanced     CourseLevel = "advanced"     // Продвинутый
)

type Course struct {
	ID          int         `json:"id" gorm:"primaryKey"`
	Title       string      `json:"title" gorm:"size:200;not null"`
	Description string      `json:"description" gorm:"type:text"`
	Level       CourseLevel `json:"level"`
	Price       float64     `json:"price"`
	CoverURL    string      `json:"cover_url"` // URL картинки
	Duration    int         `json:"duration"`  // количество часов
	Lessons     int         `json:"lessons"`   // количество уроков
	AuthorID    int         `json:"author_id"`
	CategoryID  int         `json:"category_id" gorm:"index"`
	Category    Category    `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	CreatedAt   time.Time   `json:"created_at"`
}
