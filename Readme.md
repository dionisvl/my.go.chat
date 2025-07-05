# My Go Chat

A simple WebSocket-based chat application built with Go, MySQL, and Docker.

## Features
- Real-time messaging via WebSocket
- Profanity filtering
- User color generation
- Message persistence with MySQL
- Dockerized deployment

## Tech Stack
- **Backend**: Go (Golang)
- **Database**: MySQL
- **Frontend**: HTML/CSS/JavaScript
- **Deployment**: Docker & Docker Compose

## API Endpoints
- `GET /` - Chat client HTML page
- `WebSocket /ws` - WebSocket connection for real-time messaging

## Quick Start

1. **Setup environment**:
   ```bash
   cp .env.example .env
   ```

2. **Start the application**:
   ```bash
   docker-compose up --build
   ```

3. **Run database migrations**:
   ```bash
   migrate -source "file://migrations" -database "mysql://root:password@tcp(localhost:3310)/chat" up
   ```

4. **Access the application**:
   - Chat interface: `http://localhost:8011/`
   - WebSocket endpoint: `ws://localhost:8011/ws`

## Database Migrations

### Install migrate tool
```bash
wget https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz
tar -zxvf migrate.linux-amd64.tar.gz
sudo mv migrate.linux-amd64 /usr/local/bin/migrate
sudo chmod +x /usr/local/bin/migrate
```

### Migration commands
```bash
# Create new migration
migrate create -ext sql -dir migrations -seq migration_name

# Run migrations
migrate -source "file://migrations" -database "mysql://root:password@tcp(localhost:3310)/chat" up

# Rollback migrations
migrate -source "file://migrations" -database "mysql://root:password@tcp(localhost:3310)/chat" down

# Fix migration version
migrate -source "file://migrations" -database "mysql://root:password@tcp(localhost:3310)/chat" force 4
```

## Testing
```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/chat
go test ./internal/database
go test ./internal/pkg/utils
```

## Project Structure
```
.
├── cmd/server/          # Application entry point
├── internal/
│   ├── chat/           # Chat logic and profanity filtering
│   ├── database/       # Database operations
│   ├── handler/        # HTTP and WebSocket handlers
│   └── pkg/utils/      # Utility functions
├── migrations/         # Database migrations
├── web/                # Static files
└── compose.yml         # Docker configuration
```
