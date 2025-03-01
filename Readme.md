# My Go Chat

This is Websocket super simple chat.
- MySQL
- GoLang
- Docker
- JS/HTML/CSS

## Route list
- HOST:PORT/ - index HTML client page
- HOST:PORT/ws - web socket listener

## How to install
- cp .env.example .env
- docker-compose up --build
- run migrations
- Profit!

## index HTML client page
- `http://localhost:8011/`

## WebSocket Usage
With this setup, your WebSocket server will be available at:
- `ws://localhost:8011/ws`

Your client JavaScript can automatically use the correct WebSocket protocol (ws:// or wss://) based on how the page was loaded (HTTP or HTTPS).

## Migrations
- wget https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz
- tar -zxvf migrate.linux-amd64.tar.gz
- sudo mv migrate.linux-amd64 /usr/local/bin/migrate
- sudo chmod +x /usr/local/bin/migrate
- migrate create -ext sql -dir migrations -seq init_schema
- `migrate -source "file://migrations" -database "mysql://root:password@tcp(localhost:3310)/chat" up`
- Profit!

### mig down
- `migrate -source "file://migrations" -database "mysql://root:password@tcp(localhost:3310)/chat" down`

### mig fix
- `migrate -source "file://migrations" -database "mysql://root:password@tcp(localhost:3310)/chat" force 4`
