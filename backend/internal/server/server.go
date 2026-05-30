package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"backend/internal/database"
	"backend/internal/handler"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/migrations"
)

type Server struct {
	port        int
	db          database.Service
	jwtSecret   string
	uploadDir   string
	authHandler *handler.AuthHandler
	userHandler *handler.UserHandler
	gameHandler *handler.GameHandler
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	dbService := database.New()

	if err := database.RunMigrations(dbService.DB(), migrations.FS); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	return NewServerWithDB(port, dbService, jwtSecret, uploadDir)
}

// NewServerWithDB builds the HTTP server from pre-constructed dependencies.
// Intended for integration tests; production code uses NewServer().
func NewServerWithDB(port int, dbService database.Service, jwtSecret, uploadDir string) *http.Server {
	userRepo := repository.NewUserRepository(dbService.DB())
	fileRepo := repository.NewFileRepository(dbService.DB())
	gameRepo := repository.NewGameRepository(dbService.DB())
	relRepo := repository.NewRelationshipRepository(dbService.DB())

	authSvc := service.NewAuthService(userRepo, jwtSecret)
	userSvc := service.NewUserService(userRepo, fileRepo, relRepo, uploadDir)
	gameSvc := service.NewGameService(gameRepo)

	srv := &Server{
		port:        port,
		db:          dbService,
		jwtSecret:   jwtSecret,
		uploadDir:   uploadDir,
		authHandler: handler.NewAuthHandler(authSvc),
		userHandler: handler.NewUserHandler(userSvc),
		gameHandler: handler.NewGameHandler(gameSvc),
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", srv.port),
		Handler:      srv.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}
