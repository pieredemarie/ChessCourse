package storage

import (
	"chesscourse/internal/models"
	"errors"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedAdmin(db *gorm.DB) {
	if os.Getenv("APP_ENV") != "development" {
		return
	}

	adminEmail := "admin@chesscourse.com"

	var existing models.User
	result := db.Where("email = ?", adminEmail).First(&existing)
	if result.Error == nil {
		log.Println("admin user already exists, skipping")
		return
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Fatalf("failed to check admin: %v", result.Error)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("Admin123!"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("failed to hash admin password")
	}

	admin := &models.User{
		Name:         "Admin",
		Email:        adminEmail,
		PasswordHash: string(hash),
		Role:         models.RoleAdmin,
	}

	if err := db.Create(admin).Error; err != nil {
		log.Fatalf("failed to seed admin: %v", err)
	}

	log.Println("admin user seeded")
}

func SeedCategories(db *gorm.DB) {
	if os.Getenv("APP_ENV") != "development" {
		log.Println("skipping categories seed (not development environment)")
		return
	}

	categories := []models.Category{
		{Name: "Дебюты", Description: "Начало партии: принципы и популярные дебюты"},
		{Name: "Миттельшпиль", Description: "Середина партии: тактика и стратегия"},
		{Name: "Эндшпиль", Description: "Конец партии: пешечные и ладейные окончания"},
		{Name: "Тактика", Description: "Вилки, связки, двойные удары, матовые атаки"},
		{Name: "Стратегия", Description: "Позиционная игра, слабые поля, планы"},
	}

	created := 0
	for _, cat := range categories {
		var existing models.Category
		result := db.Where("name = ?", cat.Name).First(&existing)

		if result.Error == nil {
			log.Printf("category '%s' already exists, skipping", cat.Name)
			continue
		}

		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Fatalf("failed to check category '%s': %v", cat.Name, result.Error)
		}

		if err := db.Create(&cat).Error; err != nil {
			log.Fatalf("failed to seed category '%s': %v", cat.Name, err)
		}
		created++
	}

	log.Printf("categories seeded: %d created, %d skipped", created, len(categories)-created)
}

func SeedCourses(db *gorm.DB) {
	if os.Getenv("APP_ENV") != "development" {
		return
	}
	// Получаем категории
	var categories []models.Category
	db.Find(&categories)

	if len(categories) == 0 {
		log.Println("No categories found, skipping courses seed")
		return
	}

	courses := []models.Course{
		{
			Title:       "Шахматная тактика для начинающих",
			Description: "Изучите основные тактические приемы: вилка, связка, двойной удар и другие",
			Level:       models.LevelBeginner,
			Price:       29.99,
			CoverURL:    "https://i.1.creatium.io/disk2/4b/75/9d/6fe99842f6ac1fc3a2f4feb6bbd171053c/01ad4fcc83688913d8c7f01ff70996e6.jpg",
			Duration:    10,
			Lessons:     15,
			CategoryID:  categories[0].ID,
		},
		{
			Title:       "Стратегия миттельшпиля",
			Description: "Планирование атаки, оценка позиции, слабые поля",
			Level:       models.LevelIntermediate,
			Price:       49.99,
			CoverURL:    "https://i.1.creatium.io/disk2/4b/75/9d/6fe99842f6ac1fc3a2f4feb6bbd171053c/01ad4fcc83688913d8c7f01ff70996e6.jpg",
			Duration:    15,
			Lessons:     20,
			CategoryID:  categories[1].ID,
		},
		{
			Title:       "Эндшпиль: путь к мастерству",
			Description: "Пешечные, ладейные и ферзевые окончания",
			Level:       models.LevelAdvanced,
			Price:       59.99,
			CoverURL:    "https://i.1.creatium.io/disk2/4b/75/9d/6fe99842f6ac1fc3a2f4feb6bbd171053c/01ad4fcc83688913d8c7f01ff70996e6.jpg",
			Duration:    12,
			Lessons:     18,
			CategoryID:  categories[2].ID,
		},
		{
			Title:       "Дебютная подготовка",
			Description: "Изучите основные дебюты: Испанская партия, Сицилианская защита",
			Level:       models.LevelIntermediate,
			Price:       39.99,
			CoverURL:    "https://i.1.creatium.io/disk2/4b/75/9d/6fe99842f6ac1fc3a2f4feb6bbd171053c/01ad4fcc83688913d8c7f01ff70996e6.jpg",
			Duration:    8,
			Lessons:     12,
			CategoryID:  categories[0].ID,
		},
		{
			Title:       "Атакующие комбинации",
			Description: "Как найти и провести красивую атаку на короля",
			Level:       models.LevelAdvanced,
			Price:       69.99,
			CoverURL:    "https://i.1.creatium.io/disk2/4b/75/9d/6fe99842f6ac1fc3a2f4feb6bbd171053c/01ad4fcc83688913d8c7f01ff70996e6.jpg",
			Duration:    14,
			Lessons:     22,
			CategoryID:  categories[3].ID,
		},
	}

	created := 0
	for _, course := range courses {
		var existing models.Course
		result := db.Where("title = ?", course.Title).First(&existing)
		if result.Error == nil {
			continue
		}
		if err := db.Create(&course).Error; err != nil {
			log.Printf("Failed to create course '%s': %v", course.Title, err)
			continue
		}
		created++
	}

	log.Printf("courses seeded: %d created", created)
}

func SeedModulesAndLessons(db *gorm.DB) {
	if os.Getenv("APP_ENV") != "development" {
		return
	}

	// Получаем курсы
	var courses []models.Course
	db.Find(&courses)

	if len(courses) == 0 {
		return
	}

	for _, course := range courses {
		var existingModule models.Module
		result := db.Where("course_id = ? AND title = ?", course.ID, "Введение").First(&existingModule)
		if result.Error == nil {
			continue
		}

		// Модуль 1
		module1 := models.Module{
			CourseID:    course.ID,
			Title:       "Введение в курс",
			Description: "Основные понятия и первые шаги",
			SortOrder:   1,
		}
		db.Create(&module1)

		lessons1 := []models.Lesson{
			{ModuleID: module1.ID, Title: "Введение", Description: "Что вы узнаете из этого курса", SortOrder: 1, Duration: 10},
			{ModuleID: module1.ID, Title: "Установка и настройка", Description: "Подготовка к обучению", SortOrder: 2, Duration: 15},
		}
		for _, l := range lessons1 {
			db.Create(&l)
		}

		// Модуль 2
		module2 := models.Module{
			CourseID:    course.ID,
			Title:       "Основы",
			Description: "Базовые знания и навыки",
			SortOrder:   2,
		}
		db.Create(&module2)

		lessons2 := []models.Lesson{
			{ModuleID: module2.ID, Title: "Урок 1: Базовые понятия", Description: "Изучаем основы", SortOrder: 1, Duration: 20},
			{ModuleID: module2.ID, Title: "Урок 2: Практика", Description: "Закрепляем материал", SortOrder: 2, Duration: 25},
		}
		for _, l := range lessons2 {
			db.Create(&l)
		}
	}

	log.Println("modules and lessons seeded")
}

func SeedTeacher(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Where("email = ?", "teacher@chess.com").Count(&count)
	if count == 0 {
		// Пароль должен быть захеширован так же, как в SeedAdmin
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("teacher123"), bcrypt.DefaultCost)
		teacher := models.User{
			Name:         "ChessMaster",
			Email:        "teacher@chess.com",
			PasswordHash: string(hashedPassword),
			Role:         "Teacher",
		}
		db.Create(&teacher)
	}
}
