# Critical Rules

1. **Create a feature branch off an updated main before writing any code.** `git checkout main && git pull origin main && git checkout -b feature/{name}` — always branch from the latest main. Never branch from another feature branch. Never commit directly to main.
2. **Always run tests with Makefile targets** (`make test-e2e-ci`, `make test-backend`, etc.) — never run test commands directly.
3. **Write tests for every new backend file.** Every new repository, controller, and updater must have a corresponding `_test.go` file. When asked to "write tests," cover all three layers — not just controllers and updaters.
4. **All changes must have tests.** Every code change — bug fixes, new features, refactors — must include corresponding test coverage. No code changes without tests.
5. **Create/update feature docs in `docs/features/`** for every new feature or significant change. Use `lowercase-kebab-case.md` naming (e.g., `sde-import.md`, `reactions-calculator.md`).
6. **Go slices: initialize as `items := []*Type{}` NOT `var items []*Type`.** Prevents nil JSON marshaling (`null` instead of `[]`).
7. **Do NOT include Discord usernames or other personal attributions in GitHub issues.**
8. **Check feature docs first.** Before exploring code or planning a feature, read the relevant `docs/features/` doc (if one exists). Feature docs contain schema, API, key decisions, and file paths — use them as the starting point.

---

## Git Workflow

- Branch naming: `feature/{feature-name}` or `fix/{bug-name}`
- Before starting work: `git checkout main && git pull origin main && git checkout -b feature/your-feature-name`
- Commit frequently with clear messages
- Push branch and create PR when ready for review

---

## Backend (Go) Rules

- Use transactions with deferred `tx.Rollback()` for multi-statement operations
- Wrap errors with `github.com/pkg/errors` for context
- Follow `repository → controller → router` pattern
- Use `./scripts/new-migration.sh <name>` for timestamped migration files (generates both `.up.sql` and `.down.sql`)
- Migration format: `{YYYYMMDDHHMMSS}_{name}.up.sql` — lowercase SQL keywords, tab indentation

---

## Frontend (React/Next.js) Rules

- **MUI SSR**: ThemeRegistry must use Emotion cache (see `ThemeRegistry.tsx`)
- **Formatting**: Use utilities from `packages/utils/formatting.ts` (`formatISK`, `formatNumber`, `formatCompact`)
- **Authentication**: Check session status before rendering protected content
- **API Routes**: Proxy to backend, don't implement business logic
- Read existing components before creating similar ones

---

## Testing Requirements

### Backend
- Write integration tests in `*_test.go` files
- Test repository methods with real database (testcontainers)
- Cover success cases, edge cases, and error scenarios
- Use table-driven tests for multiple scenarios

### Frontend
- **Snapshot testing**: Create snapshots for all new components
  - Run `npm test -- -u` to update snapshots after intentional changes
  - Location: `__tests__/{ComponentName}.test.tsx`
  - Test loading, error, and success states
- Verify edge cases (empty data, errors, null values)
- Test both character and corporation flows if applicable

### Test Commands
```bash
make test-backend        # Backend Go tests with coverage
make test-frontend       # Frontend Jest tests with coverage
make test-all            # Run both backend and frontend tests
make test-e2e            # E2E tests headless (Playwright)
make test-e2e-ui         # E2E tests with Playwright UI
```

---

## Common Task Checklists

### Add New Repository
1. Create `internal/repositories/myrepo.go`
2. Implement struct with `*sql.DB`
3. Add methods with transactions
4. Create test file `myrepo_test.go`
5. Wire up in `cmd/cmd/root.go`

### Add New API Endpoint
1. Create handler in `internal/controllers/`
2. Create controller test file `mycontroller_test.go`
3. Register route in `internal/web/router.go`
4. Add client method in `frontend/packages/client/api.ts`
5. Create frontend API route in `frontend/pages/api/`
6. Call from component via `getServerSideProps`

### Add New Component
1. Create in `frontend/packages/components/`
2. Use MUI components for consistency
3. Follow naming: `Item` (card), `List` (grid)
4. Create snapshot test in `__tests__/`

### Update Database Schema
1. Run `./scripts/new-migration.sh migration_name`
2. Write SQL in generated `.up.sql` and `.down.sql`
3. Restart server to auto-apply
4. Update repository methods as needed

---

## UI/UX Standards

- Dark theme: Background `#0a0e1a`, Cards `#12151f`, Primary `#3b82f6`
- Color coding: Green `#10b981` for revenue/success, Red `#ef4444` for costs/errors
- Icons: Use MUI icons from `@mui/icons-material`
- Tables: Dark header (`#0f1219`), alternating row colors, right-align numbers
- Loading states: Use `<Loading />` component, not custom spinners
- Empty states: Centered message in table cell with `colSpan`

---

## File Organization

- Backend: `internal/repositories/` → `internal/controllers/` → `cmd/cmd/root.go`
- Frontend components: `packages/components/{feature}/{ComponentName}.tsx`
- API routes: `pages/api/{feature}/{action}.ts`
- Types/interfaces: Define in component file or `internal/models/models.go`
- Feature docs: `docs/features/`

---

## Common Pitfalls

1. **Go nil slices → JSON null**: Always initialize `items := []*Type{}`
2. **MUI FOUC**: Ensure ThemeRegistry has Emotion cache setup
3. **Missing auth headers**: Backend needs `BACKEND-KEY` and `USER-ID`
4. **Incomplete transactions**: Always defer `tx.Rollback()` before operations
5. **Hardcoded IDs**: Use session providerAccountId, not hardcoded values

---

## Planning Complex Features

- For multi-file changes or architectural decisions, use plan mode first
- Present options to the user when multiple approaches are valid
- Break down large features into phases with clear verification steps
- Always check for existing feature plans in `docs/features/` before starting work

---

@context.md
