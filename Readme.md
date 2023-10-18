# My Go Chat

Это Websocket чат совсем простой. 
- MySQL
- GoLang 
- Docker 
- JS/HTML/CSS


## How to install
- cp .env.example .env
- docker-compose up --build
- run migrations
- Profit!


## migrations in go
- wget https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz
- tar -zxvf migrate.linux-amd64.tar.gz
- mv migrate.linux-amd64 /usr/local/bin/migrate
- sudo chmod +x /usr/local/bin/migrate
- migrate create -ext sql -dir migrations -seq init_schema
- `migrate -source "file://migrations" -database "mysql://root:password@tcp(localhost:3310)/chat" up`
- Profit!