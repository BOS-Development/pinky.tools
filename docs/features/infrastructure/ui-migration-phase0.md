# UI Migration — Phase 0: Foundation

## Status
Complete — branch `feature/116-ui-migration-phase0`

## Overview
Phase 0 establishes the Tailwind CSS v4 + shadcn/ui infrastructure alongside the existing MUI stack. No pages change visually. This foundation enables the gradual component-by-component migration from MUI to shadcn/ui.

## What Was Added

### Tailwind CSS v4
- `tailwindcss` + `@tailwindcss/postcss` installed
- `postcss.config.mjs` — activates `@tailwindcss/postcss` plugin
- `tailwind.config.ts` — content paths, CSS variable–based theme tokens, font families, shadcn/ui semantic token mappings
- **Preflight disabled at CSS layer level** to prevent MUI base reset conflicts — `globals.css` imports only `tailwindcss/theme` and `tailwindcss/utilities` layers

### Neocom Dark Design Tokens
All palette values defined as CSS custom properties in `app/globals.css`:

| Token | CSS Variable | Value |
|-------|-------------|-------|
| Background Void | `--color-bg-void` | `#080c14` |
| Background Panel | `--color-bg-panel` | `#0f1420` |
| Surface Elevated | `--color-surface-elevated` | `#1a1f2e` |
| Border Dim | `--color-border-dim` | `rgba(0,212,255,0.06)` |
| Border Active | `--color-border-active` | `rgba(0,212,255,0.20)` |
| Primary Cyan | `--color-primary-cyan` | `#00d4ff` |
| Cyan Muted | `--color-cyan-muted` | `#00a8cc` |
| Manufacturing Amber | `--color-manufacturing-amber` | `#fbbf24` |
| Science Blue | `--color-science-blue` | `#60a5fa` |
| Success Teal | `--color-success-teal` | `#2dd4bf` |
| Danger Rose | `--color-danger-rose` | `#f43f5e` |
| Text Primary | `--color-text-primary` | `#e2e8f0` |
| Text Secondary | `--color-text-secondary` | `#94a3b8` |
| Text Muted | `--color-text-muted` | `#64748b` |

shadcn/ui HSL semantic tokens (`--background`, `--foreground`, `--primary`, `--card`, etc.) are also defined, mapped to the Neocom Dark palette.

### shadcn/ui
- `components.json` — style: new-york, RSC: true, CSS variables enabled, base color: neutral
- `lib/utils.ts` — `cn()` helper (clsx + tailwind-merge)
- Core primitives generated and Neocom-themed in `components/ui/`:
  - `button.tsx` — includes custom `neocom` variant (cyan border + glow)
  - `card.tsx`, `input.tsx`, `badge.tsx`, `dialog.tsx`, `select.tsx`, `table.tsx`, `tabs.tsx`, `tooltip.tsx`, `alert.tsx`

### Typography
- `Exo_2` font added to `app/layout.tsx` via `next/font/google` as `--font-exo2`
- Headers (`h1`–`h6`) use `var(--font-exo2)` via CSS
- Body: Geist Sans (`--font-geist-sans`) — already present
- Monospace/Data: JetBrains Mono (`--font-jetbrains-mono`) — already present

## CSS Specificity Strategy

MUI uses Emotion (CSS-in-JS). Tailwind v4 uses CSS cascade layers (`theme`, `utilities`). By importing only `tailwindcss/theme` and `tailwindcss/utilities` (skipping `preflight`), MUI's base styling is preserved. Tailwind utility classes can coexist without conflict because:
1. No base reset overrides MUI's CssBaseline
2. Tailwind utilities sit in the `utilities` layer — lower priority than Emotion's inline styles by default
3. To override MUI with Tailwind on a specific element, use CSS custom properties in `style={}` props

## File Structure

```
frontend/
├── app/
│   ├── globals.css          # Tailwind import + all CSS variables
│   └── layout.tsx           # Added Exo_2 font
├── components/
│   └── ui/                  # All shadcn/ui primitives (Neocom-themed)
│       ├── alert.tsx
│       ├── badge.tsx
│       ├── button.tsx       # Includes `neocom` variant
│       ├── card.tsx
│       ├── dialog.tsx
│       ├── input.tsx
│       ├── select.tsx
│       ├── table.tsx
│       ├── tabs.tsx
│       └── tooltip.tsx
├── lib/
│   └── utils.ts             # cn() helper
├── components.json          # shadcn/ui config
├── tailwind.config.ts       # Theme + content paths
└── postcss.config.mjs       # @tailwindcss/postcss enabled
```

## Key Decisions

1. **Tailwind v4** chosen over v3 — aligns with `@tailwindcss/postcss` already referenced in `packages/pages`, modern architecture
2. **No preflight** — preserves MUI CssBaseline; only `theme` and `utilities` layers imported
3. **New-york style** for shadcn/ui — more refined, suits the EVE-style dark aesthetic
4. **Components go in `frontend/components/ui/`** — separate from existing `packages/components/` MUI workspace to avoid conflicts
5. **Exo 2 headers** — condensed, technical font fits the space/industry aesthetic
6. **CSS variables for all tokens** — enables runtime theming and matches shadcn/ui's CSS variable architecture

## Next Phases

- Phase 1: Migrate leaf components (buttons, inputs, chips → shadcn primitives)
- Phase 2: Migrate layout components (cards, dialogs, tables)
- Phase 3: Migrate page-level layouts and navigation
