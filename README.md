# Industry Tool

[![CI](https://github.com/YOUR_USERNAME/industry-tool/actions/workflows/ci.yml/badge.svg)](https://github.com/YOUR_USERNAME/industry-tool/actions/workflows/ci.yml)

An EVE Online industry and asset management tool.

## Features

- Asset inventory tracking (character and corporation)
- Stockpile markers for tracking desired quantities
- Support for multiple characters and corporations
- Real-time asset synchronization with EVE Online ESI

## Development

### Prerequisites

- Docker & Docker Compose
- Go 1.25+
- Node.js 20+

### Running Locally

```bash
# Start development environment
make dev

# Run tests
make test

# Run backend tests only
make test-backend

# Run frontend tests only
make test-frontend

# Clean up
make dev-clean
```

### Testing

All tests run in Docker containers to ensure consistency:

- **Backend**: Go tests with PostgreSQL database
- **Frontend**: Jest snapshot tests with React Testing Library

Coverage reports are generated in `artifacts/coverage/`:
- Backend: `artifacts/coverage/backend/coverage.html`
- Frontend: `artifacts/coverage/frontend/lcov-report/index.html`

### Architecture

- **Backend**: Go with PostgreSQL
- **Frontend**: Next.js (React) with TypeScript
- **API**: EVE Online ESI (External Swagger Interface)

## CI/CD

GitHub Actions automatically runs tests on all pull requests. PRs will be blocked if tests fail.

## License

[Add your license here]
