.PHONY: build test test-backend test-frontend test-all test-clean dev dev-clean dev-build

# Docker Compose command (v1: docker-compose, v2: docker compose)
DOCKER_COMPOSE ?= docker-compose

build: clean
	$(DOCKER_COMPOSE) build

dev: dev-build
	$(DOCKER_COMPOSE) \
		-f docker-compose.yaml \
		-f docker-compose.dev.yaml \
		up

clean:
	$(DOCKER_COMPOSE) down

dev-build: dev-clean
	$(DOCKER_COMPOSE) \
		-f docker-compose.yaml \
		-f docker-compose.dev.yaml \
		build

dev-clean:
	$(DOCKER_COMPOSE) \
		-f docker-compose.yaml \
		-f docker-compose.dev.yaml \
		down

dev-destroy:
	$(DOCKER_COMPOSE) \
		-f docker-compose.yaml \
		-f docker-compose.dev.yaml \
		down --volumes

integration-test: integration-test-build
	$(DOCKER_COMPOSE) \
		-f docker-compose.ci.yaml \
		up

integration-test-build: integration-test-clean
	$(DOCKER_COMPOSE) \
		-f docker-compose.ci.yaml \
		build

integration-test-clean:
	$(DOCKER_COMPOSE) \
		-f docker-compose.ci.yaml \
		down --remove-orphans

generate:
	go generate ./internal/...

# Test targets
test-backend:
	@echo "Running backend tests with coverage..."
	@mkdir -p artifacts/coverage/backend
	$(DOCKER_COMPOSE) -f docker-compose.test.yaml run --rm backend-test \
		sh -c "go test -v -coverprofile=/artifacts/coverage/backend/coverage.out ./internal/... && \
		       go tool cover -html=/artifacts/coverage/backend/coverage.out -o /artifacts/coverage/backend/coverage.html"
	@echo "✓ Backend coverage report: artifacts/coverage/backend/coverage.html"

test-frontend:
	@echo "Running frontend tests with coverage..."
	@mkdir -p artifacts/coverage/frontend
	$(DOCKER_COMPOSE) -f docker-compose.test.yaml run --rm frontend-test \
		sh -c "npm install && npm run test:coverage -- --coverageDirectory=/artifacts/coverage/frontend"
	@echo "✓ Frontend coverage report: artifacts/coverage/frontend/lcov-report/index.html"

test-all: test-clean
	@echo "Running all tests with coverage..."
	@$(MAKE) test-backend
	@$(MAKE) test-frontend
	@echo ""
	@echo "========================================"
	@echo "✓ All tests completed successfully!"
	@echo "========================================"
	@echo "Backend coverage:  artifacts/coverage/backend/coverage.html"
	@echo "Frontend coverage: artifacts/coverage/frontend/lcov-report/index.html"
	@echo ""

test-clean:
	@echo "Cleaning up test containers..."
	@$(DOCKER_COMPOSE) -f docker-compose.test.yaml down --remove-orphans 2>/dev/null || true

test: test-all