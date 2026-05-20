COMPOSE := docker compose
GOOSE_TAGS := no_sqlite3 no_clickhouse no_mssql no_mysql no_vertica no_ydb no_libsql no_duckdb

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

dev:
	$(COMPOSE) up --build -d

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down --remove-orphans

logs:
	$(COMPOSE) logs -f app

sh:
	$(COMPOSE) exec app sh

build:
	go build -tags="$(GOOSE_TAGS)" -o bin/mygochat ./cmd/server

test:
	go test -race ./...

lint:
	gofmt -w .
	go vet ./...

tidy:
	go mod tidy
