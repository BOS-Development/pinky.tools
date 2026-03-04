# Design Token System — Neocom Dark

## Overview
Centralized design token system for the EVE Industry Tool frontend. All visual styling should reference these tokens rather than hardcoded values.

## Status
- **Phase**: Active
- **Token file**: `frontend/app/globals.css` (CSS custom properties)
- **Tailwind integration**: `frontend/tailwind.config.ts`

## Colors (18 tokens)

### Backgrounds (3-layer depth)
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--color-bg-void` | `#0a0a0f` | `bg-background-void` | Page background, deepest layer |
| `--color-bg-panel` | `#12141a` | `bg-background-panel` | Cards, panels, containers |
| `--color-surface-elevated` | `#1a1d24` | `bg-background-elevated` | Elevated elements, modals |

### Primary Cyan (2 variants)
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--color-primary-cyan` | `#00d4ff` | `text-primary` / `bg-primary` | Primary actions, active states |
| `--color-cyan-muted` | `#00a8cc` | `text-primary-muted` / `bg-primary-muted` | Secondary cyan, hover states |

### Text Hierarchy (4 levels)
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--color-text-emphasis` | `#e0e4eb` | `text-text-emphasis` | Headers, important text |
| `--color-text-primary` | `#c0c5d0` | `text-text-primary` | Body text, main content |
| `--color-text-secondary` | `#808898` | `text-text-secondary` | Secondary labels, descriptions |
| `--color-text-muted` | `#64748b` | `text-text-muted` | Tertiary text, timestamps |

### Borders (2 opacities)
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--color-border-dim` | `rgba(0, 212, 255, 0.10)` | `border-border-dim` | Subtle dividers, card borders |
| `--color-border-active` | `rgba(0, 212, 255, 0.30)` | `border-border-active` | Active/focused borders |

### Semantic Status (4 colors)
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--color-manufacturing-amber` | `#fbbf24` | `text-amber-manufacturing` | Manufacturing, warnings |
| `--color-science-blue` | `#60a5fa` | `text-blue-science` | Science, research, info |
| `--color-success-teal` | `#2dd4bf` | `text-teal-success` | Success, positive values |
| `--color-danger-rose` | `#f43f5e` | `text-rose-danger` | Errors, negative values |

### Interactive Tints (4 states)
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--color-interactive-hover` | `rgba(0, 212, 255, 0.08)` | `bg-interactive-hover` | Hover background tint |
| `--color-interactive-active` | `rgba(0, 212, 255, 0.15)` | `bg-interactive-active` | Pressed/active background |
| `--color-interactive-selected` | `rgba(0, 212, 255, 0.12)` | `bg-interactive-selected` | Selected row/item background |
| `--color-interactive-disabled` | `rgba(100, 116, 139, 0.30)` | `bg-interactive-disabled` | Disabled state overlay |

### Neutral Overlays (3 opacities)
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--color-overlay-subtle` | `rgba(148, 163, 184, 0.10)` | `bg-overlay-subtle` | Row backgrounds, subtle fills |
| `--color-overlay-medium` | `rgba(148, 163, 184, 0.15)` | `bg-overlay-medium` | Hover surfaces |
| `--color-overlay-strong` | `rgba(148, 163, 184, 0.20)` | `bg-overlay-strong` | Active surfaces, dividers |

## Typography

### Font Families (3)
| Token | Stack | Tailwind | Usage |
|-------|-------|----------|-------|
| `--font-exo2` | Exo 2 | `font-display` | Display headings, page titles |
| `--font-geist-sans` | Geist | `font-sans` | Body text, UI labels |
| `--font-jetbrains-mono` | JetBrains Mono | `font-mono` | Numbers, ISK values, code |

### Size Scale (8 steps)
| Token | Value | Px | Usage |
|-------|-------|-----|-------|
| `--font-size-xs` | `0.625rem` | 10px | Fine print, badges |
| `--font-size-sm` | `0.75rem` | 12px | Small labels, captions |
| `--font-size-base` | `0.875rem` | 14px | Default body text |
| `--font-size-md` | `1rem` | 16px | Prominent text, subheaders |
| `--font-size-lg` | `1.125rem` | 18px | Section headers |
| `--font-size-xl` | `1.25rem` | 20px | Page sub-titles |
| `--font-size-2xl` | `1.5rem` | 24px | Page titles |
| `--font-size-3xl` | `1.875rem` | 30px | Hero text |

### Weights (4)
| Token | Value | Usage |
|-------|-------|-------|
| `--font-weight-normal` | 400 | Body text |
| `--font-weight-medium` | 500 | Labels, UI elements |
| `--font-weight-semibold` | 600 | Subheaders, emphasis |
| `--font-weight-bold` | 700 | Display headings |

### Line Heights (3)
| Token | Value | Usage |
|-------|-------|-------|
| `--line-height-tight` | 1.2 | Headings, compact layouts |
| `--line-height-normal` | 1.5 | Body text, default |
| `--line-height-relaxed` | 1.8 | Long-form content |

### Letter Spacing (5)
| Token | Value | Usage |
|-------|-------|-------|
| `--letter-spacing-tight` | -0.01em | Large display text |
| `--letter-spacing-normal` | 0em | Default |
| `--letter-spacing-wide` | 0.02em | Headings (h1–h3) |
| `--letter-spacing-wider` | 0.05em | Subheaders |
| `--letter-spacing-widest` | 0.08em | Uppercase labels (h4–h6) |

## Spacing

### Scale (12 steps, 4px base)
| Token | Value | Px |
|-------|-------|-----|
| `--space-1` | `0.25rem` | 4px |
| `--space-2` | `0.5rem` | 8px |
| `--space-3` | `0.75rem` | 12px |
| `--space-4` | `1rem` | 16px |
| `--space-5` | `1.25rem` | 20px |
| `--space-6` | `1.5rem` | 24px |
| `--space-8` | `2rem` | 32px |
| `--space-10` | `2.5rem` | 40px |
| `--space-12` | `3rem` | 48px |
| `--space-16` | `4rem` | 64px |
| `--space-20` | `5rem` | 80px |
| `--space-24` | `6rem` | 96px |

### Layout Tokens
| Token | Value | Usage |
|-------|-------|-------|
| `--layout-page-px` | `1.5rem` | Page horizontal padding |
| `--layout-card-padding` | `1.25rem` | Card internal padding |
| `--layout-navbar-height` | `4rem` | Navbar height (h-16) |
| `--layout-section-gap` | `1.5rem` | Gap between page sections |
| `--layout-sidebar-width` | `16rem` | Sidebar width |

## Borders & Radii

### Border Width
- Always `1px` — no other border widths in the system

### Border Radii
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--radius-none` | `0` | `rounded-none` | Sharp corners (tables, inline elements) |
| `--radius-default` | `0.125rem` (2px) | `rounded` | Default (cards, buttons, inputs) |
| `--radius-full` | `9999px` | `rounded-full` | Pills, avatars, badges |

## Shadows & Glows

### Cyan Glow Tiers
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--glow-cyan-sm` | `0 0 8px rgba(0, 212, 255, 0.25)` | `shadow-glow-sm` | Subtle element glow |
| `--glow-cyan-md` | `0 0 12px rgba(0, 212, 255, 0.35)` | `shadow-glow-md` | Medium glow (buttons, hover) |
| `--glow-cyan-lg` | `0 0 20px ...` | `shadow-glow-lg` | Large glow (hero, focused) |
| `--glow-text-heading` | `0 0 12px rgba(0, 212, 255, 0.3)` | — | H1/H2 heading text-shadow |

**No neutral shadows** — use only cyan-tinted glows.

## Migration Guide

### Replacing Hardcoded Hex Values

| Hardcoded Value | Replace With (Tailwind) |
|----------------|------------------------|
| `text-[#e2e8f0]` | `text-text-emphasis` |
| `text-[#cbd5e1]` | `text-text-primary` |
| `text-[#94a3b8]` | `text-text-secondary` |
| `text-[#64748b]` | `text-text-muted` |
| `text-[#00d4ff]` | `text-primary` |
| `bg-[#0f1219]` / `bg-[#0a0a0f]` | `bg-background-void` |
| `bg-[#12151f]` / `bg-[#12141a]` | `bg-background-panel` |
| `bg-[#1a1d24]` | `bg-background-elevated` |
| `text-[#10b981]` / `text-[#2dd4bf]` | `text-teal-success` |
| `text-[#ef4444]` / `text-[#f43f5e]` | `text-rose-danger` |
| `text-[#f59e0b]` / `text-[#fbbf24]` | `text-amber-manufacturing` |
| `text-[#60a5fa]` | `text-blue-science` |
| `border-[rgba(148,163,184,0.1)]` | `border-overlay-subtle` |
| `bg-[rgba(148,163,184,0.1)]` | `bg-overlay-subtle` |
| `bg-[rgba(0,212,255,0.08)]` | `bg-interactive-hover` |

### Key Decisions
- No harsh white text — `text-emphasis` (#e0e4eb) is the brightest
- Cyan-first visual language — borders, glows, and interactive states use cyan tints
- 2px max border radius — angular/industrial aesthetic
- Monospace for all numbers — use `font-mono` for ISK values, quantities

## File Paths
| File | Purpose |
|------|---------|
| `frontend/app/globals.css` | CSS custom property definitions |
| `frontend/tailwind.config.ts` | Tailwind theme extensions |
| `frontend/packages/utils/formatting.ts` | Color utility functions |
