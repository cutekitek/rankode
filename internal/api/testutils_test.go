package api

import (
	"context"
	"rankode/internal/config"
	"rankode/internal/middleware"
	"rankode/internal/models"
	db "rankode/internal/repository"
	"rankode/internal/services/assignments"
	"rankode/internal/services/attempts"
	"rankode/internal/services/auth"
	"rankode/internal/services/courses"
	"rankode/internal/services/files"
	"rankode/internal/services/grades"
	"rankode/internal/services/groups"
	"rankode/internal/services/tasks"
	"rankode/internal/services/test_cases"
	"rankode/internal/services/users"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

type mockQueue struct{}

func (m *mockQueue) SendAttempt(req models.AttemptRequest) error {
	return nil
}

type testApp struct {
	App              *fiber.App
	DB               *pgxpool.Pool
	Queries          *db.Queries
	AuthService      *auth.AuthService
	UsersService     *users.UserService
	TaskService      *tasks.TaskService
	TestCasesService *test_cases.TestCasesService
}

func setupTestApp(ctx context.Context) (*testApp, error) {
	cfg := &config.Config{
		PostgresString: "postgres://rankode:fobeagTB8Ojo3R@localhost:5432/rankode?sslmode=disable",
		JWTSecret:      "test_secret",
		S3Endpoint:     "localhost:8333",
		S3AccessKey:    "tasks",
		S3SecretKey:    "fobeagTB8Ojo3R",
	}

	pgPool, err := pgxpool.New(ctx, cfg.PostgresString)
	if err != nil {
		return nil, err
	}

	// Seed languages
	pgPool.Exec(ctx, "INSERT INTO languages (name) VALUES ('python'), ('go'), ('cpp'), ('java') ON CONFLICT DO NOTHING")
	// Seed a topic
	pgPool.Exec(ctx, "INSERT INTO topics (name) VALUES ('General') ON CONFLICT DO NOTHING")
	pgPool.Exec(ctx, "ALTER TABLE tasks ADD COLUMN IF NOT EXISTS verification_file TEXT")

	execer := db.New(pgPool)
	fileStorage := files.NewFileStorage(cfg)
	usersService := users.NewUserService(execer)
	authService := auth.NewAuthService(cfg)
	taskService := tasks.NewTaskService(pgPool, fileStorage)
	testCasesService := test_cases.NewTestCasesService(taskService, fileStorage, execer)
	coursesService := courses.NewCourseService(execer, pgPool)
	assignmentsService := assignments.NewAssignmentService(execer, pgPool)
	gradesService := grades.NewGradeService(execer, pgPool)
	groupsService := groups.NewGroupService(execer, pgPool)
	attemptsService := attempts.NewAttemptsService(execer, pgPool, &mockQueue{})

	app := fiber.New()
	authMiddleware := middleware.NewAuthMiddleware(authService)
	apiGroup := app.Group("/api")

	NewAuthHandler(usersService, authService).RegisterRoutes(apiGroup)
	NewTasksHandler(taskService, testCasesService).RegisterRoutes(apiGroup, authMiddleware)
	NewtopicsHandler(taskService).RegisterRoutes(apiGroup, authMiddleware)
	NewTestCasesHandler(testCasesService).RegisterRoutes(apiGroup, authMiddleware)
	NewAttemptsHandler(attemptsService).RegisterRoutes(apiGroup, authMiddleware)
	NewLeaderboardHandler(usersService).RegisterRoutes(apiGroup)
	NewCoursesHandler(coursesService).RegisterRoutes(apiGroup, authMiddleware)
	NewAssignmentsHandler(assignmentsService).RegisterRoutes(apiGroup, authMiddleware)
	NewGradesHandler(gradesService).RegisterRoutes(apiGroup, authMiddleware)
	NewGroupsHandler(groupsService).RegisterRoutes(apiGroup, authMiddleware)

	return &testApp{
		App:              app,
		DB:               pgPool,
		Queries:          execer,
		AuthService:      authService,
		UsersService:     usersService,
		TaskService:      taskService,
		TestCasesService: testCasesService,
	}, nil
}

func testEndpoint(path string) string {
	return "http://127.0.0.1" + path
}
