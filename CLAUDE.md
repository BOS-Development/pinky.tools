# Critical Rules

1. **Create a feature branch off an updated main before writing any code.** `git checkout main && git pull origin main && git checkout -b feature/{name}` — always branch from the latest main. Never branch from another feature branch. Never commit directly to main.
2. **Always run tests with Makefile targets** (`make test-e2e-ci`, `make test-backend`, etc.) — never run test commands directly.
3. **Write tests for every new backend file.** Every new repository, controller, and updater must have a corresponding `_test.go` file. When asked to "write tests," cover all three layers — not just controllers and updaters.
4. **All changes must have tests.** Every code change — bug fixes, new features, refactors — must include corresponding test coverage. No code changes without tests.
5. **Frontend-touching features require E2E tests.** Any new page, UI feature, or user-facing workflow change must include E2E test coverage. Spawn the `sdet` agent for test files. Use the mock ESI admin API (`e2e/helpers/mock-esi.ts`) when tests need to control ESI responses at runtime.
6. **Create/update feature docs in `docs/features/{category}/`** for every new feature or significant change. Use `lowercase-kebab-case.md` naming. Categories: `core/`, `market/`, `social/`, `trading/`, `industry/`, `infrastructure/`. Agent docs go in `docs/agents/`. Spawn the `docs` agent to create docs and update `docs/features/INDEX.md`.
7. **Go slices: initialize as `items := []*Type{}` NOT `var items []*Type`.** Prevents nil JSON marshaling (`null` instead of `[]`).
8. **Do NOT include Discord usernames or other personal attributions in GitHub issues.**
9. **Check feature docs first.** Before exploring code or planning a feature, check `docs/features/INDEX.md` to find the relevant doc. Feature docs contain schema, API, key decisions, and file paths — use them as the starting point.
10. Always use the executor sub-agent for bash commands instead of running them directly.
11. **Delegate all implementation work to domain agents.** Never write Go, SQL, or migration code directly — use the `backend-dev` agent. Never write React, TypeScript, or MUI code directly — use the `frontend-dev` agent. Never write Playwright tests, mock ESI code, or E2E seed data directly — use the `sdet` agent. **For database schema design, migration review, or query optimization, spawn the `dba` agent first** — it provides schema context, migration drafts, and optimization recommendations before backend-dev implements. The main thread plans and orchestrates; agents execute. For cross-cutting tasks (e.g., new API endpoint), spawn both backend and frontend agents.

---

## Git Workflow

- Branch naming: `feature/{feature-name}` or `fix/{bug-name}`
- Before starting work: `git checkout main && git pull origin main && git checkout -b feature/your-feature-name`
- Commit frequently with clear messages
- Push branch and create PR when ready for review
- **Only the main planner thread manages branches.** Sub-agents (backend-dev, frontend-dev, executor, etc.) must NEVER create, switch, or manage git branches. They write code on whatever branch is already checked out.

---

## Test Commands

```bash
make test-backend        # Backend Go tests with coverage
make test-frontend       # Frontend Jest tests with coverage
make test-all            # Run both backend and frontend tests
make test-e2e            # E2E tests headless (Playwright)
make test-e2e-ui         # E2E tests with Playwright UI
```

## Pre-PR Verification

Before pushing and creating a PR, run the production builds to catch strict TypeScript and compilation errors that tests alone may miss:

```bash
make build-production    # Builds both backend and frontend Docker images
```

This catches issues like untyped arrays (`const x = []` → `never[]`) that pass Jest but fail under Next.js strict TypeScript checking.

---

## Common Task Checklists

### Add New API Endpoint

1. Create handler in `internal/controllers/`
2. Create controller test file `mycontroller_test.go`
3. Register route in `internal/web/router.go`
4. Add client method in `frontend/packages/client/api.ts`
5. Create frontend API route in `frontend/pages/api/`
6. Call from component via `getServerSideProps`

### Update Database Schema

1. Spawn `dba` agent to analyze existing schema, check for naming conflicts, and draft migration SQL
2. Spawn `backend-dev` with the DBA's migration draft — it runs `./scripts/new-migration.sh` and writes the SQL
3. backend-dev implements/updates repository methods and tests
4. Restart server to auto-apply migration
5. Spawn `dba` to update `docs/database-schema.md`

### Add New Frontend Feature

1. Implement the feature (frontend-dev, backend-dev agents)
2. If feature uses ESI data not yet mocked → update `cmd/mock-esi/main.go` (backend-dev agent)
3. If feature needs new seed data → update `e2e/seed.sql` (sdet agent)
4. Write E2E tests in `e2e/tests/NN-feature.spec.ts` (sdet agent)
5. Run `make test-e2e` to verify all tests pass
6. Update feature docs (docs agent)

---

## Planning Complex Features

- For multi-file changes or architectural decisions, use plan mode first
- Present options to the user when multiple approaches are valid
- Break down large features into phases with clear verification steps
- Always check `docs/features/INDEX.md` for existing feature docs before starting work

---

## Agent Improvement

After completing a task (all agents done, tests passing), review agent output for improvement opportunities:
- If an agent was missing context, add it to `.claude/agents/{agent-name}.md`
- If an agent used workarounds, add the correct pattern to its conventions
- If a convention was unclear, make it more explicit with examples
- If the agent discovered a new project pattern, document it in the agent instructions
- **Review agent memory files** (`.claude/agent-memory/{agent}/MEMORY.md`): Agents accumulate session-specific learnings here. Agent memory files are gitignored and ephemeral — promote anything worth keeping:
  - Reusable patterns and conventions → `.claude/agents/{agent-name}.md`
  - Domain knowledge, algorithms, bug fixes, key decisions → `docs/features/{category}/` feature docs
- **Spawn the `docs` agent** after feature work to create/update feature docs and keep `docs/features/INDEX.md` current

---

@context.md
