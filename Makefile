# Help
## ==============================
help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)



DATABASE_URL := postgres://postgres:postgres@localhost:5432/filestore?sslmode=disable

.PHONY: run
run: db-up ## Run the server
	docker compose up --build -d server

.PHONY: db-up
db-up: ## Start the database
	./scripts/migrations_db_up.sh

.PHONY: db-down
db-down: ## Stop the database
	./scripts/migrations_db_down.sh

.PHONY: db-migrate-up
db-migrate-up: ## Migrate the database
	./scripts/migrations_db_migrate.sh

.PHONY: db-migrate-down
db-migrate-down: ## Rollback the database
	./scripts/migrations_db_rollback.sh

.PHONY: db-create db-new
db-new: ## Create a new migration
db-create: ## Create a new migration
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make db-create NAME=migration_name"; \
		exit 1; \
	fi
	./scripts/migrations_db_create.sh "$(NAME)"

.PHONY: build-migrations-image
build-migrations-image: ## Build the migrations image
	docker build -t filestore-migrations:latest -f Dockerfile.migrations .
