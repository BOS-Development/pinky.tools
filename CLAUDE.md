# Critical Rules

1. **Create a feature branch before writing any code.** `git checkout -b feature/{name}` — do this before touching any files. Never commit directly to main.
2. **Always run tests with Makefile targets** (`make test-e2e-ci`, `make test-backend`, etc.) — never run test commands directly.
3. **Create/update feature docs in `docs/features/`** for every new feature or significant change. Use `lowercase-kebab-case.md` naming (e.g., `sde-import.md`, `reactions-calculator.md`).

@context.md
