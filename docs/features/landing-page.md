# Landing Page

## Overview

Server-side rendered landing page that serves as both a marketing page (unauthenticated) and dashboard home (authenticated).

**File:** `frontend/packages/pages/index.tsx`

## Structure

### Hero Section (Full Viewport)
- **Layout:** 100vh height with flexbox column, purple gradient background (`#667eea` → `#764ba2`)
- **Header:** Ragnarok titan image (EVE type ID 23773) with drop shadow
- **Headline:** "Master Your EVE Online Assets"
- **Subheadline:** Feature description
- **CTAs:** Conditional based on authentication state

### Authenticated Users
**Navigation buttons:**
- Characters (primary) → `/characters`
- View Assets (outlined) → `/inventory`
- Manage Stockpiles (outlined) → `/stockpiles`

**Live metrics cards:**
- **Total Asset Value** - Blue gradient (#1e3a8a → #3b82f6), shows ISK total or "Dirty Poor" if 0
- **Stockpile Deficit** - Red gradient if deficits (#991b1b → #dc2626), green if none (#166534 → #22c55e)

### Unauthenticated Users
**Single CTA:**
- "Sign In with EVE Online" → `/api/auth/signin`

### Footer
EVE Online disclaimer (positioned at bottom of viewport)

## Technical Implementation

### Server-Side Rendering
```tsx
const session = await getServerSession(authOptions);
const isAuthenticated = !!session;
```

### Asset Metrics Fetching
- Fetches from `${BACKEND_URL}v1/assets/summary` on server-side
- Backend calculates totals via SQL aggregation (efficient)
- Returns only `{ totalValue, totalDeficit }` (minimal payload)
- Silently fails to 0 values on error
- Uses `cache: 'no-store'` for fresh data

### Responsive Design
- Metrics grid: 1 column mobile, 2 columns desktop
- Button stack: vertical mobile, horizontal tablet+
- Single viewport height (no scroll)

### Styling
- **Components:** MUI Box, Container, Typography, Button, Card, Stack
- **Hover effects:** `translateY(-4px)` with `boxShadow: 6` on metric cards
- **Typography:** h1 for headline, h5 for subheadline, h4 for metric values
- **Spacing:** Compact to fit all content in 100vh

## Key Features

- ✅ No loading states (server-rendered)
- ✅ Real-time asset metrics
- ✅ Conditional UI based on auth
- ✅ Responsive mobile → desktop
- ✅ EVE-themed imagery (Ragnarok titan)
- ✅ Fits single viewport (no scroll)
