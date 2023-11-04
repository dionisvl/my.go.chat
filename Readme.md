# My Go Chat

This is Websocket super simple chat.
- MySQL
- GoLang
- Docker
- JS/HTML/CSS
- microservice edition

## Route list
- HOST:PORT/ - index page
- HOST:PORT/ws - web socket listener

## How to install
- cp .env.example .env
- docker-compose up --build
- run migrations
- Profit!

## Migrations up
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
