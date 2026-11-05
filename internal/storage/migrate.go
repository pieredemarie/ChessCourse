package storage

import (
	"chesscourse/internal/models"
	"log"
	"time"

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
	// Repeatable backfill for courses created before group-course fields existed.
	if err := db.Model(&models.Course{}).Where("format IS NULL OR format = ''").Update("format", models.FormatOnlineGroup).Error; err != nil {
		log.Fatalf("course format migration failed: %v", err)
	}
	if err := db.Model(&models.Course{}).Where("capacity IS NULL OR capacity = 0").Update("capacity", 10).Error; err != nil {
		log.Fatalf("course capacity migration failed: %v", err)
	}
	start := time.Now().AddDate(0, 1, 0).Truncate(24 * time.Hour)
	if err := db.Model(&models.Course{}).Where("start_date IS NULL OR start_date = ?", time.Time{}).Update("start_date", start).Error; err != nil {
		log.Fatalf("course date migration failed: %v", err)
	}
	if db.Name() == "postgres" {
		if err := db.Exec("UPDATE enrollments SET payment_due_at = updated_at + INTERVAL '3 days' WHERE status = 'approved' AND payment_due_at IS NULL").Error; err != nil {
			log.Fatalf("payment due migration failed: %v", err)
		}
		if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS one_active_enrollment_per_course ON enrollments (user_id, course_id) WHERE status IN ('pending', 'approved', 'paid')").Error; err != nil {
			log.Fatalf("active enrollment index failed: %v", err)
		}
	}

	log.Println("migrations applied successfully")
}

func BackfillCourseOwners(db *gorm.DB) {
	var teacher models.User
	if err := db.Where("email = ? AND role = ?", "teacher@chess.com", models.RoleTeacher).First(&teacher).Error; err != nil {
		log.Printf("legacy courses have no default teacher: %v", err)
		return
	}
	if err := db.Model(&models.Course{}).Where("author_id = 0 OR author_id IS NULL").Update("author_id", teacher.ID).Error; err != nil {
		log.Fatalf("course owner migration failed: %v", err)
	}
}
