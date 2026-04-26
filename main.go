package main

import (
	"context"
	"log"
	"log/slog"
	"os/signal"
	"rankode/internal/api"
	"rankode/internal/config"
	"rankode/internal/middleware"
	db "rankode/internal/repository"
	"rankode/internal/services/assignments"
	"rankode/internal/services/attempts"
	"rankode/internal/services/auth"
	"rankode/internal/services/courses"
	"rankode/internal/services/files"
	"rankode/internal/services/grades"
	"rankode/internal/services/groups"
	"rankode/internal/services/lti"
	"rankode/internal/services/rabbit_runner"
	"rankode/internal/services/tasks"
	tasksvalidator "rankode/internal/services/tasks_validator"
	"rankode/internal/services/test_cases"
	"rankode/internal/services/users"
	"syscall"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
)


// @title           Rankode
// @version         1.0

// @host      localhost:4000
// @BasePath  /api

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @securitydefinitions.apikey.description JWT

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalln("failed to init config:", err)
	}

	pgPool, err := cfg.NewPgxPool(ctx)
	if err != nil {
		log.Fatalln("failed to init db:", err)
	}

	execer := db.New(pgPool)
	fileStorage := files.NewFileStorage(cfg)
	for _, bucket := range []string{"tasks", "verification"} {
		if err := fileStorage.EnsureBucket(ctx, bucket); err != nil {
			log.Fatalf("failed to initialize bucket %s: %v", bucket, err)
		}
	}
	execer.CreateTopic(ctx, "test")
	usersService := users.NewUserService(execer)
	authService := auth.NewAuthService(cfg)
	taskService := tasks.NewTaskService(pgPool, fileStorage)
	testCasesService := test_cases.NewTestCasesService(taskService, fileStorage, execer)
	coursesService := courses.NewCourseService(execer, pgPool)
	assignmentsService := assignments.NewAssignmentService(execer, pgPool)
	gradesService := grades.NewGradeService(execer, pgPool)
	groupsService := groups.NewGroupService(execer, pgPool)
	ltiService := lti.NewLtiService(cfg, execer, usersService)

	attemptsValidator := tasksvalidator.NewTasksValidator(execer, pgPool, fileStorage)
	runner, err := rabbitrunner.NewRabbitMQRunner(rabbitrunner.RabbitMQRunnerConfig{
		Login:        cfg.RabbitMQLogin,
		Password:     cfg.RabbitMQPassword,
		Host:         cfg.RabbitMQHost,
		Port:         cfg.RabbitMQPort,
		WorkersCount: 10,
	}, attemptsValidator)
	if err != nil {
		log.Fatalln("failed to init rabbit runner:", err)
	}
	if err := runner.Start(); err != nil {
		log.Fatalln("failed to start rabbit runner:", err)
	}

	attemptsService := attempts.NewAttemptsService(execer, pgPool, runner)

	app := fiber.New(fiber.Config{
		ServerHeader:             "supaserver-3000",
		StreamRequestBody:        true,
		BodyLimit:                20 * 1024 * 1024,
		EnableSplittingOnParsers: true,
	})
	app.Use(cors.New(cors.ConfigDefault))
	app.Use(compress.New())
	app.Use(logger.New())


	authMiddleware := middleware.NewAuthMiddleware(authService)
	apiGroup := app.Group("/api")
	api.NewAuthHandler(usersService, authService).RegisterRoutes(apiGroup)
	api.NewLtiHandler(cfg, ltiService, authService).RegisterRoutes(apiGroup)
	api.NewTasksHandler(taskService, testCasesService).RegisterRoutes(apiGroup, authMiddleware)
	api.NewtopicsHandler(taskService).RegisterRoutes(apiGroup, authMiddleware)
	api.NewTestCasesHandler(testCasesService).RegisterRoutes(apiGroup, authMiddleware)
	api.NewAttemptsHandler(attemptsService).RegisterRoutes(apiGroup, authMiddleware)
	api.NewLeaderboardHandler(usersService).RegisterRoutes(apiGroup)
	api.NewCoursesHandler(coursesService).RegisterRoutes(apiGroup, authMiddleware)
	api.NewAssignmentsHandler(assignmentsService).RegisterRoutes(apiGroup, authMiddleware)
	api.NewGradesHandler(gradesService).RegisterRoutes(apiGroup, authMiddleware)
	api.NewGroupsHandler(groupsService).RegisterRoutes(apiGroup, authMiddleware)
	slog.Info("HTTP server started", "listenAddr", cfg.ListenAddr)
	app.Listen(cfg.ListenAddr)
	<-ctx.Done()
	app.Shutdown()
	pgPool.Close()
}
