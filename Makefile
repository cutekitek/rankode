export VITE_API_BASE_PATH=/api

.PHONY: sql build run frontend docs

run: sql build 
	@./bin/app

prod: frontend build

frontend:
	cd frontend && yarn build

build:
	@go build -o bin/app .

sql:
	cd internal/repository && sqlc generate

docs:
	swag init --pd