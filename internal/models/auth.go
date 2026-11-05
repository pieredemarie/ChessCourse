package models

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID           int    `json:"id" gorm:"primaryKey"`
	Name         string `json:"name"`
	LastName     string `json:"last_name"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Role         string `json:"role" gorm:"default:'student'"`
	CreatedAt    time.Time
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}
