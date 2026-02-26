# Critical Rules

1. **Create a feature branch off an updated main before writing any code.** `git checkout main && git pull origin main && git checkout -b feature/{name}` — always branch from the latest main. Never branch from another feature branch. Never commit directly to main.
2. **Always run tests with Makefile targets** (`make test-e2e-ci`, `make test-backend`, etc.) — never run test commands directly.
3. **Write tests for every new backend file.** Every new repository, controller, and updater must have a corresponding `_test.go` file. When asked to "write tests," cover all three layers — not just controllers and updaters.
4. **All changes must have tests.** Every code change — bug fixes, new features, refactors — must include corresponding test coverage. No code changes without tests.
5. **Create/update feature docs in `docs/features/`** for every new feature or significant change. Use `lowercase-kebab-case.md` naming (e.g., `sde-import.md`, `reactions-calculator.md`).
6. **Go slices: initialize as `items := []*Type{}` NOT `var items []*Type`.** Prevents nil JSON marshaling (`null` instead of `[]`).
7. **Do NOT include Discord usernames or other personal attributions in GitHub issues.**
8. **Check feature docs first.** Before exploring code or planning a feature, read the relevant `docs/features/` doc (if one exists). Feature docs contain schema, API, key decisions, and file paths — use them as the starting point.
9. Always use the executor sub-agent for bash commands instead of running them directly.
10. **Delegate all implementation work to domain agents.** Never write Go, SQL, or migration code directly — use the `backend-dev` agent. Never write React, TypeScript, or MUI code directly — use the `frontend-dev` agent. **For database schema design, migration review, or query optimization, spawn the `dba` agent first** — it provides schema context, migration drafts, and optimization recommendations before backend-dev implements. The main thread plans and orchestrates; agents execute. For cross-cutting tasks (e.g., new API endpoint), spawn both backend and frontend agents.

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

---

## Planning Complex Features

- For multi-file changes or architectural decisions, use plan mode first
- Present options to the user when multiple approaches are valid
- Break down large features into phases with clear verification steps
- Always check for existing feature plans in `docs/features/` before starting work

---

## Agent Improvement

After completing a task (all agents done, tests passing), review agent output for improvement opportunities:
- If an agent was missing context, add it to `.claude/agents/{agent-name}.md`
- If an agent used workarounds, add the correct pattern to its conventions
- If a convention was unclear, make it more explicit with examples
- If the agent discovered a new project pattern, document it in the agent instructions
- **Review agent memory files** (`.claude/agent-memory/{agent}/MEMORY.md`): Agents accumulate session-specific learnings here. Agent memory files are gitignored and ephemeral — promote anything worth keeping:
  - Reusable patterns and conventions → `.claude/agents/{agent-name}.md`
  - Domain knowledge, algorithms, bug fixes, key decisions → `docs/features/` feature docs

---

@context.md
