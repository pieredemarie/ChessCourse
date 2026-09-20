package models

import "time"

type User struct {
	ID           int    `json:"id" gorm:"primaryKey"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Role         string `json:"role" gorm:"default:'student'"`
	CreatedAt    time.Time
}
