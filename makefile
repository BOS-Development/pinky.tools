.PHONY: build test test-backend test-backend-sde-integration test-frontend test-all test-clean dev dev-clean dev-build build-production build-production-backend build-production-frontend test-e2e test-e2e-ui test-e2e-clean test-e2e-debug test-e2e-ci

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

test-backend-sde-integration:
	@echo "Running SDE integration test (downloads real SDE from CCP)..."
	$(DOCKER_COMPOSE) -f docker-compose.test.yaml run --rm backend-test \
		sh -c "SDE_INTEGRATION_TEST=1 go test -v -run Test_SdeClient_Integration -timeout 300s ./internal/client/"
	@echo "✓ SDE integration test passed"

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

# E2E test targets
test-e2e: test-e2e-clean
	@echo "Starting E2E test environment..."
	$(DOCKER_COMPOSE) -f docker-compose.e2e.yaml up -d --build
	@echo "Waiting for frontend to be ready..."
	@for i in $$(seq 1 60); do curl -sf http://localhost:3000/ > /dev/null 2>&1 && break; echo "  attempt $$i/60..."; sleep 5; done
	@curl -sf http://localhost:3000/ > /dev/null 2>&1 || (echo "Frontend failed to start"; exit 1)
	@echo "Installing Playwright dependencies..."
	cd e2e && npm ci
	@echo "Running Playwright tests..."
	cd e2e && npx playwright test
	@echo "✓ E2E tests completed"

test-e2e-ui: test-e2e-clean
	@echo "Starting E2E test environment..."
	$(DOCKER_COMPOSE) -f docker-compose.e2e.yaml up -d --build
	@echo "Waiting for frontend to be ready..."
	@for i in $$(seq 1 60); do curl -sf http://localhost:3000/ > /dev/null 2>&1 && break; echo "  attempt $$i/60..."; sleep 5; done
	@curl -sf http://localhost:3000/ > /dev/null 2>&1 || (echo "Frontend failed to start"; exit 1)
	@echo "Installing Playwright dependencies..."
	cd e2e && npm ci
	@echo "Opening Playwright UI..."
	cd e2e && npx playwright test --ui

test-e2e-clean:
	@echo "Cleaning up E2E test containers..."
	@$(DOCKER_COMPOSE) -f docker-compose.e2e.yaml down --remove-orphans --volumes 2>/dev/null || true
	@docker run --rm -v "$$(pwd):/work" alpine sh -c "rm -rf /work/e2e/node_modules /work/e2e/test-results /work/e2e/auth-state.json /work/artifacts/e2e-report" 2>/dev/null || true

# E2E tests in Docker (for CI pipelines — no local Node.js/Playwright needed)
# The playwright service has depends_on with health checks, so it waits for frontend automatically
test-e2e-ci: test-e2e-clean
	@echo "Starting E2E test environment..."
	$(DOCKER_COMPOSE) -f docker-compose.e2e.yaml up -d --build
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Running Playwright tests in Docker..."
	$(DOCKER_COMPOSE) -f docker-compose.e2e.yaml run --rm playwright
	@echo "✓ E2E CI tests completed"

# E2E environment for debugging (start services without running tests)
test-e2e-debug: test-e2e-clean
	@echo "Starting E2E test environment for debugging..."
	$(DOCKER_COMPOSE) -f docker-compose.e2e.yaml up -d --build
	@echo "Waiting for frontend to be ready..."
	@for i in $$(seq 1 60); do curl -sf http://localhost:3000/ > /dev/null 2>&1 && break; echo "  attempt $$i/60..."; sleep 5; done
	@curl -sf http://localhost:3000/ > /dev/null 2>&1 || (echo "Frontend failed to start"; exit 1)
	@echo ""
	@echo "========================================"
	@echo "E2E environment is ready for debugging"
	@echo "========================================"
	@echo "Frontend: http://localhost:3000"
	@echo "Backend:  http://localhost:8080"
	@echo "Mock ESI: http://localhost:8090"
	@echo "Database: localhost:19237"
	@echo ""
	@echo "Run 'cd e2e && npx playwright test' to execute tests"
	@echo "Run 'cd e2e && npx playwright test --ui' to open Playwright UI"
	@echo "Run 'make test-e2e-clean' to tear down"
	@echo ""

# Production build targets
build-production-backend:
	@echo "Building production backend image..."
	docker build --target final-backend -t industry-tool-backend:latest -f Dockerfile .
	@echo "✓ Backend production image built successfully"

build-production-frontend:
	@echo "Building production frontend image..."
	docker build --target publish-ui -t industry-tool-frontend:latest -f Dockerfile.ui .
	@echo "✓ Frontend production image built successfully"

build-production: build-production-backend build-production-frontend
	@echo ""
	@echo "========================================"
	@echo "✓ All production images built successfully!"
	@echo "========================================"
	@echo "Backend:  industry-tool-backend:latest"
	@echo "Frontend: industry-tool-frontend:latest"
	@echo ""