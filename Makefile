ENV_FILE := $(if $(wildcard .env),--env-file .env)
COMPOSE  := docker compose $(ENV_FILE)
INFRA    := -f infra/compose.infra.yml
# dev and prod are separate compose projects, so each keeps its own database volume.
DEV      := $(COMPOSE) -p chaintrack $(INFRA) -f infra/compose.dev.yml
PROD     := $(COMPOSE) -p chaintrack-prod $(INFRA) -f infra/compose.prod.yml

.PHONY: infra dev prod down prod-down logs test

infra: ## database + migrations only
	$(DEV) up -d db migrate

dev: ## database + API + web, reloading on save
	$(DEV) up -d

prod: ## built images (needs POSTGRES_PASSWORD)
	$(PROD) up --build -d

down:
	$(DEV) down

prod-down:
	$(PROD) down

logs:
	$(DEV) logs -f

test:
	cd backend && go vet ./... && go test ./...
	cd frontend && pnpm lint && pnpm build
	cd contracts && pnpm compile
