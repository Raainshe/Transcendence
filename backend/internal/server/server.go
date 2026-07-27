package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"

	"backend/internal/database"
	"backend/internal/handler"
	"backend/internal/mailer"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/ws"
	"backend/migrations"
)

type Server struct {
	port                int
	db                  database.Service
	jwtSecret           string
	uploadDir           string
	onSeen              func(uuid.UUID)
	authHandler         *handler.AuthHandler
	userHandler         *handler.UserHandler
	gameHandler         *handler.GameHandler
	lobbyHandler        *handler.LobbyHandler
	matchHandler        *handler.MatchHandler
	wsHandler           *ws.Handler
	achievementsHandler *handler.AchievementsHandler
	chatHandler         *handler.ChatHandler
	apiHandler          *handler.APIKeyHandler
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
	twoFactorRepo := repository.NewTwoFactorRepository(dbService.DB())
	messageRepo := repository.NewMessageRepository(dbService.DB())
	achievementsRepo := repository.NewAchievementsRepository(dbService.DB())
	apiRepo := repository.NewAPIKeyRepository(dbService.DB())

	var mail mailer.Mailer
	if host := os.Getenv("SMTP_HOST"); host != "" {
		mail = mailer.NewSMTPMailer(
			host, os.Getenv("SMTP_PORT"),
			os.Getenv("SMTP_USERNAME"), os.Getenv("SMTP_PASSWORD"),
			os.Getenv("SMTP_FROM"),
		)
	} else {
		mail = mailer.LogMailer{}
	}

	hub := ws.NewHub()
	gamificationSvc := service.NewGamificationService(achievementsRepo, gameRepo, userRepo)
	chatSvc := service.NewChatService(messageRepo, relRepo)
	authSvc := service.NewAuthService(userRepo, twoFactorRepo, mail, jwtSecret)
	userSvc := service.NewUserService(userRepo, fileRepo, relRepo, gamificationSvc, uploadDir)
	gameSvc := service.NewGameService(gameRepo, gamificationSvc)
	lobbySvc := service.NewLobbyService(lobbyRepo, gameRepo, hub, gamificationSvc)
	matchSvc := service.NewMatchService(gameRepo, lobbyRepo, gamificationSvc)
	apiSvc := service.NewAPIKeyService(apiRepo)

	onSeen := func(id uuid.UUID) {
		// Best-effort; errors don't surface to the request handler.
		_ = userRepo.UpdateLastSeen(context.Background(), id)
	}

	srv := &Server{
		port:                port,
		db:                  dbService,
		jwtSecret:           jwtSecret,
		uploadDir:           uploadDir,
		onSeen:              onSeen,
		authHandler:         handler.NewAuthHandler(authSvc),
		userHandler:         handler.NewUserHandler(userSvc),
		gameHandler:         handler.NewGameHandler(gameSvc),
		lobbyHandler:        handler.NewLobbyHandler(lobbySvc),
		matchHandler:        handler.NewMatchHandler(matchSvc),
		wsHandler:           ws.NewHandler(hub, jwtSecret, lobbySvc, gameRepo, matchSvc, chatSvc),
		achievementsHandler: handler.NewAchievementsHandler(gamificationSvc),
		chatHandler:         handler.NewChatHandler(chatSvc),
		apiHandler:          handler.NewAPIKeyHandler(apiSvc),
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", srv.port),
		Handler:      srv.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}
