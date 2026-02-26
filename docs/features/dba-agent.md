# DBA Agent

## Status
Implemented — available for use in planning and implementation workflows.

## Overview

A specialized Claude Code agent for database administration tasks. The DBA agent provides schema context during feature planning, reviews migrations before implementation, identifies query optimization opportunities, and maintains the living schema reference document.

## Key Decisions

- **Model**: Sonnet — needs to reason about complex multi-JOIN SQL, index selectivity, and cascade implications
- **Role**: Research/advisory + documentation — does NOT write Go code (that's backend-dev's job)
- **Tools**: Read, Write, Edit, Bash, Glob, Grep — no Task(executor) since its bash usage is limited to short psql queries
- **Schema doc**: Single centralized `docs/database-schema.md` rather than per-table docs

## When to Spawn

1. **During plan mode** — get schema context before designing features that touch the database
2. **Before migration writing** — draft migration SQL for backend-dev to implement
3. **After implementation** — refresh `docs/database-schema.md`
4. **Optimization reviews** — audit repository queries for view/function/index opportunities

## Capabilities

### Schema Context
Provides table definitions, relationships, and existing patterns when the planner is designing a new feature.

### Migration Review
Reviews for naming conflicts, missing indexes, cascade behavior, data type consistency, and up/down symmetry.

### Proactive Optimization
Identifies opportunities for PostgreSQL features that reduce query duplication:
- **Views** for repeated JOIN chains
- **Materialized views** for expensive aggregations
- **SQL functions** for polymorphic lookups (owner names, location names)
- **Missing indexes** for common query patterns
- **Partial indexes** for filtered queries

### Known Duplication Patterns
These patterns are documented in the agent instructions and are candidates for database-level optimization:
- Owner name resolution (character/corporation polymorphic lookup) — 4+ files
- Location name resolution (station/solar_system COALESCE) — 8+ files
- Item type + market price + stockpile JOIN chain — 6+ files
- Container location name resolution — 2+ files
- Character + corporation asset UNION ALL — 2+ files

## File Paths

- Agent instructions: `.claude/agents/dba.md`
- Agent memory: `.claude/agent-memory/dba/MEMORY.md` (gitignored, ephemeral)
- Schema reference: `docs/database-schema.md` (maintained by DBA agent)

## Workflow Integration

```
1. Planner reads feature doc
2. Planner spawns DBA: "What tables/indexes are needed for [feature]?"
3. DBA returns schema context + migration draft + optimization recommendations
4. Planner spawns backend-dev with DBA's migration design
5. backend-dev implements migration, repository, controller, tests
6. DBA updates docs/database-schema.md
```
