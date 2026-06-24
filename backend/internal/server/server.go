package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/google/uuid"

	"backend/internal/database"
	"backend/internal/handler"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/ws"
	"backend/migrations"
)

type Server struct {
	port        int
	db          database.Service
	jwtSecret   string
	uploadDir   string
	onSeen      func(uuid.UUID)
	authHandler  *handler.AuthHandler
	userHandler  *handler.UserHandler
	gameHandler  *handler.GameHandler
	lobbyHandler *handler.LobbyHandler
	matchHandler *handler.MatchHandler
	wsHandler    *ws.Handler
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
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("failed to create upload directory %s: %v", uploadDir, err)
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
	lobbyRepo := repository.NewLobbyRepository(dbService.DB())
	achieveRepo := repository.NewAchievementsRepository(dbService.DB())

	hub := ws.NewHub()
	authSvc := service.NewAuthService(userRepo, jwtSecret)
	userSvc := service.NewUserService(userRepo, fileRepo, relRepo, achieveRepo, uploadDir)
	gameSvc := service.NewGameService(gameRepo)
	lobbySvc := service.NewLobbyService(lobbyRepo, gameRepo, hub)
	matchSvc := service.NewMatchService(gameRepo, lobbyRepo)

	onSeen := func(id uuid.UUID) {
		// Best-effort; errors don't surface to the request handler.
		_ = userRepo.UpdateLastSeen(context.Background(), id)
	}

	srv := &Server{
		port:        port,
		db:          dbService,
		jwtSecret:   jwtSecret,
		uploadDir:   uploadDir,
		onSeen:      onSeen,
		authHandler:  handler.NewAuthHandler(authSvc),
		userHandler:  handler.NewUserHandler(userSvc),
		gameHandler:  handler.NewGameHandler(gameSvc),
		lobbyHandler: handler.NewLobbyHandler(lobbySvc),
		matchHandler: handler.NewMatchHandler(matchSvc),
		wsHandler:    ws.NewHandler(hub, jwtSecret, lobbySvc, gameRepo, matchSvc),
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", srv.port),
		Handler:      srv.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}
