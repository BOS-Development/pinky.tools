# Buy Orders & Demand Tracking (Phase 5)

Phase 5 of the Contact & Marketplace system enables buyers to create buy orders for items they want to purchase, and sellers to view demand from their contacts.

## Quick Links

- **[Feature Overview](#overview)** - What Phase 5 does
- **[API Documentation](API.md)** - RESTful API reference
- **[Quick Start Guide](QUICK_START.md)** - Get started in 5 minutes
- **[Testing Guide](TESTING.md)** - Test suite documentation
- **[Implementation Summary](IMPLEMENTATION.md)** - Technical implementation details

## Overview

### What Are Buy Orders?

Buy orders let users express demand for items they want to purchase. When a buyer creates a buy order, they specify:
- **Item** (autocomplete with icons)
- **Quantity desired**
- **Maximum price per unit** they're willing to pay
- **Optional notes**

### Demand Tracking

Sellers can view buy orders from their contacts (those who granted them `for_sale_browse` permission), showing:
- **Aggregated demand** per item type
- **Highest price offered** for each item
- **Total potential revenue**
- **Number of orders** per item
- **Individual order details**

## Key Features

### 🔍 Item Autocomplete
- **Real-time search** with 300ms debounce
- **EVE item icons** in dropdown and selected value
- **Relevance-based sorting**:
  1. Exact matches first
  2. Items starting with query
  3. Items containing query
- **Minimum 2 characters** to search
- **Up to 20 results** displayed

### 📊 Demand Aggregation
- Group orders by item type
- Calculate total quantity wanted
- Show highest price offered
- Calculate potential revenue
- Count number of orders

### 🔐 Permission-Based
- Only contacts with `for_sale_browse` permission see your buy orders
- Granular control over who can see your demand
- Bidirectional permission system

### ✏️ Full CRUD Operations
- **Create** buy orders with autocomplete
- **Read** your orders and demand from contacts
- **Update** quantity and price
- **Delete** (soft delete - sets `is_active = false`)

## User Flows

### Buyer Flow
1. Go to Marketplace → "My Buy Orders" tab
2. Click "Create Buy Order"
3. Search for item using autocomplete (see icon + name)
4. Enter quantity desired and max price per unit
5. Optionally add notes
6. Click "Create"
7. Order appears in list, sellers can now see it

### Seller Flow
1. Go to Marketplace → "Demand" tab
2. View aggregated demand:
   - See total quantity wanted per item
   - See highest price offered
   - See potential revenue if you sell all
3. View individual orders for details
4. Plan inventory based on demand

## Database Schema

### buy_orders Table

```sql
CREATE TABLE buy_orders (
    id BIGSERIAL PRIMARY KEY,
    buyer_user_id BIGINT NOT NULL REFERENCES users(id),
    type_id BIGINT NOT NULL REFERENCES asset_item_types(type_id),
    quantity_desired BIGINT NOT NULL,
    max_price_per_unit BIGINT NOT NULL,
    notes TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT buy_order_positive_quantity CHECK (quantity_desired > 0),
    CONSTRAINT buy_order_positive_price CHECK (max_price_per_unit >= 0)
);
```

**Indexes:**
- `idx_buy_orders_buyer` - ON (buyer_user_id)
- `idx_buy_orders_type` - ON (type_id)
- `idx_buy_orders_active` - ON (is_active)

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/buy-orders` | Get my buy orders |
| POST | `/v1/buy-orders` | Create a buy order |
| PUT | `/v1/buy-orders/{id}` | Update a buy order |
| DELETE | `/v1/buy-orders/{id}` | Delete a buy order |
| GET | `/v1/buy-orders/demand` | Get demand from contacts |
| GET | `/v1/item-types/search?q={query}` | Search for item types |

See [API Documentation](API.md) for detailed endpoint specifications.

## Frontend Components

### BuyOrders Component
**Location:** `/frontend/packages/components/marketplace/BuyOrders.tsx`

Manages buy orders with:
- Item autocomplete with icons
- Create/edit dialog
- Table view with actions
- Validation and error handling

### DemandViewer Component
**Location:** `/frontend/packages/components/marketplace/DemandViewer.tsx`

Shows demand from contacts with:
- Aggregated summary table
- Individual orders detail table
- Search functionality
- Revenue calculations

### EVE Images Utility
**Location:** `/frontend/packages/utils/eveImages.ts`

Provides helper functions for EVE image URLs:
- `getItemIconUrl(typeId, size)`
- `getItemRenderUrl(typeId, size)`
- `getCharacterPortraitUrl(characterId, size)`
- `getCorporationLogoUrl(corporationId, size)`
- `getAllianceLogoUrl(allianceId, size)`

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Frontend                             │
├─────────────────────────────────────────────────────────────┤
│  BuyOrders.tsx          │  DemandViewer.tsx                 │
│  - Create buy order     │  - View aggregated demand         │
│  - Edit buy order       │  - View individual orders         │
│  - Delete buy order     │  - Search demand                  │
│  - List my orders       │                                   │
└─────────────┬───────────────────────────┬───────────────────┘
              │                           │
              │  Next.js API Routes       │
              │                           │
┌─────────────▼───────────────────────────▼───────────────────┐
│                      Go Backend                              │
├─────────────────────────────────────────────────────────────┤
│  Controllers                                                 │
│  - BuyOrders Controller (5 endpoints)                       │
│  - ItemTypes Controller (1 endpoint)                        │
├─────────────────────────────────────────────────────────────┤
│  Repositories                                                │
│  - BuyOrders Repository (6 methods)                         │
│  - ItemTypes Repository (2 new methods)                     │
├─────────────────────────────────────────────────────────────┤
│  Database                                                    │
│  - buy_orders table                                          │
│  - asset_item_types table (existing)                        │
│  - contact_permissions table (for authorization)            │
└─────────────────────────────────────────────────────────────┘
```

## Permission System

Buy orders respect the `for_sale_browse` permission:

```
User A creates buy order → User B (contact with permission) can see it in Demand tab
User A creates buy order → User C (no permission) CANNOT see it
```

### Granting Permission

1. Go to Contacts page
2. Find the contact
3. Toggle "For-Sale Browse" permission to ON
4. Contact can now see your buy orders in their Demand tab

## Testing

Phase 5 includes comprehensive test coverage:

- **8 Repository Tests** - Database operations, queries, edge cases
- **7 Controller Tests** - API endpoints, validation, permissions
- **15 Total Tests** - All passing ✅

See [Testing Guide](TESTING.md) for details.

## Next Steps

- **Phase 6: Sales Metrics Dashboard** - Analytics and charts for sellers
- Match buy orders to for-sale items automatically
- Notifications when demand matches inventory
- Price suggestions based on buy orders

## Related Documentation

- [Phase 4: Purchase System](../purchases/README.md)
- [Contact System](../contacts/README.md)
- [Permissions System](../permissions/README.md)
