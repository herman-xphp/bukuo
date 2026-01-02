DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_NAME=bukuo_db
TEST_DB_NAME=bukuo_test
DATABASE_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
TEST_DATABASE_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(TEST_DB_NAME)?sslmode=disable
MIGRATE_CMD=migrate -path internal/database/migrations -database "$(DATABASE_URL)" -verbose
TEST_MIGRATE_CMD=migrate -path internal/database/migrations -database "$(TEST_DATABASE_URL)" -verbose

.PHONY: run
run:
	go run cmd/api/main.go

.PHONY: seed
seed:
	go run cmd/seeder/main.go

.PHONY: test
test:
	@echo "Running tests with TEST_DB_NAME=$(TEST_DB_NAME)..."
	export DB_NAME=$(TEST_DB_NAME) && go test ./... -count=1

.PHONY: setup-test-db
setup-test-db:
	@echo "Setting up test database: $(TEST_DB_NAME)..."
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -c "DROP DATABASE IF EXISTS $(TEST_DB_NAME);" || true
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -c "CREATE DATABASE $(TEST_DB_NAME);"
	@echo "Running migrations on test database..."
	$(TEST_MIGRATE_CMD) up
	@echo "Test database setup complete."

.PHONY: migrate-up
migrate-up:
	$(MIGRATE_CMD) up

.PHONY: migrate-down
migrate-down:
	$(MIGRATE_CMD) down

.PHONY: migrate-force
migrate-force:
	$(MIGRATE_CMD) force $(version)

lint:
	golangci-lint run

.PHONY: docker-build
docker-build:
	docker build -t bukuo-api:latest .

.PHONY: test-pkg
test-pkg:
	@if [ -z "$(PKG)" ]; then echo "Usage: make test-pkg PKG=./path/to/package"; exit 1; fi
	go test -v $(PKG) -count=1

.PHONY: test-run
test-run:
	@if [ -z "$(RUN)" ]; then echo "Usage: make test-run RUN=TestName"; exit 1; fi
	go test -v ./... -run $(RUN) -count=1
