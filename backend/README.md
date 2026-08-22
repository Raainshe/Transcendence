# Backend

Go HTTP API for Transcendence: JWT auth, users, friends, lobbies, matches, chat, uploads, gamification, and a public API. Chi serves REST; a WebSocket hub handles lobby, match, and chat rooms. Goose applies SQL migrations on startup.

The evaluated stack is the repo root (`make up` → **https://localhost:8443**). This README covers running and testing the API on its own.

## Stack

- Go, Chi
- PostgreSQL 17
- JWT sessions, Goose migrations
- Swagger / OpenAPI (`/swagger/`)
- SMTP / SendGrid for optional email 2FA

## Environment

Copy [`../.env.example`](../.env.example) to `.env` at the **repo root** (Compose reads it from there). Important variables:

| Variable | Purpose |
| --- | --- |
| `PORT` | Listen port (default `8080` inside Docker) |
| `BLUEPRINT_DB_*` | Postgres host, port, database, user, password, schema |
| `JWT_SECRET` | Signing key for access tokens |
| `UPLOAD_DIR` | Avatar files |
| `SMTP_*` | 2FA mail (optional; leave unset to skip sending) |

## Makefile

Run these from `backend/`.

```bash
make build        # compile cmd/api/main.go → ./main
make run          # go run the API (needs a reachable Postgres)
make test         # go test ./...
make itest        # integration tests (tag integration, 120s timeout)
make swagger      # regenerate docs/ from swag annotations
make watch        # live reload via air
make docker-run   # Compose: API + Postgres
make docker-down
make clean        # remove ./main
make all          # build + test
```

`make swagger` needs [`swag`](https://github.com/swaggo/swag): `go install github.com/swaggo/swag/cmd/swag@latest`. `make watch` needs [`air`](https://github.com/air-verse/air).

Public API docs (when the stack is up): [https://localhost:8443/swagger/index.html](https://localhost:8443/swagger/index.html).

## Layout

| Path | Role |
| --- | --- |
| `cmd/api/` | Process entry, graceful shutdown |
| `internal/server/` | HTTP server, Chi routes |
| `internal/handler/` | HTTP handlers |
| `internal/service/` | Business logic (auth, lobby, match, chat, …) |
| `internal/repository/` | Postgres access |
| `internal/model/` | Domain types |
| `internal/ws/` | WebSocket hub and match/lobby/chat envelopes |
| `internal/middleware/` | JWT, API keys, rate limits, body size |
| `internal/database/` | Pool + Goose `Up` |
| `internal/mailer/` | SMTP |
| `internal/itest/` | Integration tests |
| `migrations/` | Goose SQL (source of truth for the schema) |
| `docs/` | Generated Swagger |

Handler → service → repository. Migrations live in `migrations/` and run automatically when the process starts.
