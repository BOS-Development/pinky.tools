# Purchase System API Reference

Quick API reference for the purchase system endpoints.

## Base URL

```
Production: https://your-domain.com/api
Development: http://localhost:3000/api
```

All endpoints require authentication via JWT token in Authorization header:
```
Authorization: Bearer <token>
```

---

## Endpoints

### POST /v1/purchases

Create a new purchase transaction.

**Request:**
```json
{
  "forSaleItemId": 123,
  "quantityPurchased": 50000
}
```

**Response (201):**
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

**Errors:**
- `400` - Invalid request, quantity exceeded
- `403` - No permission or self-purchase
- `404` - Item not found
- `500` - Internal error

---

### GET /v1/purchases/buyer

Get authenticated user's purchase history.

**Response (200):**
```json
[
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
    "status": "completed",
    "contractKey": "123456789",
    "locationId": 30000142,
    "locationName": "Jita IV - Moon 4 - Caldari Navy Assembly Plant",
    "purchasedAt": "2026-02-15T10:30:00Z"
  }
]
```

**Notes:**
- Ordered by `purchasedAt DESC` (newest first)
- Includes all statuses

---

### GET /v1/purchases/seller

Get authenticated user's sales history.

**Response (200):**
Same format as `/v1/purchases/buyer`

---

### GET /v1/purchases/pending

Get authenticated user's pending sales (seller view).

**Response (200):**
```json
[
  {
    "id": 456,
    "forSaleItemId": 123,
    "buyerUserId": 1001,
    "buyerName": "John Doe",
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
]
```

**Notes:**
- Only returns `status = 'pending'`
- Excludes contract_created, completed, cancelled
- Includes buyer character name

---

### POST /v1/purchases/:id/mark-contract-created

Seller marks purchase as contract created (status: pending → contract_created).

**Authorization:** Only seller can call this

**Request:**
```json
{
  "contractKey": "123456789"
}
```

**Response (200):**
```json
{
  "id": 456,
  "status": "contract_created",
  "contractKey": "123456789"
}
```

**Errors:**
- `400` - Missing contract key
- `403` - Not the seller
- `404` - Purchase not found

---

### POST /v1/purchases/:id/complete

Buyer marks purchase as completed (status: contract_created → completed).

**Authorization:** Only buyer can call this

**Response (200):**
```json
{
  "id": 456,
  "status": "completed"
}
```

**Errors:**
- `403` - Not the buyer
- `404` - Purchase not found

---

### POST /v1/purchases/:id/cancel

Cancel purchase and restore quantity to for-sale item (status: * → cancelled).

**Authorization:** Buyer OR seller can call this

**Response (200):**
```json
{
  "id": 456,
  "status": "cancelled"
}
```

**Notes:**
- Atomically updates status and restores quantity
- Reactivates for-sale item if it was inactive

**Errors:**
- `403` - Not buyer or seller
- `404` - Purchase not found

---

### POST /v1/purchases/bulk-mark-contract-created

Mark multiple purchases with same contract key (for batch contracts).

**Authorization:** Must be seller for all purchases

**Request:**
```json
{
  "purchaseIds": [456, 457, 458],
  "contractKey": "123456789"
}
```

**Response (200):**
```json
{
  "updatedCount": 3
}
```

**Errors:**
- `400` - Empty purchase IDs or missing contract key
- `403` - Not seller for one or more purchases
- `404` - One or more purchases not found

---

## Status Workflow

```
pending → contract_created → completed
   ↓              ↓              ↓
cancelled ← ← ← ← ← ← ← ← ← ← ← ←
```

**Status Values:**
- `pending` - Purchase created, awaiting seller action
- `contract_created` - Seller created in-game contract
- `completed` - Buyer completed contract
- `cancelled` - Purchase cancelled, quantity restored

---

## Field Descriptions

### Purchase Transaction Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | integer | Unique purchase ID |
| `forSaleItemId` | integer | Reference to for-sale item |
| `buyerUserId` | integer | Buyer's user ID |
| `sellerUserId` | integer | Seller's user ID |
| `typeId` | integer | EVE item type ID |
| `typeName` | string | EVE item type name |
| `quantityPurchased` | integer | Quantity purchased |
| `pricePerUnit` | integer | ISK per unit |
| `totalPrice` | integer | Total ISK (quantity × price) |
| `status` | string | pending/contract_created/completed/cancelled |
| `contractKey` | string | In-game contract ID (optional) |
| `locationId` | integer | Solar system ID |
| `locationName` | string | Solar system name |
| `buyerName` | string | Buyer character name (pending sales only) |
| `purchasedAt` | timestamp | Purchase creation time (ISO 8601) |

---

## Common Error Responses

### 400 Bad Request
```json
{
  "error": "quantity requested (75000) exceeds available quantity (50000)"
}
```

### 403 Forbidden
```json
{
  "error": "you do not have permission to purchase from this seller"
}
```

or

```json
{
  "error": "only the seller can mark contract created"
}
```

### 404 Not Found
```json
{
  "error": "purchase transaction not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "internal server error"
}
```

---

## Rate Limiting

Currently no rate limits enforced. Future versions may implement:
- 10 purchases per minute per user
- 100 API calls per minute per user

---

## Pagination

History endpoints currently return all results. Future versions may implement pagination:
```
GET /v1/purchases/buyer?limit=50&offset=0
```

---

## Filtering & Sorting

Future versions may support:
```
GET /v1/purchases/buyer?status=completed&sort=purchasedAt:desc
GET /v1/purchases/seller?typeId=34&dateFrom=2026-01-01
```

---

## Webhooks

Future versions may support webhooks for:
- New purchase notification (seller)
- Contract created notification (buyer)
- Purchase completed notification (seller)

---

## TypeScript Types

```typescript
interface PurchaseTransaction {
  id: number;
  forSaleItemId: number;
  buyerUserId: number;
  sellerUserId: number;
  typeId: number;
  typeName: string;
  quantityPurchased: number;
  pricePerUnit: number;
  totalPrice: number;
  status: 'pending' | 'contract_created' | 'completed' | 'cancelled';
  contractKey?: string;
  locationId: number;
  locationName: string;
  buyerName?: string;  // Only in pending sales
  purchasedAt: string;  // ISO 8601
}

interface PurchaseRequest {
  forSaleItemId: number;
  quantityPurchased: number;
}

interface MarkContractCreatedRequest {
  contractKey: string;
}

interface BulkMarkContractCreatedRequest {
  purchaseIds: number[];
  contractKey: string;
}
```

---

## Related APIs

- **For-Sale Items API**: `/v1/for-sale/*` - Browse and manage listings
- **Contacts API**: `/v1/contacts/*` - Manage contact relationships
- **Permissions API**: `/v1/contacts/:id/permissions` - Manage service permissions

---

## Further Reading

- [PURCHASES.md](PURCHASES.md) - Complete documentation
- [QUICK_START_PURCHASES.md](QUICK_START_PURCHASES.md) - Quick start guide
- [TESTING_PURCHASES.md](TESTING_PURCHASES.md) - Test suite
