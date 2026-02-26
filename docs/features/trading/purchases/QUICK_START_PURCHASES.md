# Quick Start: Purchase System

A quick guide to using the purchase system API.

## Prerequisites

1. Two users with accepted contact relationship
2. Seller has granted `for_sale_browse` permission to buyer
3. Seller has active for-sale listings

## Step-by-Step Purchase Flow

### 1. Buyer: Browse Available Items

```bash
GET /v1/for-sale/browse
Authorization: Bearer <buyer-token>
```

**Response:**
```json
[
  {
    "id": 123,
    "userId": 1002,
    "userName": "Seller Corp",
    "typeId": 34,
    "typeName": "Tritanium",
    "quantityAvailable": 1000000,
    "pricePerUnit": 6,
    "locationId": 30000142,
    "locationName": "Jita IV - Moon 4 - Caldari Navy Assembly Plant",
    "isActive": true
  }
]
```

### 2. Buyer: Purchase Items

```bash
POST /v1/purchases
Authorization: Bearer <buyer-token>
Content-Type: application/json

{
  "forSaleItemId": 123,
  "quantityPurchased": 50000
}
```

**Response:**
```json
{
  "id": 456,
  "forSaleItemId": 123,
  "buyerUserId": 1001,
  "sellerUserId": 1002,
  "typeId": 34,
  "typeName": "Tritanium",
  "quantityPurchased": 50000,
  "pricePerUnit": 6,
  "totalPrice": 300000,
  "status": "pending",
  "locationId": 30000142,
  "locationName": "Jita IV - Moon 4 - Caldari Navy Assembly Plant",
  "purchasedAt": "2026-02-15T10:30:00Z"
}
```

### 3. Seller: View Pending Sales

```bash
GET /v1/purchases/pending
Authorization: Bearer <seller-token>
```

**Response:**
```json
[
  {
    "id": 456,
    "buyerUserId": 1001,
    "buyerName": "John Doe",
    "typeId": 34,
    "typeName": "Tritanium",
    "quantityPurchased": 50000,
    "pricePerUnit": 6,
    "totalPrice": 300000,
    "status": "pending",
    "locationId": 30000142,
    "locationName": "Jita IV - Moon 4 - Caldari Navy Assembly Plant",
    "purchasedAt": "2026-02-15T10:30:00Z"
  }
]
```

### 4. Seller: Create In-Game Contract

In EVE Online:
1. Right-click in station → Create Contract
2. Type: Item Exchange
3. Availability: Private (to buyer character)
4. Add items matching purchase quantities
5. Set price = 0 ISK (already paid via tool)
6. Copy contract ID from contract details

### 5. Seller: Mark Contract Created

```bash
POST /v1/purchases/456/mark-contract-created
Authorization: Bearer <seller-token>
Content-Type: application/json

{
  "contractKey": "123456789"
}
```

**Response:**
```json
{
  "id": 456,
  "status": "contract_created",
  "contractKey": "123456789"
}
```

### 6. Buyer: View Contract and Accept

```bash
GET /v1/purchases/buyer
Authorization: Bearer <buyer-token>
```

**Response:**
```json
[
  {
    "id": 456,
    "status": "contract_created",
    "contractKey": "123456789",
    "typeName": "Tritanium",
    "quantityPurchased": 50000,
    ...
  }
]
```

In EVE Online:
1. Open Contracts window
2. Find contract by ID (contractKey)
3. Accept contract

### 7. Buyer: Mark Purchase Complete

```bash
POST /v1/purchases/456/complete
Authorization: Bearer <buyer-token>
```

**Response:**
```json
{
  "id": 456,
  "status": "completed"
}
```

## Common Operations

### Cancel a Purchase

**Before contract is created:**
```bash
POST /v1/purchases/456/cancel
Authorization: Bearer <buyer-or-seller-token>
```

This automatically restores the quantity to the for-sale listing.

### View Purchase History

**As buyer:**
```bash
GET /v1/purchases/buyer
Authorization: Bearer <buyer-token>
```

**As seller:**
```bash
GET /v1/purchases/seller
Authorization: Bearer <seller-token>
```

### Batch Contract Creation

If buyer purchases multiple items from same seller/location:

```bash
# Purchase item 1
POST /v1/purchases
{ "forSaleItemId": 123, "quantityPurchased": 50000 }
# Returns: { "id": 456 }

# Purchase item 2
POST /v1/purchases
{ "forSaleItemId": 124, "quantityPurchased": 25000 }
# Returns: { "id": 457 }

# Seller creates ONE contract with both items
# Then marks both purchases with same contract key:
POST /v1/purchases/bulk-mark-contract-created
{
  "purchaseIds": [456, 457],
  "contractKey": "123456789"
}
```

## Error Handling

### 403 Forbidden - No Permission

```json
{
  "error": "you do not have permission to purchase from this seller"
}
```

**Fix:** Ask seller to grant `for_sale_browse` permission in Contacts settings.

### 400 Bad Request - Quantity Exceeded

```json
{
  "error": "quantity requested (75000) exceeds available quantity (50000)"
}
```

**Fix:** Reduce purchase quantity or wait for seller to restock.

### 404 Not Found - Item Not Found

```json
{
  "error": "for-sale item not found"
}
```

**Fix:** Item may have been delisted. Browse again for current listings.

## Testing with cURL

### Setup

```bash
# Set your auth tokens
BUYER_TOKEN="eyJhbGc..."
SELLER_TOKEN="eyJhbGc..."

# Set base URL
API_BASE="http://localhost:3000/api"
```

### Complete Flow

```bash
# 1. Browse items (as buyer)
curl -X GET "$API_BASE/for-sale/browse" \
  -H "Authorization: Bearer $BUYER_TOKEN"

# 2. Purchase item
curl -X POST "$API_BASE/purchases" \
  -H "Authorization: Bearer $BUYER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"forSaleItemId": 123, "quantityPurchased": 50000}'

# 3. View pending sales (as seller)
curl -X GET "$API_BASE/purchases/pending" \
  -H "Authorization: Bearer $SELLER_TOKEN"

# 4. Mark contract created
curl -X POST "$API_BASE/purchases/456/mark-contract-created" \
  -H "Authorization: Bearer $SELLER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"contractKey": "123456789"}'

# 5. Complete purchase (as buyer)
curl -X POST "$API_BASE/purchases/456/complete" \
  -H "Authorization: Bearer $BUYER_TOKEN"
```

## Next Steps

- **Full Documentation:** See [PURCHASES.md](PURCHASES.md)
- **Testing:** See [TESTING_PURCHASES.md](TESTING_PURCHASES.md)
- **Phase 5:** Buy Orders & Demand Tracking (coming soon)
