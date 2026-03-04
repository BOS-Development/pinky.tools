# Landing Page

## Overview

Server-side rendered landing page that serves as both a marketing page (unauthenticated) and dashboard home (authenticated).

**File:** `frontend/packages/pages/index.tsx`

## Status
Implemented

## Structure

### Hero Section
- **Layout:** `min-h-screen` (allows below-fold content) with flexbox column, purple gradient background (`#667eea` → `#764ba2`)
- **Image:** Ragnarok titan (EVE type ID 23773) at 1024px for retina display support, with drop shadow
- **Headline:** "Master Your EVE Online Assets"
- **Subheadline:** Feature description
- **CTAs:** Conditional based on authentication state

### Authenticated Users

**Hero CTAs (buttons):**
- Characters (primary) → `/characters`
- View Assets (outlined) → `/inventory`
- Manage Stockpiles (outlined) → `/stockpiles`
- ~~Open Dashboard~~ (removed — already on dashboard)

**Live metrics cards:**
- **Total Asset Value** - Blue gradient (#1e3a8a → #3b82f6), displays ISK value or "0 ISK" if none
- **Stockpile Deficit** - Red gradient if deficits (#991b1b → #dc2626), green if none (#166534 → #22c55e), displays quantity or "0" if none
- **Active Industry Jobs** - Shows count from `AssetsSummary.ActiveJobs`

**Quick Access navigation grid (below hero):**
Eight section links for rapid navigation:
- Characters
- Inventory
- Stockpiles
- Industry Jobs
- Reactions
- Transportation
- Contacts
- Settings

### Unauthenticated Users
**Single CTA:**
- "Sign In with EVE Online" → `/api/auth/signin`

### Footer
EVE Online disclaimer (positioned below hero section)

## Technical Implementation

### Server-Side Rendering
```tsx
const session = await getServerSession(authOptions);
const isAuthenticated = !!session;
```

### Asset Metrics Fetching
- Fetches from `${BACKEND_URL}v1/assets/summary` on server-side
- Backend response includes: `totalValue`, `totalDeficit`, `activeJobs`
- Returns aggregated totals via SQL (efficient)
- Silently fails to 0 values on error
- Uses `cache: 'no-store'` for fresh data

### Backend AssetsSummary Response
```go
type AssetsSummary struct {
  TotalValue   int64 `json:"totalValue"`   // ISK, aggregated from character + corp assets
  TotalDeficit int64 `json:"totalDeficit"` // ISK, sum of all stockpile deficits
  ActiveJobs   int   `json:"activeJobs"`   // Count of active industry jobs from esi_industry_jobs
}
```

### Responsive Design
- Metrics grid: 1 column mobile, 3 columns desktop (with job count)
- Button stack: vertical mobile, horizontal tablet+
- Quick Access grid: 2 columns mobile, 4 columns desktop
- Content scrolls below fold (min-h-screen allows overflow)

### Styling
- **Components:** MUI Box, Container, Typography, Button, Card, Stack, Grid
- **Hover effects:** `translateY(-4px)` with `boxShadow: 6` on metric cards
- **Typography:** h1 for headline, h5 for subheadline, h4 for metric values
- **Spacing:** Allows content to extend below viewport
- **Image:** 1024px source for retina-ready display

## Key Features

- ✅ No loading states (server-rendered)
- ✅ Real-time asset metrics with industry job count
- ✅ Conditional UI based on auth
- ✅ Responsive mobile → desktop
- ✅ EVE-themed imagery (Ragnarok titan) at high resolution
- ✅ Quick Access navigation for authenticated users
- ✅ Scrollable below-fold content
- ✅ Clean zero-state display ("0 ISK", "0")
