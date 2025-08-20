export TESTSERVER_URL ?= http://localhost:8080/

# Load .env file if it exists
-include .env
export

run:
	go run main.go

update_docs_run:
	swag init; go run main.go;

test:
	go clean -testcache && go test ./tests -v -p 1 -parallel 1

truncate-remote-db:
	@echo "Truncating all tables except schema_migrations..."
	@psql postgresql://$(PG_USER):$(PG_PASSWORD)@$(PG_HOST):$(PG_PORT)/$(PG_DB_NAME)?sslmode=$(PG_SSLMODE) -c "DO \$$\$$ DECLARE r RECORD; BEGIN FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename != 'schema_migrations') LOOP EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' CASCADE'; END LOOP; END \$$\$$;"
	@echo "Database truncated successfully!"

clean-test: truncate-remote-db test

# Docker commands
DOCKER_IMAGE_NAME ?= fusioncat
DOCKER_TAG ?= latest
DOCKER_FULL_NAME = $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

# Build the Docker image
docker-build:
	@echo "Building Docker image $(DOCKER_FULL_NAME)..."
	docker build -t $(DOCKER_FULL_NAME) -f deploy/Dockerfile .
	@echo "Docker image built successfully!"

# Run the Docker container for testing
docker-run: docker-build
	@echo "Running Docker container..."
	docker run --rm \
		-p 8080:8080 \
		-e PG_HOST=$(PG_HOST) \
		-e PG_PORT=$(PG_PORT) \
		-e PG_USER=$(PG_USER) \
		-e PG_PASSWORD=$(PG_PASSWORD) \
		-e PG_DB_NAME=$(PG_DB_NAME) \
		-e PG_SSLMODE=$(PG_SSLMODE) \
		-e JWT_SECRET=$(JWT_SECRET) \
		-e ADMIN_URL=$(ADMIN_URL) \
		-e PATH_TO_STUBS_TEMPLATES_FOLDER=/app/templates \
		-e JSON_SCHEMA_CONVERTOR_CMD=/usr/bin/quicktype \
		--name fusioncat-test \
		$(DOCKER_FULL_NAME)

# Test Docker build and basic functionality
docker-test: docker-build
	@echo "Testing Docker build and container startup..."
	@echo "Starting container in background..."
	docker run -d --rm \
		-p 8080:8080 \
		-e PG_HOST=$(PG_HOST) \
		-e PG_PORT=$(PG_PORT) \
		-e PG_USER=$(PG_USER) \
		-e PG_PASSWORD=$(PG_PASSWORD) \
		-e PG_DB_NAME=$(PG_DB_NAME) \
		-e PG_SSLMODE=$(PG_SSLMODE) \
		-e JWT_SECRET=$(JWT_SECRET) \
		-e ADMIN_URL=$(ADMIN_URL) \
		-e PATH_TO_STUBS_TEMPLATES_FOLDER=/app/templates \
		-e JSON_SCHEMA_CONVERTOR_CMD=/usr/bin/quicktype \
		--name fusioncat-test \
		$(DOCKER_FULL_NAME)
	@echo "Waiting for container to start..."
	@sleep 5
	@echo "Checking if container is running..."
	@docker ps | grep fusioncat-test || (echo "Container failed to start!" && exit 1)
	@echo "Testing health endpoint..."
	@curl -f http://localhost:8080/health || (echo "Health check failed!" && docker logs fusioncat-test && docker stop fusioncat-test && exit 1)
	@echo "Container started successfully! Stopping test container..."
	@docker stop fusioncat-test
	@echo "Docker test completed successfully!"

# Clean up Docker images
docker-clean:
	@echo "Removing Docker image $(DOCKER_FULL_NAME)..."
	docker rmi $(DOCKER_FULL_NAME) || true
	@echo "Docker cleanup completed!"

# Show Docker logs
docker-logs:
	docker logs fusioncat-test

# Execute shell in running container
docker-shell:
	docker exec -it fusioncat-test /bin/bash

# GitHub Actions testing
gh-actions-test:
	@echo "Testing GitHub Actions locally with act..."
	@echo "'authentication required' may occur when validation is successful."
	@./deploy/test-github-actions.sh
