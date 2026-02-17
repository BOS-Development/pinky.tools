# Industry Tool

[![CI](https://github.com/BOS-Development/pinky.tools/actions/workflows/ci.yml/badge.svg)](https://github.com/YOUR_USERNAME/industry-tool/actions/workflows/ci.yml)
[![Coverage Status](https://coveralls.io/repos/github/YOUR_USERNAME/industry-tool/badge.svg?branch=main)](https://coveralls.io/github/YOUR_USERNAME/industry-tool?branch=main)

An EVE Online industry and asset management tool.

## Features

### Asset Management
- Asset inventory tracking (character and corporation)
- Stockpile markers for tracking desired quantities
- Support for multiple characters and corporations
- Real-time asset synchronization with EVE Online ESI

### Marketplace System
- **Contact System**: Bidirectional contact relationships with granular permissions
- **For-Sale Listings**: List items from any inventory location with custom pricing
- **Purchase System**: Complete transaction workflow with contract integration
  - Atomic purchase transactions
  - Multi-stage workflow: pending → contract_created → completed
  - Cancel and restore functionality
  - Full transaction history

See [Purchase System Documentation](docs/features/purchases/) for complete documentation.

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

### Documentation

**Purchase System (Phase 4):**
- **[PURCHASES.md](docs/features/purchases/PURCHASES.md)** - Complete technical documentation
  - Architecture and database schema
  - Business logic and workflows
  - Error handling and troubleshooting
  - Performance and security considerations
- **[API_PURCHASES.md](docs/features/purchases/API_PURCHASES.md)** - API reference
  - Endpoint specifications
  - Request/response formats
  - Error codes and handling
  - TypeScript types
- **[QUICK_START_PURCHASES.md](docs/features/purchases/QUICK_START_PURCHASES.md)** - Quick start guide
  - Step-by-step purchase flow
  - Common operations
  - cURL examples
- **[TESTING_PURCHASES.md](docs/features/purchases/TESTING_PURCHASES.md)** - Test suite documentation
  - Test coverage summary (19 tests)
  - Running tests
  - Bug fixes verified

**Other Features:**
- **[Contact & Marketplace](docs/features/contact-marketplace.md)** - Contact system and marketplace overview
- **[Jita Market Pricing](docs/features/jita-market-pricing.md)** - Market price integration

## CI/CD

GitHub Actions automatically runs tests on all pull requests. PRs will be blocked if tests fail.

## License

[Add your license here]
