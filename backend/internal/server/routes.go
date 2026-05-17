package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimid "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	mw "backend/internal/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(chimid.Logger)
	r.Use(chimid.Recover)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/", s.HelloWorldHandler)
	r.Get("/health", s.healthHandler)

	r.Route("/api/v1", func(r chi.Router) {
		// Auth — public
		r.Post("/auth/register", s.authHandler.Register)
		r.Post("/auth/login", s.authHandler.Login)
		r.Post("/auth/refresh", s.authHandler.Refresh)
		r.Get("/auth/oauth/{provider}", s.authHandler.OAuthRedirect)
		r.Get("/auth/oauth/{provider}/callback", s.authHandler.OAuthCallback)

		// Auth — protected
		r.Group(func(r chi.Router) {
			r.Use(mw.JWTAuth(s.jwtSecret))
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
			r.Use(mw.JWTAuth(s.jwtSecret))
			r.Get("/users/me", s.userHandler.GetMe)
			r.Patch("/users/me", s.userHandler.UpdateMe)
			r.Post("/users/me/avatar", s.userHandler.UploadAvatar)
			r.Get("/users/me/friends", s.userHandler.GetFriends)
			r.Post("/users/me/friends/{id}", s.userHandler.AddFriend)
			r.Delete("/users/me/friends/{id}", s.userHandler.RemoveFriend)
			r.Post("/users/me/block/{id}", s.userHandler.BlockUser)
			r.Delete("/users/me/block/{id}", s.userHandler.UnblockUser)
		})

		// Games
		r.Get("/games", s.gameHandler.ListGames)
		r.Get("/games/{id}", s.gameHandler.GetGame)
		r.Get("/leaderboard", s.gameHandler.GetLeaderboard)
		r.Group(func(r chi.Router) {
			r.Use(mw.JWTAuth(s.jwtSecret))
			r.Post("/games", s.gameHandler.CreateGame)
		})
	})

	return r
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.db.Health())
	_, _ = w.Write(jsonResp)
}
