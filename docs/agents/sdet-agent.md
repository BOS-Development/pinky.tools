# SDET Agent

## Status
Implemented — available for use in planning and implementation workflows.

## Overview

A specialized Claude Code agent for end-to-end test creation and maintenance. The SDET agent writes Playwright browser tests, updates the mock ESI server with new canned data, manages seed SQL, and maintains auth fixtures. It owns all files in the `e2e/` directory and the mock ESI server at `cmd/mock-esi/main.go`.

## Key Decisions

- **Model**: Sonnet — needs to understand complex UI flows, async timing, and the full-stack data pipeline
- **Role**: Implementation — writes test code, mock data, and seed SQL. Does NOT write application code or documentation
- **Tools**: Read, Write, Edit, Bash, Glob, Grep, Task(executor) — same toolset as backend-dev and frontend-dev
- **Scope**: E2E tests only — unit tests remain with backend-dev (Go) and frontend-dev (Jest)

## When to Spawn

1. **New feature E2E coverage** — after a feature is implemented, spawn to add E2E tests
2. **Mock ESI updates** — when a new ESI endpoint needs canned data for testing
3. **E2E test failures** — to debug and fix broken tests
4. **Test maintenance** — when app changes break existing tests (updated selectors, changed flows)
5. **Seed data updates** — when new static data is needed for test scenarios

## Capabilities

### Test Creation
Writes Playwright test specs following project conventions: numbered sequential files, proper auth fixture selection (single-user vs multi-user), accessible locator patterns, and appropriate timeout strategies.

### Mock ESI Management
Extends the mock ESI server (`cmd/mock-esi/main.go`) with new endpoints and canned data when features require ESI data not yet mocked.

### Seed Data Management
Updates `e2e/seed.sql` when tests need new static universe data (regions, stations, item types) that can't be created through the app.

### Auth Fixture Management
Maintains `e2e/fixtures/auth.ts` for multi-user test scenarios. Understands the four test users (Alice, Bob, Charlie, Diana) and their roles.

## File Paths

- Agent instructions: `.claude/agents/sdet.md`
- Agent memory: `.claude/agent-memory/sdet/MEMORY.md` (gitignored, ephemeral)
- Test specs: `e2e/tests/*.spec.ts`
- Auth fixtures: `e2e/fixtures/auth.ts`
- Playwright config: `e2e/playwright.config.ts`
- Global setup: `e2e/global-setup.ts`
- Seed data: `e2e/seed.sql`
- Mock ESI server: `cmd/mock-esi/main.go`
- E2E API routes: `frontend/pages/api/e2e/`
- Docker config: `docker-compose.e2e.yaml`

## Workflow Integration

```
1. Planner identifies feature needing E2E coverage
2. Planner spawns SDET: "Write E2E tests for [feature]. Here's what the UI does..."
3. SDET reads existing tests to understand state dependencies
4. SDET reads feature doc and relevant frontend components for selectors
5. SDET creates test spec, updates mock ESI / seed data if needed
6. SDET runs make test-e2e to verify all tests pass
7. Planner reviews and commits
```
