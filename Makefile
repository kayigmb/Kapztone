COMPOSE := docker compose -f infra/docker-compose.yml

.PHONY: up down logs api web test

up: ## build and start db + api + web
	$(COMPOSE) up --build -d

down:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f

api: ## run the API natively on :8080
	cd backend && go run ./cmd/api

web: ## run the web app natively on :3000
	cd frontend && pnpm dev

test:
	cd backend && go vet ./... && go test ./...
	cd frontend && pnpm lint && pnpm build
	cd contracts && pnpm compile
