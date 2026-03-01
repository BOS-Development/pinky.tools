# Feature Documentation Index

Quick-reference for all feature docs. Each links to the full documentation.

## Core Systems

| Feature | Doc | Summary |
|---------|-----|---------|
| Characters | [characters.md](core/characters.md) | ESI OAuth, token storage, character asset refresh |
| Corporations | [corporations.md](core/corporations.md) | Corp management, divisions, corp asset refresh |
| Authentication | [consolidate-oauth.md](core/consolidate-oauth.md) | OAuth consolidation, scopes, single callback |
| ESI Scope Warnings | [esi-scope-warnings.md](core/esi-scope-warnings.md) | Scope update detection, re-auth warnings |
| Background Updates | [background-asset-updates.md](core/background-asset-updates.md) | Asset refresh runners (1h), concurrency |
| SDE Import | [sde-import.md](core/sde-import.md) | Static data pipeline, all `sde_*` tables, background runners |
| NPC Station Names | [npc-station-names.md](core/npc-station-names.md) | Station name resolution via ESI bulk endpoint |
| Landing Page | [landing-page.md](core/landing-page.md) | Hero section, asset value metrics |

## Market & Pricing

| Feature | Doc | Summary |
|---------|-----|---------|
| Asset Aggregation | [asset-aggregation.md](market/asset-aggregation.md) | SQL-level asset stacking/aggregation within scopes |
| Jita Market Pricing | [jita-market-pricing.md](market/jita-market-pricing.md) | Market orders, asset valuation |
| Stockpile Markers | [stockpile-markers.md](market/stockpile-markers.md) | Stockpile targets, deficit tracking, inventory UI |
| Stockpile Multibuy | [stockpile-multibuy.md](market/stockpile-multibuy.md) | Shopping lists, delta calculation, bulk ops |

## Social & Marketplace

| Feature | Doc | Summary |
|---------|-----|---------|
| Contact Marketplace | [contact-marketplace.md](social/contact-marketplace.md) | Contacts, permissions, for-sale marketplace |
| Contact Rules | [contact-rules.md](social/contact-rules.md) | Auto-create contacts, cascade cleanup |
| Discord Notifications | [discord-notifications.md](social/discord-notifications.md) | Discord bot, OAuth linking, event notifications |

## Purchase & Trade

| Feature | Doc | Summary |
|---------|-----|---------|
| Purchases | [purchases/](trading/purchases/) | Purchase transactions, contract workflow |
| Buy Orders | [buy-orders/](trading/buy-orders/) | Demand tracking, seller demand endpoints |
| Auto-Sell Containers | [auto-sell-containers.md](trading/auto-sell-containers.md) | Auto-sell config, Jita pricing, for-sale sync |
| Auto-Buy | [auto-buy.md](trading/auto-buy.md) | Auto-buy config, buy order management |
| Auto-Fulfill | [auto-fulfill.md](trading/auto-fulfill.md) | Match buy orders to for-sale listings |
| Contract Sync | [contract-sync.md](trading/contract-sync.md) | ESI contract polling, auto-complete |
| Contract Notifications | [contract-created-notification.md](trading/contract-created-notification.md) | Discord alerts on contract creation |
| Job Slot Rental Exchange | [job-slot-rental-exchange.md](trading/job-slot-rental-exchange.md) | Marketplace for renting idle industry job slots |

## Industry & Production

| Feature | Doc | Summary |
|---------|-----|---------|
| Industry Job Manager | [industry-job-manager/](industry/industry-job-manager/) | Skills sync, job tracking, manufacturing calc, job queue |
| Auto-Production | [auto-production.md](industry/auto-production.md) | Stockpile-driven background production plan runs |
| Reactions Calculator | [reactions-calculator.md](industry/reactions-calculator.md) | Moon reactions, batch ME, shopping list |
| Planetary Industry | [planetary-industry.md](industry/planetary-industry.md) | PI data, stall detection, profit calc |
| Transportation | [transportation.md](industry/transportation.md) | Transport profiles, JF routes, cost calc |
| Hauling Runs | [hauling-runs.md](industry/hauling-runs.md) | Hub-to-hub arbitrage, run planning, fill tracking |

## Infrastructure

| Feature | Doc | Summary |
|---------|-----|---------|
| E2E Testing | [e2e-testing.md](infrastructure/e2e-testing.md) | Mock ESI, Playwright, test users |
| Railway Deployment | [railway-deployment.md](infrastructure/railway-deployment.md) | PostgreSQL, backend, frontend deployment |
| DB Performance Optimizations | [db-performance-optimizations.md](infrastructure/db-performance-optimizations.md) | Strategic indexing, SQL functions, N+1 elimination |

## Agent Documentation

Agent docs live in [`docs/agents/`](../agents/):

| Agent | Doc | Summary |
|-------|-----|---------|
| DBA | [dba-agent.md](../agents/dba-agent.md) | Schema research, migration review, query optimization |
| Docs | [docs-agent.md](../agents/docs-agent.md) | Documentation maintenance, index updates, organization |
| SDET | [sdet-agent.md](../agents/sdet-agent.md) | E2E test creation, mock ESI updates, Playwright maintenance |
