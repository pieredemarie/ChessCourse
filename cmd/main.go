package main

import (
	"chesscourse/internal/handlers"
	authservice "chesscourse/internal/service/auth"
	"chesscourse/internal/service/category"
	"chesscourse/internal/service/course"
	"chesscourse/internal/service/enrollment"
	"chesscourse/internal/service/learning"
	"chesscourse/internal/service/user"
	"chesscourse/internal/storage"

	_ "chesscourse/docs"

	"github.com/gin-contrib/cors" // добавьте этот импорт
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Chess Course API
// @version         1.0
// @description     API для платформы шахматных курсов

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен в формате: Bearer {token}

func main() {
	db := storage.NewPostgresDB()

	storage.RunMigrations(db)
	storage.SeedAdmin(db)
	storage.SeedCategories(db)
	storage.SeedCourses(db)
	storage.SeedModulesAndLessons(db)
	storage.SeedTeacher(db)

	store := storage.NewStorage(db)
	authSvc := authservice.NewAuthService(store)
	authHandler := handlers.NewAuthHandler(authSvc)

	categoryService := category.NewCategoryService(store)
	categoryHandler := handlers.NewCategoryHandler(categoryService)

	userService := user.NewUserService(store)
	userHandler := handlers.NewUserHandler(userService)

	courseService := course.NewCourseService(store)
	courseHandler := handlers.NewCourseHandler(courseService)
	learningService := learning.NewLearningService(store)
	enrollmentService := enrollment.NewEnrollmentService(store)
	enrollmentHandler := handlers.NewEnrollmentHandler(enrollmentService, learningService)

	learningHandler := handlers.NewLearningHandler(learningService)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) { c.Status(200) })
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")

	// Публичные auth роуты (регистрация и логин)
	authRoutes := api.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
	}

	// Все остальные API требуют авторизации
	authorized := api.Group("/")
	authorized.Use(authservice.AuthMiddleware())
	{
		// Auth
		authorized.GET("/auth/me", authHandler.Me)

		authorized.GET("/enrollments", enrollmentHandler.GetMyEnrollments)
		authorized.POST("/enrollments", enrollmentHandler.CreateEnrollment)
		authorized.POST("/enrollments/:id/pay", enrollmentHandler.PayForEnrollment)
		// Courses - доступны всем авторизованным (и студентам, и админам)
		authorized.GET("/courses", courseHandler.GetCourses)
		authorized.GET("/courses/:id", courseHandler.GetCourseByID)

		authorized.GET("/categories", categoryHandler.GetCategories)
		authorized.GET("/categories/:id", categoryHandler.GetCategoryByID)

		authorized.GET("/courses/:id/modules", learningHandler.GetCourseModules)
		authorized.GET("/courses/:id/progress", learningHandler.GetCourseProgress)
		authorized.PUT("/lessons/:id/progress", learningHandler.UpdateLessonProgress)

		authorized.GET("/my-courses", enrollmentHandler.GetMyCourses)

		// Admin routes - только для админов
		adminRoutes := authorized.Group("/admin")
		adminRoutes.Use(authservice.RequireAdminMiddleware())
		{
			// Categories management
			adminRoutes.GET("/categories", categoryHandler.GetCategories)
			adminRoutes.GET("/categories/:id", categoryHandler.GetCategoryByID)
			adminRoutes.POST("/categories", categoryHandler.CreateCategory)
			adminRoutes.PUT("/categories/:id", categoryHandler.UpdateCategory)
			adminRoutes.DELETE("/categories/:id", categoryHandler.DeleteCategory)

			// Users management
			adminRoutes.GET("/users", userHandler.GetUsers)
			adminRoutes.PUT("/users/:id/role", userHandler.UpdateUserRole)

			adminRoutes.POST("/courses", courseHandler.CreateCourse)
			adminRoutes.PUT("/courses/:id", courseHandler.UpdateCourse)
			adminRoutes.DELETE("/courses/:id", courseHandler.DeleteCourse)

		}
		teacherRoutes := authorized.Group("/teacher")
		teacherRoutes.Use(authservice.RequireTeacher())
		{
			teacherRoutes.GET("/enrollments", enrollmentHandler.GetAllEnrollments)
			teacherRoutes.PUT("/enrollments/:id", enrollmentHandler.UpdateEnrollmentStatus)
		}
	}

	r.Run(":8080")
}
