# Design Tokens

## Status
Implemented

## Overview

The design token system provides a comprehensive, centralized color and typography palette for the Neocom Dark interface. Tokens are defined as CSS custom properties in `frontend/app/globals.css` and mapped to Tailwind classes in `frontend/tailwind.config.ts`. This approach eliminates hardcoded colors, enables consistent styling across the application, and facilitates theme evolution.

The system enforces a **zero-hardcoded-colors policy**: all colors are token references, never literal hex values in components or inline styles.

## Key Decisions

- **CSS Custom Properties for Inline Styles**: CSS variables (`var(--color-*)`) can be used inline when Tailwind classes are insufficient. This is the only exception to strict Tailwind-only styling.
- **Two-Layer Implementation**: CSS variables in the `:root` scope + Tailwind color extensions allow both dynamic runtime styles and static build-time optimization.
- **Tailwind Opacity Modifiers**: For color variants (e.g., hover, active states), use Tailwind's opacity syntax (`/50`) rather than duplicating tokens. Reduces token count while maintaining clarity.
- **Industrial Aesthetic**: Minimal border radius (0.125rem default), tight letter spacing, and angular components reflect EVE Online's neon-industrial visual language.
- **Three-Layer Background Depth System**: Void (deepest), Panel (middle), Elevated (surface) provide visual hierarchy without gradients.
- **HSL Semantic Tokens for shadcn/ui**: Complementary HSL variables enable shadcn/ui components to blend naturally with custom Neocom styling.

## Color Tokens

### Background Palette (Three-Layer Depth)
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--color-bg-void` | `#0a0a0f` | `bg-background-void` | Page background, deepest layer |
| `--color-bg-panel` | `#12141a` | `bg-background-panel` | Cards, panels, containers |
| `--color-surface-elevated` | `#1a1d24` | `bg-background-elevated` | Elevated elements, modals |

### Border Palette
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--color-border-dim` | `rgba(0, 212, 255, 0.10)` | `border-border-dim` | Subtle dividers, card borders |
| `--color-border-active` | `rgba(0, 212, 255, 0.30)` | `border-border-active` | Active/focused borders |

### Primary Interactive (Cyan)
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--color-primary-cyan` | `#00d4ff` | `text-primary`, `bg-primary` | Primary actions, active states, headings |
| `--color-cyan-muted` | `#00a8cc` | `text-primary-muted` | Secondary cyan, link hover states |

### Interactive States
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--color-interactive-hover` | `rgba(0, 212, 255, 0.08)` | `bg-interactive-hover` | Hover background tint |
| `--color-interactive-active` | `rgba(0, 212, 255, 0.15)` | `bg-interactive-active` | Pressed/active background |
| `--color-interactive-selected` | `rgba(0, 212, 255, 0.12)` | `bg-interactive-selected` | Selected row/item background |
| `--color-interactive-disabled` | `rgba(100, 116, 139, 0.30)` | `bg-interactive-disabled` | Disabled state overlay |

### Semantic Status Colors
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--color-manufacturing-amber` | `#fbbf24` | `text-amber-manufacturing`, `bg-amber-manufacturing` | Manufacturing jobs, warnings |
| `--color-science-blue` | `#60a5fa` | `text-blue-science` | Science jobs, research, info |
| `--color-success-teal` | `#2dd4bf` | `text-teal-success` | Success, complete, positive values |
| `--color-danger-rose` | `#f43f5e` | `text-rose-danger` | Danger, errors, negative values |

### Status Tints (Subtle Backgrounds)
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--color-success-tint` | `rgba(45, 212, 191, 0.10)` | `bg-status-success-tint` | Success message background |
| `--color-warning-tint` | `rgba(251, 191, 36, 0.10)` | `bg-status-warning-tint` | Warning message background |
| `--color-error-tint` | `rgba(244, 63, 94, 0.10)` | `bg-status-error-tint` | Error message background |
| `--color-info-tint` | `rgba(0, 212, 255, 0.10)` | `bg-status-info-tint` | Info message background |
| `--color-neutral-tint` | `rgba(148, 163, 184, 0.08)` | `bg-status-neutral-tint` | Neutral message background |

### Category Colors (Data Visualization)
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--color-category-violet` | `#8b5cf6` | `text-category-violet` | Primary category color |
| `--color-category-pink` | `#ec4899` | `text-category-pink` | Secondary category color |
| `--color-category-orange` | `#f97316` | `text-category-orange` | Tertiary category color |
| `--color-category-teal` | `#06b6d4` | `text-category-teal` | Quaternary category color |
| `--color-category-slate` | `#94a3b8` | `text-category-slate` | Neutral category color |

### Accent Blue (Secondary Action)
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--color-accent-blue` | `#3b82f6` | `bg-accent-blue`, `text-accent-blue` | Secondary action color |
| `--color-accent-blue-hover` | `#2563eb` | `bg-accent-blue-hover` | Accent hover state |
| `--color-accent-blue-muted` | `rgba(59, 130, 246, 0.15)` | `bg-accent-blue-muted` | Subtle accent tint |

### Semantic Backgrounds (Activity-Specific)
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--color-bg-manufacturing` | `#422006` | `bg-bg-manufacturing` | Dark amber background for manufacturing |
| `--color-bg-science` | `#1e3a5f` | `bg-bg-science` | Dark blue background for science |
| `--color-bg-warning` | `#2d2000` | `bg-bg-warning` | Dark amber background for warnings |

### Text Hierarchy
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--color-text-emphasis` | `#e0e4eb` | `text-text-emphasis` | Brightest text, headers, important content |
| `--color-text-primary` | `#c0c5d0` | `text-text-primary` | Main body text, default |
| `--color-text-secondary` | `#808898` | `text-text-secondary` | Secondary labels, descriptions |
| `--color-text-muted` | `#64748b` | `text-text-muted` | Tertiary text, timestamps, metadata |

### Neutral Overlays
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--color-overlay-subtle` | `rgba(148, 163, 184, 0.10)` | `bg-overlay-subtle` | Row backgrounds, subtle fills |
| `--color-overlay-medium` | `rgba(148, 163, 184, 0.15)` | `bg-overlay-medium` | Hover surfaces, mid-emphasis |
| `--color-overlay-strong` | `rgba(148, 163, 184, 0.20)` | `bg-overlay-strong` | Active surfaces, strong emphasis |

## Glow & Shadow Effects

### Cyan Glow Tiers
| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--glow-cyan-sm` | `0 0 8px rgba(0, 212, 255, 0.25)` | `shadow-glow-sm` | Subtle element glow |
| `--glow-cyan-md` | `0 0 12px rgba(0, 212, 255, 0.35)` | `shadow-glow-md` | Medium glow (buttons, hover) |
| `--glow-cyan-lg` | `0 0 20px rgba(0, 212, 255, 0.3), 0 0 40px rgba(0, 212, 255, 0.1)` | `shadow-glow-lg` | Large glow (hero, focused elements) |
| `--glow-text-heading` | `0 0 12px rgba(0, 212, 255, 0.3)` | — | H1/H2 heading text-shadow |

**Note**: No neutral shadows in the system — use only cyan-tinted glows.

## Typography Tokens

### Font Sizes (8-step scale)
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

### Font Weights
| Token | Value | Usage |
|-------|-------|-------|
| `--font-weight-normal` | 400 | Body text |
| `--font-weight-medium` | 500 | Labels, UI elements |
| `--font-weight-semibold` | 600 | Subheaders, emphasis |
| `--font-weight-bold` | 700 | Display headings |

### Line Heights
| Token | Value | Usage |
|-------|-------|-------|
| `--line-height-tight` | 1.2 | Headings, compact layouts |
| `--line-height-normal` | 1.5 | Body text, default |
| `--line-height-relaxed` | 1.8 | Long-form content |

### Letter Spacing
| Token | Value | Usage |
|-------|-------|-------|
| `--letter-spacing-tight` | -0.01em | Large display text |
| `--letter-spacing-normal` | 0em | Default |
| `--letter-spacing-wide` | 0.02em | Headings (h1–h3) |
| `--letter-spacing-wider` | 0.05em | Subheaders |
| `--letter-spacing-widest` | 0.08em | Uppercase labels (h4–h6) |

## Spacing Scale (12 steps, 4px base)

| Token | Value | Px | Token | Value | Px |
|-------|-------|-----|-------|-------|-----|
| `--space-1` | `0.25rem` | 4px | `--space-12` | `3rem` | 48px |
| `--space-2` | `0.5rem` | 8px | `--space-16` | `4rem` | 64px |
| `--space-3` | `0.75rem` | 12px | `--space-20` | `5rem` | 80px |
| `--space-4` | `1rem` | 16px | `--space-24` | `6rem` | 96px |
| `--space-5` | `1.25rem` | 20px | — | — | — |
| `--space-6` | `1.5rem` | 24px | — | — | — |
| `--space-8` | `2rem` | 32px | — | — | — |
| `--space-10` | `2.5rem` | 40px | — | — | — |

## Layout Constants

| Token | Value | Usage |
|-------|-------|-------|
| `--layout-page-px` | `1.5rem` | Page horizontal padding |
| `--layout-card-padding` | `1.25rem` | Card internal padding |
| `--layout-navbar-height` | `4rem` | Navigation bar height |
| `--layout-section-gap` | `1.5rem` | Gap between page sections |
| `--layout-sidebar-width` | `16rem` | Sidebar width |

## Border Radii

| Token | Value | Tailwind | Usage |
|-------|-------|----------|-------|
| `--radius-none` | `0` | `rounded-none` | Sharp corners (tables, inline elements) |
| `--radius-default` | `0.125rem` (2px) | `rounded` | Default (cards, buttons, inputs) |
| `--radius-full` | `9999px` | `rounded-full` | Pills, avatars, badges |

## Usage Examples

### Background Layering
```jsx
// Page container (deepest)
<div className="bg-background-void">
  {/* Panel card (middle) */}
  <div className="bg-background-panel border border-border-dim rounded">
    {/* Elevated content (surface) */}
    <div className="bg-background-elevated p-4">
      Content here
    </div>
  </div>
</div>
```

### Interactive States
```jsx
<button className="bg-background-panel hover:bg-interactive-hover active:bg-interactive-active disabled:bg-interactive-disabled">
  Click me
</button>
```

### Semantic Status
```jsx
<div className="bg-status-success-tint text-teal-success">
  Success message
</div>

<div className="bg-status-error-tint text-rose-danger">
  Error message
</div>
```

### Text Hierarchy
```jsx
<h1 className="text-primary text-3xl font-bold">Main Heading</h1>
<p className="text-text-primary">Primary body text</p>
<p className="text-text-secondary">Secondary information</p>
<p className="text-text-muted">Tertiary metadata</p>
```

### Glow Effects
```jsx
<h2 className="text-primary text-2xl shadow-glow-text">
  Neon Heading
</h2>

<div className="border border-border-active shadow-glow-md">
  Highlighted card
</div>
```

### Inline CSS Variables
Only use CSS variables directly for styles Tailwind doesn't cover:
```jsx
<div style={{ boxShadow: 'var(--glow-cyan-lg)' }}>
  Custom glow effect
</div>
```

## Tailwind Mapping

All design tokens are mapped as Tailwind utilities in `frontend/tailwind.config.ts`. Example structure:

```typescript
colors: {
  background: { void, panel, elevated },
  border: { dim, active },
  primary: { DEFAULT, muted },
  amber: { manufacturing },
  blue: { science },
  teal: { success },
  rose: { danger },
  status: { success-tint, warning-tint, error-tint, info-tint, neutral-tint },
  category: { violet, pink, orange, teal, slate },
  accent-blue: { DEFAULT, hover, muted },
  bg-manufacturing, bg-science, bg-warning,
  text: { primary, secondary, muted, emphasis },
  interactive: { hover, active, selected, disabled },
  overlay: { subtle, medium, strong },
}

boxShadow: {
  "glow-sm", "glow-md", "glow-lg", "glow-text"
}
```

## Architecture Notes

### Tailwind v4 Integration
The design token system coexists with **Tailwind CSS v4** in `frontend/app/globals.css`:
- Layer order: `base < theme < utilities`
- Tailwind's theme layer provides utilities
- Custom tokens are CSS variables in `:root`
- Tailwind extends custom theme with `@config` reference

### shadcn/ui Semantic Tokens
HSL semantic variables (e.g., `--foreground: 221 15% 78%`) enable shadcn/ui components to use Neocom Dark colors. These are complementary to the primary design tokens and do not replace them.

### Empty State Component
A consistent `.empty-state` class provides uniform styling for all empty lists across the application:
```css
.empty-state {
  background: var(--color-bg-panel);
  border: 1px solid var(--color-border-dim);
  border-radius: var(--radius);
  padding: 3rem 2rem;
  text-align: center;
}
```

## Migration Guide (MUI → Tailwind)

When migrating MUI components to Tailwind:

1. **Replace MUI `sx` props** with Tailwind classes using token names.
   ```jsx
   // Before (MUI)
   <Box sx={{ backgroundColor: '#12141a', color: '#c0c5d0' }}>

   // After (Tailwind)
   <div className="bg-background-panel text-text-primary">
   ```

2. **Use Tailwind opacity modifiers** for color variants instead of creating new tokens.
   ```jsx
   {/* Hover state with opacity modifier */}
   <button className="hover:bg-primary/20">
   ```

3. **Reference glow tokens** for emphasis and highlights.
   ```jsx
   <h2 className="text-primary shadow-glow-text">
   ```

4. **Preserve CSS variable refs** only when Tailwind classes are insufficient.
   ```jsx
   {/* Rare exception: custom inline shadow */}
   <div style={{ boxShadow: 'var(--glow-cyan-lg)' }}>
   ```

## File Paths

- **Token Definitions**: `/home/benjamin/mordecai/design-tokens-phase2/frontend/app/globals.css` (lines 43–214)
- **Tailwind Configuration**: `/home/benjamin/mordecai/design-tokens-phase2/frontend/tailwind.config.ts` (lines 13–106)

## Related Documentation
- See [UI Migration Phase 0](./ui-migration-phase0.md) for Tailwind integration strategy
- See [Railway Deployment](./railway-deployment.md) for production build verification
