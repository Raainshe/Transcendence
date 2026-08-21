package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/server"
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop() // Allow Ctrl+C to force shutdown

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

// @title           Transcendence Public API
// @version         1.0
// @description     Public API for third-party integrations: Leaderboard, User Stats, and Lobby Management.
// @description
// @description     ## Getting started
// @description     1. Log in to the app and create a key under User > Profile > API Keys (or call `POST /api/v1/users/me/api-keys` with your JWT).
// @description     2. In this page, click Authorize and paste your key to try requests live. In your own client, send it as the X-API-Key header.
// @description
// @description     ## Rate limits
// @description     Unauthenticated requests are capped per IP at 10 req/s (burst 30). Authenticated requests are capped per key: 5 req/s (burst 20) on GET endpoints, 1 req/s (burst 5) on POST/PUT/DELETE endpoints. Exceeding a limit returns `429` with a `Retry-After` header.
// @description
// @description     ## Errors
// @description     Every error response shares the shape `{"error": "message"}`, with the HTTP status indicating the failure category (`400` invalid input, `401` missing/invalid key, `403` not the resource owner, `404` not found, `409` conflicting state, `429` rate limited).
// @host            localhost
// @schemes         https
// @BasePath        /api/public/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in              header
// @name            X-API-Key
func main() {

	server := server.NewServer()

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, done)

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
