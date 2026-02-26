# Contact System & Industrial Services Marketplace

## Overview

Complete contact system with granular permissions and extensible industrial services marketplace. Users can add contacts, grant service permissions, list items for sale, browse contacts' listings, and complete purchases with full transaction tracking.

**Status:** Planned (4 phases)

---

## Core Features

### 1. Contact System
- **Bidirectional Relationships**: Users send requests, recipients accept/reject
- **Status Tracking**: Pending → Accepted/Rejected
- **Contact Management**: View all contacts, sent/received requests, remove contacts

### 2. Permission System
- **Service-Based**: Permissions tied to service types (`for_sale_browse`, `manufacturing`, etc.)
- **Unidirectional Control**: Each user independently grants permissions to contacts
- **Extensible**: New services can be added without schema changes

### 3. For-Sale Items Marketplace
- **List from Any Inventory**: Character/corp assets, hangars/containers/divisions
- **Price Per Unit**: ISK pricing for items
- **Permission-Filtered Browsing**: Only see items from contacts who granted access
- **Search & Filter**: By item, contact, location, price

### 4. Purchase Transaction System
- **Atomic Transactions**: Quantity updates and purchase records succeed/fail together
- **Concurrency Handling**: Database row locks prevent overselling
- **Transaction History**: Immutable audit log for buyers and sellers
- **Permission Checks**: Re-verify permissions before purchase

---

## Database Schema

### Tables

**contacts**
- `id` (PK)
- `requester_user_id` (FK → users)
- `recipient_user_id` (FK → users)
- `status` ('pending', 'accepted', 'rejected')
- `requested_at`, `responded_at`
- UNIQUE constraint on (requester, recipient)
- CHECK constraint prevents self-contacts

**contact_permissions**
- `id` (PK)
- `contact_id` (FK → contacts, CASCADE DELETE)
- `granting_user_id` (FK → users)
- `receiving_user_id` (FK → users)
- `service_type` (varchar)
- `can_access` (boolean)
- UNIQUE constraint on (contact, granting, receiving, service_type)
- Index on (receiving_user_id, service_type, can_access) for permission checks

**for_sale_items**
- `id` (PK)
- `user_id` (FK → users)
- `type_id` (FK → asset_item_types)
- `owner_type` ('character', 'corporation')
- `owner_id`
- `location_id`, `container_id`, `division_number`
- `quantity_available`, `price_per_unit`
- `notes`, `is_active`
- UNIQUE index on (user, type, owner_type, owner, location, container, division) WHERE active
- Mirrors `stockpile_markers` pattern

**purchase_transactions**
- `id` (PK)
- `for_sale_item_id` (FK → for_sale_items)
- `buyer_user_id` (FK → users)
- `seller_user_id` (FK → users)
- `type_id` (FK → asset_item_types)
- `quantity_purchased`, `price_per_unit`, `total_price`
- `status` ('completed', 'cancelled', 'failed')
- `purchased_at`
- Indexes on (buyer_user_id, purchased_at DESC) and (seller_user_id, purchased_at DESC)
- CHECK constraint prevents self-purchase

**buy_orders** (Phase 5)
- `id` (PK)
- `buyer_user_id` (FK → users)
- `type_id` (FK → asset_item_types)
- `location_id` (bigint NOT NULL) — station/structure where buyer wants delivery
- `quantity_desired`, `max_price_per_unit`
- `notes`, `is_active`
- Indexes on (buyer_user_id), (type_id), (is_active)
- CHECK constraint positive quantity and price
- Location name resolved via LEFT JOIN to `stations` and `solar_systems`

---

## Backend API

### Contacts Endpoints
- `GET /v1/contacts` - List all contacts
- `POST /v1/contacts` - Send contact request
  - Body: `{ "recipientUserId": 123 }`
- `POST /v1/contacts/{id}/accept` - Accept request
- `POST /v1/contacts/{id}/reject` - Reject request
- `DELETE /v1/contacts/{id}` - Remove contact

### Permissions Endpoints
- `GET /v1/contacts/{id}/permissions` - Get all permissions for contact
- `POST /v1/contacts/{id}/permissions` - Update permission
  - Body: `{ "serviceType": "for_sale_browse", "canAccess": true }`

### For-Sale Items Endpoints
- `GET /v1/for-sale` - Get user's listings
- `GET /v1/for-sale/browse` - Browse contacts' listings (permission-filtered)
- `POST /v1/for-sale` - Create listing
  - Body: `{ "typeId": 34, "ownerType": "character", "ownerId": 789, "locationId": 60003760, "quantityAvailable": 1000, "pricePerUnit": 50000 }`
- `PUT /v1/for-sale/{id}` - Update listing
- `DELETE /v1/for-sale/{id}` - Delete listing (soft-delete)

### Purchase Endpoints
- `POST /v1/purchases` - Purchase item
  - Body: `{ "forSaleItemId": 1, "quantityPurchased": 100 }`
- `GET /v1/purchases/buyer` - Buyer's purchase history
- `GET /v1/purchases/seller` - Seller's sales history

### Buy Orders Endpoints (Phase 5)
- `GET /v1/buy-orders` - Get user's buy orders
- `GET /v1/buy-orders/demand` - View buy orders from contacts (seller view)
- `POST /v1/buy-orders` - Create buy order
  - Body: `{ "typeId": 34, "locationId": 60003760, "quantityDesired": 1000, "maxPricePerUnit": 50000, "notes": "Urgent" }`
- `PUT /v1/buy-orders/{id}` - Update buy order
  - Body: `{ "locationId": 60003760, "quantityDesired": 1000, "maxPricePerUnit": 50000, "notes": "Urgent", "isActive": true }`
- `DELETE /v1/buy-orders/{id}` - Cancel buy order

### Station Search Endpoint
- `GET /v1/stations/search?q=...` - Search stations by name (partial, case-insensitive)

### Analytics Endpoints (Phase 6)
- `GET /v1/analytics/sales` - Sales metrics with time filter
  - Query params: `?period=30d` (7d, 30d, 90d, 1y, all-time)
  - Response: volume sold per item, revenue, time-series data
- `GET /v1/analytics/demand-comparison` - Compare sales to stockpile markers

---

## Frontend Pages

### /contacts
**Contact Management UI**

- **Tabs**: My Contacts | Pending Requests | Sent Requests
- **Actions**: Send request, accept, reject, remove
- **Features**:
  - Search/filter contacts
  - Inline permission toggles
  - "Add Contact" button

**Components:**
- `ContactsList.tsx` - Main list component
- `PermissionsDialog.tsx` - Modal for managing permissions

### /marketplace
**Marketplace UI**

- **Tabs**: Browse | My Listings | Demand | Analytics | History

**Browse Tab:**
- Filter by contact, item, location
- Sort by price, quantity
- Purchase button with quantity selector
- Shows: item name, seller, location, quantity, price

**My Listings Tab:**
- List all active listings
- Add new listing (select from inventory)
- Edit price/quantity
- Delist item
- View sales history per item

**History Tab:**
- Purchases (as buyer)
- Sales (as seller)
- Pagination
- Filter by date range, item type

**Demand Tab:** (Phase 5)
- View buy orders from contacts
- Filter by item type, quantity, max price
- Shows: buyer name, item, quantity desired, max price per unit
- "Create Listing" button to fulfill buy orders

**Analytics Tab:** (Phase 6)
- Time-series chart: quantity sold over time (daily/weekly/monthly)
- Bar chart: top selling items
- Revenue summary cards
- Stockpile recommendation widget
- Time period filters: 7d, 30d, 90d, 1y, all-time
- Export to CSV

**Components:**
- `MarketplaceBrowser.tsx` - Browse & purchase
- `MyListings.tsx` - Listing management
- `PurchaseDialog.tsx` - Purchase confirmation modal
- `PurchaseHistory.tsx` - Transaction history
- `BuyOrderDialog.tsx` - Buy order creation (Phase 5)
- `DemandViewer.tsx` - View contact buy orders (Phase 5)
- `SalesMetrics.tsx` - Analytics dashboard with charts (Phase 6)

---

## Implementation Phases

### Phase 1: Foundation (Contacts & Permissions)
**Goal:** Users can add contacts and manage permissions

**Deliverables:**
- Migrations 7-8 (contacts, contact_permissions tables)
- Contacts repository & controller
- ContactPermissions repository & controller
- ContactsList component
- PermissionsDialog component
- `/contacts` page

**Verification:**
- Send/accept/reject contact requests
- Toggle permissions per contact
- View bidirectional permissions

---

### Phase 2: For-Sale Items (Listings)
**Goal:** Users can list items for sale

**Deliverables:**
- Migration 9 (for_sale_items table)
- ForSaleItems repository & controller
- MyListings component
- `/marketplace` page with "My Listings" tab

**Verification:**
- Create listing from any inventory location
- Edit price/quantity
- Delete listing

---

### Phase 3: Marketplace Browsing
**Goal:** Users can browse contacts' items

**Deliverables:**
- GetBrowsableItems repository method
- BrowseListings controller endpoint
- MarketplaceBrowser component
- "Browse" tab in marketplace

**Verification:**
- Browse items (only permitted contacts visible)
- Filter/sort functionality
- Permission changes update view

---

### Phase 4: Purchase System
**Goal:** Users can purchase items with transaction tracking

**Deliverables:**
- Migration 10 (purchase_transactions table)
- PurchaseTransactions repository
- Purchases controller (atomic transaction logic)
- PurchaseDialog component
- PurchaseHistory component
- "History" tab in marketplace

**Verification:**
- Purchase item (quantity decreases)
- Concurrent purchase handling
- View purchase/sales history
- Transaction rollback on failure

---

### Phase 5: Buy Orders & Demand Tracking
**Goal:** Buyers can place buy orders for out-of-stock items; sellers see demand

**Deliverables:**
- Migration 11 (buy_orders table)
- BuyOrders repository & controller
- BuyOrderDialog component
- DemandViewer component
- "Place Buy Order" button in MarketplaceBrowser
- "Demand" tab in marketplace

**Verification:**
- Create buy order for out-of-stock item
- Edit buy order (quantity, max price)
- Cancel buy order
- Seller sees buy orders from contacts
- Filter/sort buy orders by item type

---

### Phase 6: Sales Metrics Dashboard
**Goal:** Sellers see analytics to plan stockpiles based on sales data

**Deliverables:**
- SalesAnalytics repository (aggregate purchase_transactions)
- Analytics controller
- SalesMetrics component with charts
- "Analytics" tab in marketplace

**Verification:**
- View sales volume chart (7d/30d/90d/1y)
- See top 10 selling items
- View total revenue by item type
- Compare sales velocity to stockpile markers
- Export sales data to CSV

---

## Critical Edge Cases

### Contact System
- **Duplicate requests**: UNIQUE constraint prevents
- **Self-contact**: CHECK constraint prevents
- **Unauthorized accept**: Verify user is recipient

### Permissions
- **Permission without contact**: CheckPermission returns false
- **Partial permissions**: Unidirectional (A allows B, B doesn't allow A)
- **Permission revocation mid-browse**: Re-check before purchase

### Purchases
- **Overselling (concurrent purchases)**: Database row locks (SELECT FOR UPDATE)
- **Quantity exceeds available**: Validate before transaction
- **Transaction rollback**: defer tx.Rollback() handles failures
- **Inactive item purchase**: GetByID filters is_active=true
- **Self-purchase**: Check buyer != seller

---

## Extensibility

### Future Services Architecture

The permission system supports additional industrial services:

**Planned Services:**
- Manufacturing (build jobs for contacts)
- Hauling (transport items between systems)
- Material Exchange (buy orders)
- Reaction slot sharing
- Research slot sharing

**Adding New Service:**
1. Add service_type value (e.g., `'manufacturing'`) - no schema change
2. Create service-specific table (e.g., `manufacturing_services`)
3. Create repository & controller
4. Reuse `CheckPermission` with new service_type
5. Add frontend components

---

## Technical Implementation

### Backend Pattern
**Repository → Controller → Routes** (existing pattern)

**Files:**
- `/internal/repositories/contacts.go`
- `/internal/repositories/contactPermissions.go`
- `/internal/repositories/forSaleItems.go`
- `/internal/repositories/purchaseTransactions.go`
- `/internal/controllers/contacts.go`
- `/internal/controllers/contactPermissions.go`
- `/internal/controllers/forSaleItems.go`
- `/internal/controllers/purchases.go`

### Frontend Pattern
**Next.js API Routes → Backend → MUI Components** (existing pattern)

**Files:**
- `/frontend/pages/api/contacts/*` - Proxy routes
- `/frontend/pages/api/for-sale/*` - Proxy routes
- `/frontend/pages/api/purchases/*` - Proxy routes
- `/frontend/packages/components/contacts/*` - Contact components
- `/frontend/packages/components/marketplace/*` - Marketplace components
- `/frontend/packages/pages/contacts.tsx` - Contacts page
- `/frontend/packages/pages/marketplace.tsx` - Marketplace page

### Atomic Purchase Transaction Logic

```go
func (c *Purchases) PurchaseItem(args *web.HandlerArgs) (any, *web.HttpError) {
    // 1. Get item
    item, _ := c.forSaleRepo.GetByID(ctx, req.ForSaleItemID)

    // 2. Verify permission
    hasPermission, _ := c.permissionsRepo.CheckPermission(ctx, item.UserID, *args.User, "for_sale_browse")
    if !hasPermission { return 403 }

    // 3. Validate quantity
    if req.QuantityPurchased > item.QuantityAvailable { return 400 }

    // 4. Begin transaction
    tx, _ := c.db.BeginTx(ctx, nil)
    defer tx.Rollback()

    // 5. Update quantity
    newQuantity := item.QuantityAvailable - req.QuantityPurchased
    c.forSaleRepo.UpdateQuantity(ctx, tx, item.ID, newQuantity)

    // 6. Create purchase record
    purchase := &models.PurchaseTransaction{
        BuyerUserID: *args.User,
        SellerUserID: item.UserID,
        QuantityPurchased: req.QuantityPurchased,
        PricePerUnit: item.PricePerUnit,
        TotalPrice: item.PricePerUnit * req.QuantityPurchased,
    }
    c.repository.Create(ctx, tx, purchase)

    // 7. Commit
    tx.Commit()
}
```

---

## Verification Checklist

### Phase 1 (Contacts)
- [ ] Send contact request
- [ ] Accept/reject request
- [ ] View pending/sent requests
- [ ] Delete contact
- [ ] Toggle permission
- [ ] View bidirectional permissions

### Phase 2 (Listings)
- [ ] Create listing (character hangar)
- [ ] Create listing (character container)
- [ ] Create listing (corp hangar)
- [ ] Create listing (corp container)
- [ ] Edit price/quantity
- [ ] Delete listing

### Phase 3 (Browsing)
- [ ] Browse marketplace
- [ ] See only permitted contacts
- [ ] Filter by item/contact/location
- [ ] Sort by price/quantity
- [ ] Permission change updates view

### Phase 4 (Purchases)
- [ ] Purchase item
- [ ] Quantity decreases correctly
- [ ] Purchase full quantity (listing inactive)
- [ ] Purchase > available (fails 400)
- [ ] Self-purchase (fails 400)
- [ ] View buyer history
- [ ] View seller history
- [ ] Concurrent purchase (row locking works)

### Phase 5 (Buy Orders)
- [ ] Create buy order for out-of-stock item
- [ ] Edit buy order (quantity, max price)
- [ ] Cancel buy order
- [ ] Seller sees buy orders from contacts (demand view)
- [ ] Filter buy orders by item type
- [ ] Sort buy orders by quantity/price
- [ ] "Place Buy Order" button appears when item out of stock

### Phase 6 (Sales Analytics)
- [ ] View sales volume chart (7d/30d/90d/1y)
- [ ] See top 10 selling items
- [ ] View total revenue by item type
- [ ] Compare sales velocity to stockpile markers
- [ ] View buyer analytics (top buyers, repeat rate)
- [ ] Filter metrics by time period
- [ ] Export sales data to CSV
- [ ] Drill down into item-specific analytics
