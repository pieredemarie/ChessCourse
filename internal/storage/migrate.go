package storage

import (
	"chesscourse/internal/models"
	"log"

	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Course{},
		&models.Enrollment{},
		&models.Module{},
		&models.Lesson{},
		&models.UserProgress{},
	)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("migrations applied successfully")
}
