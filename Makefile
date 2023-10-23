include .env

up: docker-up
down: docker-down

r:
	docker compose up --build -d
r-app:
	docker-compose up --no-deps --build app
r-db:
	docker-compose up --no-deps --build mysql-chat

docker-up:
	docker compose up -d

docker-down:
	docker compose down --remove-orphans

sh:
	docker exec -it mygochat-app-1 sh

#migrate:
#	docker compose run -w $(PROJECT_DIR) --rm app