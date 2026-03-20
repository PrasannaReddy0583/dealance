.PHONY: infra-up infra-down migrate dev test build clean keys

# --- Infrastructure ---

infra-up:
	@echo "Starting infrastructure..."
	cd infrastructure && docker compose up -d
	@echo "Waiting for services to be healthy..."
	@sleep 5
	@echo "Infrastructure is ready."

infra-down:
	@echo "Stopping infrastructure..."
	cd infrastructure && docker compose down

infra-reset:
	@echo "Resetting infrastructure (all data will be lost)..."
	cd infrastructure && docker compose down -v
	$(MAKE) infra-up

# --- Migrations ---

migrate:
ifndef service
	$(error service is required. Usage: make migrate service=auth)
endif
	@echo "Running migrations for $(service)..."
	cd services/$(service) && go run -tags migrate cmd/migrate/main.go up

migrate-down:
ifndef service
	$(error service is required. Usage: make migrate-down service=auth)
endif
	@echo "Rolling back migrations for $(service)..."
	cd services/$(service) && go run -tags migrate cmd/migrate/main.go down

migrate-sql:
ifndef service
	$(error service is required. Usage: make migrate-sql service=auth)
endif
	@echo "Running SQL migration for $(service)..."
	PGPASSWORD=dealance psql -h localhost -U dealance -d dealance_$(service) -f services/$(service)/migrations/001_schema.up.sql

migrate-sql-down:
ifndef service
	$(error service is required. Usage: make migrate-sql-down service=auth)
endif
	@echo "Rolling back SQL migration for $(service)..."
	PGPASSWORD=dealance psql -h localhost -U dealance -d dealance_$(service) -f services/$(service)/migrations/001_schema.down.sql

# --- Development ---

dev:
ifndef service
	$(error service is required. Usage: make dev service=auth)
endif
	@echo "Starting $(service) service in dev mode..."
	cd services/$(service) && APP_ENV=development go run cmd/server/main.go

# --- Testing ---

test:
	@echo "Running all tests..."
	cd shared && go test ./... -v -count=1 -coverprofile=coverage.out
	cd services/auth && go test ./... -v -count=1 -coverprofile=coverage.out

test-shared:
	@echo "Testing shared packages..."
	cd shared && go test ./... -v -count=1

test-auth:
	@echo "Testing auth service..."
	cd services/auth && go test ./... -v -count=1

test-coverage:
	@echo "Running tests with coverage..."
	cd services/auth && go test ./... -v -count=1 -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html

# --- Build ---

build:
ifndef service
	$(error service is required. Usage: make build service=auth)
endif
	@echo "Building $(service)..."
	cd services/$(service) && CGO_ENABLED=0 GOOS=linux go build -o bin/server cmd/server/main.go

# --- Keys ---

keys:
	@echo "Generating RSA key pair for JWT..."
	mkdir -p infrastructure/keys
	openssl genrsa -out infrastructure/keys/private.pem 2048
	openssl rsa -in infrastructure/keys/private.pem -pubout -out infrastructure/keys/public.pem
	@echo "Keys generated in infrastructure/keys/"

# --- Clean ---

clean:
	@echo "Cleaning build artifacts..."
	find services -name "bin" -type d -exec rm -rf {} + 2>/dev/null || true
	find . -name "coverage.out" -delete 2>/dev/null || true
	find . -name "coverage.html" -delete 2>/dev/null || true

# --- Vet ---

vet:
	@echo "Running go vet..."
	cd shared && go vet ./...
	cd services/auth && go vet ./...
