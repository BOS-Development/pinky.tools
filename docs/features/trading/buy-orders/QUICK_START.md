# Buy Orders Quick Start Guide

Get up and running with Buy Orders in 5 minutes.

## Prerequisites

- Phase 1-4 completed (Contacts, Permissions, For-Sale, Purchases)
- At least one contact with accepted status
- Database migrated to include `buy_orders` table

## Step 1: Database Migration

The migration runs automatically on startup. Verify it's applied:

```bash
# Check if migration 12 exists
ls internal/database/migrations/ | grep 12_buy_orders
```

You should see:
```
12_buy_orders.down.sql
12_buy_orders.up.sql
```

## Step 2: Create a Buy Order (Buyer)

### Via UI

1. Navigate to **Marketplace** → **My Buy Orders** tab
2. Click **"Create Buy Order"** button
3. In the dialog:
   - **Item Name**: Type "Trit" - autocomplete shows "Tritanium" with icon
   - **Quantity Desired**: 1000000
   - **Max Price Per Unit**: 6
   - **Notes**: "Urgent need for manufacturing"
4. Click **"Create"**

### Via API

```bash
curl -X POST http://localhost/api/v1/buy-orders \
  -H "USER-ID: 12345" \
  -H "BACKEND-KEY: your-backend-key" \
  -H "Content-Type: application/json" \
  -d '{
    "typeId": 34,
    "quantityDesired": 1000000,
    "maxPricePerUnit": 6,
    "notes": "Urgent need for manufacturing"
  }'
```

### Via Frontend API

```typescript
const response = await fetch('/api/buy-orders', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    typeId: 34,
    quantityDesired: 1000000,
    maxPricePerUnit: 6,
    notes: 'Urgent need for manufacturing'
  })
});

const order = await response.json();
console.log('Created order:', order);
```

## Step 3: Grant Permission

For your contact to see your buy order, they need permission:

### Via UI

1. Navigate to **Contacts** page
2. Find your contact
3. Toggle **"For-Sale Browse"** permission to ON
4. Contact can now see your buy orders

### Via API

```bash
curl -X POST http://localhost/api/v1/contacts/{contactId}/permissions \
  -H "USER-ID: 12345" \
  -H "BACKEND-KEY: your-backend-key" \
  -H "Content-Type: application/json" \
  -d '{
    "serviceType": "for_sale_browse",
    "canAccess": true
  }'
```

## Step 4: View Demand (Seller)

### Via UI

1. Navigate to **Marketplace** → **Demand** tab
2. See **Aggregated Demand** table showing:
   - Item name
   - Total quantity wanted across all orders
   - Highest price offered
   - Potential revenue
   - Number of orders
3. Scroll to **Individual Buy Orders** table for details

### Via API

```bash
curl -X GET http://localhost/api/v1/buy-orders/demand \
  -H "USER-ID: 67890" \
  -H "BACKEND-KEY: your-backend-key"
```

Response:
```json
[
  {
    "id": 1,
    "buyerUserId": 12345,
    "typeId": 34,
    "typeName": "Tritanium",
    "quantityDesired": 1000000,
    "maxPricePerUnit": 6,
    "notes": "Urgent need for manufacturing",
    "isActive": true,
    "createdAt": "2026-02-15T10:30:00Z",
    "updatedAt": "2026-02-15T10:30:00Z"
  }
]
```

## Step 5: Update a Buy Order

### Via UI

1. In **My Buy Orders** tab, click the **Edit** icon
2. Modify quantity or price
3. Click **"Update"**

Note: You cannot change the item type of an existing order.

### Via API

```bash
curl -X PUT http://localhost/api/v1/buy-orders/1 \
  -H "USER-ID: 12345" \
  -H "BACKEND-KEY: your-backend-key" \
  -H "Content-Type: application/json" \
  -d '{
    "quantityDesired": 1500000,
    "maxPricePerUnit": 7,
    "isActive": true
  }'
```

## Step 6: Delete a Buy Order

### Via UI

1. In **My Buy Orders** tab, click the **Delete** icon
2. Confirm deletion
3. Order is soft-deleted (sets `is_active = false`)

### Via API

```bash
curl -X DELETE http://localhost/api/v1/buy-orders/1 \
  -H "USER-ID: 12345" \
  -H "BACKEND-KEY: your-backend-key"
```

## Common Use Cases

### Use Case 1: I want to buy Tritanium

**Scenario:** You need 1 million Tritanium, willing to pay up to 6 ISK/unit.

```typescript
// Create buy order
await fetch('/api/buy-orders', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    typeId: 34,  // Tritanium
    quantityDesired: 1000000,
    maxPricePerUnit: 6,
    notes: 'Need for Rifter production'
  })
});

// Your contacts who have granted you permission can now see this
```

### Use Case 2: See what my contacts want to buy

**Scenario:** You're a seller, want to know what's in demand.

```typescript
// Fetch demand
const response = await fetch('/api/buy-orders/demand');
const demand = await response.json();

// Group by item type for planning
const demandByType = demand.reduce((acc, order) => {
  if (!acc[order.typeId]) {
    acc[order.typeId] = {
      typeName: order.typeName,
      totalQuantity: 0,
      maxPrice: 0,
      orders: []
    };
  }
  acc[order.typeId].totalQuantity += order.quantityDesired;
  acc[order.typeId].maxPrice = Math.max(acc[order.typeId].maxPrice, order.maxPricePerUnit);
  acc[order.typeId].orders.push(order);
  return acc;
}, {});

console.log('Demand by type:', demandByType);
```

### Use Case 3: Update buy order based on market changes

**Scenario:** Market prices increased, you're willing to pay more.

```typescript
// Get current order
const response = await fetch('/api/buy-orders');
const orders = await response.json();
const tritOrder = orders.find(o => o.typeId === 34);

// Update price
await fetch(`/api/buy-orders/${tritOrder.id}`, {
  method: 'PUT',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    quantityDesired: tritOrder.quantityDesired,
    maxPricePerUnit: 7,  // Increased from 6
    isActive: true
  })
});
```

### Use Case 4: Cancel buy order (got the items elsewhere)

**Scenario:** You found Tritanium from another source, no longer need the buy order.

```typescript
// Delete the order
await fetch(`/api/buy-orders/${orderId}`, {
  method: 'DELETE'
});

// Order is now inactive, won't show in demand view
```

## Autocomplete Feature

The item autocomplete provides a smooth search experience:

### How It Works

1. **Type minimum 2 characters**: "tr"
2. **Debounce 300ms**: Waits for you to stop typing
3. **Search backend**: Query sent to `/api/item-types/search?q=tr`
4. **Relevance sorting**:
   - "Tritanium" (starts with "tr")
   - "Compressed Tritanium" (contains "tr")
5. **Display with icons**: Shows EVE item icons next to names
6. **Select item**: Icon appears in input field

### Customizing Autocomplete

```typescript
// In your component
const [itemOptions, setItemOptions] = useState([]);
const [searchLoading, setSearchLoading] = useState(false);

const searchItems = async (query: string) => {
  if (query.length < 2) return;

  setSearchLoading(true);
  try {
    const response = await fetch(`/api/item-types/search?q=${encodeURIComponent(query)}`);
    const items = await response.json();
    setItemOptions(items);
  } finally {
    setSearchLoading(false);
  }
};

// Debounce function
const debouncedSearch = useMemo(
  () => debounce(searchItems, 300),
  []
);
```

## EVE Image URLs

Use the helper functions for consistent image URLs:

```typescript
import { getItemIconUrl } from "@industry-tool/utils/eveImages";

// 32px icon for dropdown
<img src={getItemIconUrl(34, 32)} alt="Tritanium" />

// 64px icon for larger display
<img src={getItemIconUrl(34, 64)} alt="Tritanium" />

// Available sizes: 32, 64, 128, 256, 512
```

## Troubleshooting

### Buy order not showing in demand view

**Problem:** Created a buy order, but contact can't see it.

**Solutions:**
1. Check contact status is "accepted"
2. Verify you granted contact `for_sale_browse` permission
3. Ensure order is `isActive = true`
4. Refresh the demand view

### Autocomplete not working

**Problem:** Item search returns no results.

**Solutions:**
1. Type at least 2 characters
2. Check database has `asset_item_types` populated
3. Verify backend endpoint `/v1/item-types/search` is working
4. Check browser console for API errors

### Cannot update item type

**Problem:** Trying to change typeId of existing order fails.

**Solution:**
- You cannot change item type of existing orders
- Delete the order and create a new one with different item

### Permission denied errors

**Problem:** Getting 403 errors when trying to update/delete orders.

**Solutions:**
1. Verify you own the order (buyerUserId matches your ID)
2. Check authentication headers are correct
3. Ensure you're not trying to modify another user's order

## Next Steps

- [API Documentation](API.md) - Full API reference
- [Testing Guide](TESTING.md) - Test suite details
- [Implementation Summary](IMPLEMENTATION.md) - Technical architecture
- [Phase 6: Sales Metrics](../sales-metrics/README.md) - Analytics dashboard (coming soon)

## Tips & Best Practices

1. **Always use autocomplete** - Don't hardcode typeIds, use the search
2. **Validate inputs** - Check quantity > 0 and price >= 0 before submitting
3. **Handle errors gracefully** - Show user-friendly error messages
4. **Debounce searches** - Don't hammer the API on every keystroke
5. **Cache item data** - Once you have an item's data, cache it
6. **Permission first** - Grant permissions before creating orders if you want contacts to see them immediately
7. **Soft deletes** - Remember deleted orders are still in database, just `is_active = false`
8. **ISK formatting** - Always format ISK values with thousand separators: `1,000,000 ISK`
