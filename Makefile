# All configuration comes from one file. Copy .env.example to .env first.
ENV_FILE ?= .env
COMPOSE  := docker compose --env-file $(ENV_FILE)
INFRA    := -f infra/compose.infra.yml
# dev and prod are separate compose projects, so each keeps its own database volume.
DEV      := $(COMPOSE) -p chaintrack-dev $(INFRA) -f infra/compose.dev.yml
PROD     := $(COMPOSE) -p chaintrack-prod $(INFRA) -f infra/compose.prod.yml
# Puts every value from the env file into the environment of the commands that follow.
LOAD_ENV := set -a && . $(abspath $(ENV_FILE)) && set +a

.PHONY: infra api relayer chain deploy bindings dev prod down prod-down logs test test-db

# Targets that cannot run without the env file.
infra api relayer deploy dev prod down prod-down logs test-db: $(ENV_FILE)

$(ENV_FILE):
	@echo "$(ENV_FILE) not found: copy .env.example to $(ENV_FILE) and fill in every value" && exit 1

infra: ## database + migrations + redis
	$(DEV) up -d --wait db redis
	$(DEV) run --rm migrate

api: ## run the API on this machine (needs `make infra`)
	$(LOAD_ENV) && cd backend && go run ./cmd/api

relayer: ## carry queued changes to the chain (needs `make infra`, a chain, deployed contracts)
	$(LOAD_ENV) && cd backend && go run ./cmd/relayer

chain: ## a local chain on http://127.0.0.1:8545 (chain id 31337)
	cd contracts && pnpm hardhat node

deploy: ## deploy the contracts to the chain in the env file; prints the address lines
	$(LOAD_ENV) && cd contracts && pnpm hardhat run scripts/deploy.ts --network target

dev: ## everything in containers, reloading on save
	$(DEV) up -d

prod: ## built images
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
	cd contracts && pnpm test

# Backend tests against the real Postgres and Redis from `make infra`, with
# coverage of the hand-written internal packages (generated contract bindings
# are left out). Redis database 1 keeps test keys apart from the API's database.
test-db: infra
	$(LOAD_ENV) && cd backend && \
	TEST_DATABASE_URL="$$DATABASE_URL" TEST_REDIS_URL="$${REDIS_URL%/*}/1" \
	go test ./tests/ -count=1 -coverpkg="$$(go list ./internal/... | grep -v /bindings | paste -sd, -)" -coverprofile=coverage.out && \
	go tool cover -func=coverage.out | tail -1

# Go bindings for the contracts, generated from the Hardhat build. Run after
# changing any contract; the generated files are committed.
CONTRACTS := OrgRegistry FacilityRegistry PassportRegistry CustodyManager AttestationRegistry
ABI_TMP   := backend/tmp/abi
# abigen is run at the same go-ethereum version the backend is built with.
GETH_VERSION = $(shell cd backend && go list -m -f '{{.Version}}' github.com/ethereum/go-ethereum)
bindings:
	cd contracts && pnpm hardhat compile
	@mkdir -p $(ABI_TMP) backend/internal/chain/bindings
	@for c in $(CONTRACTS); do \
		node -e "const a=require('./contracts/artifacts/contracts/$$c.sol/$$c.json'),fs=require('fs');fs.writeFileSync('$(ABI_TMP)/$$c.abi',JSON.stringify(a.abi));fs.writeFileSync('$(ABI_TMP)/$$c.bin',a.bytecode.slice(2))" && \
		(cd backend && go run github.com/ethereum/go-ethereum/cmd/abigen@$(GETH_VERSION) --abi tmp/abi/$$c.abi --bin tmp/abi/$$c.bin \
			--pkg bindings --type $$c --out internal/chain/bindings/$$(echo $$c | tr A-Z a-z).go) && echo "generated $$c" || exit 1; \
	done
