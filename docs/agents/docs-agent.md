# Docs Agent

## Status
Implemented — available for use in planning and implementation workflows.

## Overview

A lightweight Claude Code agent for documentation maintenance. The docs agent keeps the feature index current, ensures cross-references stay valid after doc moves, scaffolds new feature docs in the right category, and audits documentation health.

## Key Decisions

- **Model**: Haiku — text organization doesn't require complex reasoning
- **Role**: Maintenance + advisory — creates/updates docs, reports organization issues
- **Tools**: Read, Write, Edit, Glob, Grep — no Bash needed
- **Index**: Single `docs/features/INDEX.md` as the quick-reference entry point

## When to Spawn

1. **After implementing a feature** — create/update the feature doc, update INDEX.md
2. **After moving or renaming docs** — fix cross-references in context.md, README.md, inter-doc links
3. **During planning** — find all relevant existing docs for a feature area
4. **Periodic audit** — check for stale docs, broken links, misplaced docs, missing coverage

## Capabilities

### Index Maintenance
Keeps `docs/features/INDEX.md` in sync with the actual docs on disk. Adds/removes/updates entries when docs change.

### Cross-Reference Integrity
Greps for old paths after doc moves and updates all references across the project (context.md, README.md, CLAUDE.md, agent configs, inter-doc links).

### New Feature Doc Scaffolding
Creates properly structured feature docs in the correct category folder using a standard template (status, overview, key decisions, schema, API, file paths).

### Organization Recommendations
Audits the docs tree and recommends:
- Splitting overgrown categories (10+ docs)
- Moving misplaced docs to better-fitting categories
- Creating docs for undocumented code areas
- Removing or updating stale docs

### Staleness Detection
Cross-checks doc contents (file paths, table names, endpoints) against the actual codebase to find outdated references.

## Documentation Structure

```
docs/
├── agents/              # Agent docs (this file, dba-agent.md)
├── database-schema.md   # Schema reference (DBA agent)
└── features/
    ├── INDEX.md         # Quick-reference (docs agent)
    ├── core/            # Auth, characters, corps, SDE, background updates
    ├── market/          # Pricing, aggregation, stockpiles
    ├── social/          # Contacts, marketplace, Discord
    ├── trading/         # Purchases, buy orders, contracts, auto-trading
    ├── industry/        # Manufacturing, reactions, PI, transportation
    └── infrastructure/  # Testing, deployment
```

## File Paths

- Agent instructions: `.claude/agents/docs.md`
- Agent memory: `.claude/agent-memory/docs/MEMORY.md` (gitignored, ephemeral)
- Feature index: `docs/features/INDEX.md` (maintained by docs agent)

## Workflow Integration

```
1. Developer implements a feature
2. Planner spawns docs agent: "Create feature doc for [feature] in [category]"
3. Docs agent creates the doc and updates INDEX.md
4. After any doc reorganization, planner spawns docs agent to fix cross-references
5. Periodically: planner spawns docs agent for health audit
```
