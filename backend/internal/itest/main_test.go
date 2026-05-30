//go:build integration

package itest

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"backend/internal/database"
	"backend/internal/server"
	"backend/migrations"
)

// testDBService adapts a raw *sql.DB into the database.Service interface
// so we can hand it to server.NewServerWithDB without going through database.New().
type testDBService struct{ db *sql.DB }

func (s *testDBService) DB() *sql.DB             { return s.db }
func (s *testDBService) Close() error            { return s.db.Close() }
func (s *testDBService) Health() map[string]string {
	if err := s.db.Ping(); err != nil {
		return map[string]string{"status": "down", "error": err.Error()}
	}
	return map[string]string{"status": "up", "message": "ok"}
}

var _ database.Service = (*testDBService)(nil)

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		log.Fatalf("start postgres: %v", err)
	}
	defer func() { _ = container.Terminate(ctx) }()

	host, err := container.Host(ctx)
	if err != nil {
		log.Fatalf("container host: %v", err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		log.Fatalf("container port: %v", err)
	}

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())
	rawDB, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer rawDB.Close()

	if err := database.RunMigrations(rawDB, migrations.FS); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	uploadDir, err := os.MkdirTemp("", "itest-uploads-*")
	if err != nil {
		log.Fatalf("upload tmp: %v", err)
	}
	defer os.RemoveAll(uploadDir)

	apiServer := server.NewServerWithDB(0, &testDBService{db: rawDB}, "itest-secret", uploadDir)
	ts := httptest.NewServer(apiServer.Handler)
	defer ts.Close()

	db = rawDB
	srv = ts
	uploadRoot = uploadDir
	avatarsRoot = filepath.Join(uploadDir, "avatars")

	os.Exit(m.Run())
}
