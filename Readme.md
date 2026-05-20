# My Go Chat

A small but production-grade WebSocket chat written in Go. Real-time messaging,
message history, profanity filtering, structured logging, embedded migrations and
graceful shutdown — designed to be a "clone-and-run" reference.

## Stack

- **Language**: Go 1.26
- **Database**: PostgreSQL 18 (pgx/v5)
- **Migrations**: Goose v3 (embedded, auto-run on startup)
- **Routing**: chi v5
- **Realtime**: coder/websocket
- **Logging**: slog (structured)
- **Deployment**: Docker multi-stage build + Compose

## Architecture

```
cmd/server/              # entry point — builds App and runs it
internal/
├── app/                 # App lifecycle + DI container (wiring, graceful shutdown)
├── config/              # env-driven configuration
├── api/
│   ├── chat/            # websocket handler (upgrade, read loop, history, welcome)
│   ├── health/          # GET /health
│   ├── root/            # serves the embedded chat UI
│   └── middleware/      # slog request logger
├── chat/                # domain logic: Hub (fan-out) + Service (persist/censor/broadcast)
├── censor/              # profanity filtering
├── model/               # Message domain type
├── repository/message/  # pgx data access (parameterized queries)
├── platform/dbtx/       # DB interface shared by pool and tx
├── migrations/          # embedded *.sql goose migrations
└── pkg/utils/           # color helper
web/                     # embedded index.html chat client
```

**Layering**: handler → service → repository. The HTTP/websocket layer never
touches the database directly; the read/write logic lives in `chat.Service`, and
SQL lives only in `repository/message`.

**Concurrency**: every websocket connection is owned by a single writer goroutine
(`Hub.writePump`). Broadcasts and the welcome banner enqueue onto a per-client
buffered channel rather than writing to the socket directly, so there is never
more than one concurrent writer per connection. Slow clients whose buffer fills
are dropped instead of blocking the broadcaster. Verified with `go test -race`.

## API

- `GET /` — chat client HTML
- `GET /health` — health/version JSON
- `GET /ws` — WebSocket endpoint

## Quick start

```bash
cp .env.dev.example .env
make dev          # builds + starts app and postgres
```

Open http://localhost:8011/ — migrations run automatically on first boot.

```bash
make logs         # tail app logs
make down         # stop everything
```

## Local development (without Docker)

Point `DB_DSN` at a running Postgres and run:

```bash
DB_DSN="postgres://chat:chat@localhost:5433/chat?sslmode=disable" make build && ./bin/mygochat
```

## Configuration

| Env | Default | Description |
|-----|---------|-------------|
| `APP_ENV` | `dev` | environment label |
| `APP_PORT` | `:8080` | listen address (inside container) |
| `APP_EXTERNAL_PORT` | `8011` | host port mapped to the app |
| `DB_DSN` | `postgres://chat:chat@db:5432/chat?sslmode=disable` | Postgres connection string |
| `DB_MAX_CONNS` / `DB_MIN_CONNS` | `10` / `2` | pool sizing |
| `WELCOME_MESSAGE` | `""` | banner sent to new clients (empty = none) |
| `WELCOME_TIMEOUT` | `0` | seconds to delay the banner |
| `PROFANITIES` | `""` | comma-separated extra censored words |
| `CHAT_HISTORY_LIMIT` | `50` | messages replayed on connect |
| `CORS_TRUSTED_ORIGINS` | `""` | allowed WS origins (empty = allow all; **set in prod**) |
| `APP_SHUTDOWN_TIMEOUT` | `10s` | graceful shutdown grace period |

## Migrations

Migrations live in `internal/migrations/*.sql`, are embedded into the binary, and
run automatically on startup via Goose. Add a new one:

```bash
# create internal/migrations/0000N_name.sql with -- +goose Up / Down sections
```

No external `migrate` CLI is required.

## Testing

```bash
make test         # go test -race ./...
```
