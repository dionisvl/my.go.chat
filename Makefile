include .env

up: docker-up
down: docker-down

r:
	docker compose up --build -d
r-app:
	docker-compose up --no-deps --build app

docker-up:
	docker compose up -d

docker-down:
	docker compose down --remove-orphans

#migrate:
#	docker compose run -w $(PROJECT_DIR) --rm app