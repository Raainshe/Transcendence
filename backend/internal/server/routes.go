package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	chimid "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"backend/internal/handler"
	mw "backend/internal/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(chimid.Logger)
	r.Use(chimid.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/", s.HelloWorldHandler)
	r.Get("/health", s.healthHandler)

	// Serve uploaded files (avatars, etc.) — no directory listings.
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", noDirListing(s.uploadDir, http.FileServer(http.Dir(s.uploadDir)))))

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(mw.MaxBodyBytes(1 << 20)) // 1 MB cap for JSON bodies

		// Auth — public
		r.Post("/auth/register", s.authHandler.Register)
		r.Post("/auth/login", s.authHandler.Login)
		r.Post("/auth/refresh", s.authHandler.Refresh)
		r.Post("/auth/2fa/verify-login", s.authHandler.VerifyLogin)
		r.Get("/auth/oauth/{provider}", s.authHandler.OAuthRedirect)
		r.Get("/auth/oauth/{provider}/callback", s.authHandler.OAuthCallback)

		r.Group(func(r chi.Router) {
			r.Use(mw.JWTAuth(s.jwtSecret, s.onSeen))
			r.Get("/chat/messages/{friendID}", s.chatHandler.GetConversation)
			r.Post("/chat/messages/{friendID}/read", s.chatHandler.MarkRead)
			r.Get("/chat/unread", s.chatHandler.GetUnread)
		})

		// Auth — protected
		r.Group(func(r chi.Router) {
			r.Use(mw.JWTAuth(s.jwtSecret, s.onSeen))
			r.Post("/auth/logout", s.authHandler.Logout)
			r.Post("/auth/2fa/setup", s.authHandler.Setup2FA)
			r.Post("/auth/2fa/verify", s.authHandler.Verify2FA)
			r.Delete("/auth/2fa", s.authHandler.Disable2FA)
		})

		// Users — public
		r.Get("/users", s.userHandler.ListUsers)
		r.Get("/users/{id}/stats", s.gameHandler.GetUserStats)
		r.Get("/users/{id}", s.userHandler.GetUser)

		// Users — protected
		r.Group(func(r chi.Router) {
			r.Use(mw.JWTAuth(s.jwtSecret, s.onSeen))
			r.Get("/users/me", s.userHandler.GetMe)
			r.Patch("/users/me", s.userHandler.UpdateMe)
			r.Delete("/users/me", s.userHandler.DeleteMe)
			r.Post("/users/me/avatar", s.userHandler.UploadAvatar)
			r.Delete("/users/me/avatar", s.userHandler.DeleteAvatar)
			r.Get("/users/me/friends", s.userHandler.GetFriends)
			r.Get("/users/me/friends/requests", s.userHandler.GetPendingRequests)
			r.Post("/users/me/friends/{id}", s.userHandler.AddFriend)
			r.Patch("/users/me/friends/{id}", s.userHandler.AcceptFriend)
			r.Delete("/users/me/friends/{id}", s.userHandler.RemoveFriend)
			r.Get("/users/me/blocks", s.userHandler.GetBlockedUsers)
			r.Post("/users/me/block/{id}", s.userHandler.BlockUser)
			r.Delete("/users/me/block/{id}", s.userHandler.UnblockUser)
		})

		// Games
		r.Get("/games", s.gameHandler.ListGames)
		r.Get("/games/{id}", s.gameHandler.GetGame)
		r.Get("/leaderboard", s.gameHandler.GetLeaderboard)
		r.Group(func(r chi.Router) {
			r.Use(mw.JWTAuth(s.jwtSecret, s.onSeen))
			r.Post("/games", s.gameHandler.CreateGame)
		})

		// WebSocket — JWT via ?token= query param
		r.Get("/ws", s.wsHandler.ServeHTTP)

		// Matches — protected
		r.Group(func(r chi.Router) {
			r.Use(mw.JWTAuth(s.jwtSecret, s.onSeen))
			r.Get("/matches/{id}", s.matchHandler.Get)
		})

		// Lobbies — protected
		r.Group(func(r chi.Router) {
			r.Use(mw.JWTAuth(s.jwtSecret, s.onSeen))
			r.Post("/lobbies", s.lobbyHandler.Create)
			r.Post("/lobbies/join", s.lobbyHandler.JoinByCode)
			r.Get("/lobbies/by-code/{code}", s.lobbyHandler.GetByCode)
			r.Get("/lobbies/{id}", s.lobbyHandler.Get)
			r.Post("/lobbies/{id}/join", s.lobbyHandler.Join)
			r.Delete("/lobbies/{id}/leave", s.lobbyHandler.Leave)
			r.Post("/lobbies/{id}/ready", s.lobbyHandler.SetReady)
			r.Post("/lobbies/{id}/start", s.lobbyHandler.Start)
		})
	})

	return r
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	handler.WriteJSON(w, http.StatusOK, map[string]string{"message": "Hello World"})
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	stats := s.db.Health()
	code := http.StatusOK
	if stats["status"] != "up" {
		code = http.StatusServiceUnavailable
	}
	handler.WriteJSON(w, code, stats)
}

// noDirListing serves files but rejects requests targeting a directory,
// preventing http.FileServer's default auto-generated directory index.
func noDirListing(root string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path
		if urlPath == "" || strings.HasSuffix(urlPath, "/") {
			http.NotFound(w, r)
			return
		}
		full := filepath.Join(root, filepath.Clean("/"+urlPath))
		info, err := os.Stat(full)
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
