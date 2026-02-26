# Buy Orders API Documentation

RESTful API endpoints for managing buy orders and viewing demand.

## Base URL

```
http://localhost/api/v1
```

## Authentication

All endpoints require authentication via `USER-ID` header.

```http
USER-ID: 12345
BACKEND-KEY: your-backend-key
```

---

## Endpoints

### 1. Get My Buy Orders

Retrieve all buy orders created by the authenticated user.

**Endpoint:** `GET /buy-orders`

**Response:** `200 OK`

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

**Example:**

```bash
curl -X GET http://localhost/api/v1/buy-orders \
  -H "USER-ID: 12345" \
  -H "BACKEND-KEY: your-backend-key"
```

---

### 2. Create Buy Order

Create a new buy order.

**Endpoint:** `POST /buy-orders`

**Request Body:**

```json
{
  "typeId": 34,
  "quantityDesired": 1000000,
  "maxPricePerUnit": 6,
  "notes": "Urgent need for manufacturing"
}
```

**Validation:**
- `typeId` - Required, must exist in `asset_item_types`
- `quantityDesired` - Required, must be positive (> 0)
- `maxPricePerUnit` - Required, must be non-negative (>= 0)
- `notes` - Optional, text

**Response:** `200 OK`

```json
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
```

**Error Responses:**

| Status | Error | Reason |
|--------|-------|--------|
| 400 | `quantityDesired must be positive` | Quantity <= 0 |
| 400 | `maxPricePerUnit must be non-negative` | Price < 0 |
| 400 | `typeId is required` | Missing typeId |
| 500 | `failed to create buy order` | Database error |

**Example:**

```bash
curl -X POST http://localhost/api/v1/buy-orders \
  -H "USER-ID: 12345" \
  -H "BACKEND-KEY: your-backend-key" \
  -H "Content-Type: application/json" \
  -d '{
    "typeId": 34,
    "quantityDesired": 1000000,
    "maxPricePerUnit": 6,
    "notes": "Urgent need"
  }'
```

---

### 3. Update Buy Order

Update an existing buy order. You can only update your own buy orders.

**Endpoint:** `PUT /buy-orders/{id}`

**URL Parameters:**
- `id` - Buy order ID

**Request Body:**

```json
{
  "quantityDesired": 1500000,
  "maxPricePerUnit": 7,
  "notes": "Updated quantity",
  "isActive": true
}
```

**Note:** You cannot change the `typeId` of an existing buy order.

**Response:** `200 OK`

```json
{
  "id": 1,
  "buyerUserId": 12345,
  "typeId": 34,
  "typeName": "Tritanium",
  "quantityDesired": 1500000,
  "maxPricePerUnit": 7,
  "notes": "Updated quantity",
  "isActive": true,
  "createdAt": "2026-02-15T10:30:00Z",
  "updatedAt": "2026-02-15T11:00:00Z"
}
```

**Error Responses:**

| Status | Error | Reason |
|--------|-------|--------|
| 400 | `quantityDesired must be positive` | Quantity <= 0 |
| 400 | `invalid buy order ID` | Invalid ID format |
| 403 | `you do not own this buy order` | Trying to update another user's order |
| 404 | `buy order not found` | Order doesn't exist |

**Example:**

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

---

### 4. Delete Buy Order

Soft-delete a buy order (sets `is_active = false`).

**Endpoint:** `DELETE /buy-orders/{id}`

**URL Parameters:**
- `id` - Buy order ID

**Response:** `200 OK`

```json
{
  "id": 1,
  "buyerUserId": 12345,
  "typeId": 34,
  "typeName": "Tritanium",
  "quantityDesired": 1500000,
  "maxPricePerUnit": 7,
  "notes": "Updated quantity",
  "isActive": false,
  "createdAt": "2026-02-15T10:30:00Z",
  "updatedAt": "2026-02-15T12:00:00Z"
}
```

**Error Responses:**

| Status | Error | Reason |
|--------|-------|--------|
| 400 | `invalid buy order ID` | Invalid ID format |
| 403 | `you do not own this buy order` | Trying to delete another user's order |
| 404 | `buy order not found` | Order doesn't exist |

**Example:**

```bash
curl -X DELETE http://localhost/api/v1/buy-orders/1 \
  -H "USER-ID: 12345" \
  -H "BACKEND-KEY: your-backend-key"
```

---

### 5. Get Demand From Contacts

Get all active buy orders from contacts who have granted you `for_sale_browse` permission. This is the "seller view" - showing what your contacts want to buy.

**Endpoint:** `GET /buy-orders/demand`

**Permission Required:** Contacts must have granted you `for_sale_browse` permission

**Response:** `200 OK`

```json
[
  {
    "id": 1,
    "buyerUserId": 67890,
    "typeId": 34,
    "typeName": "Tritanium",
    "quantityDesired": 1000000,
    "maxPricePerUnit": 6,
    "notes": "Urgent need for manufacturing",
    "isActive": true,
    "createdAt": "2026-02-15T10:30:00Z",
    "updatedAt": "2026-02-15T10:30:00Z"
  },
  {
    "id": 2,
    "buyerUserId": 67890,
    "typeId": 35,
    "typeName": "Pyerite",
    "quantityDesired": 500000,
    "maxPricePerUnit": 15,
    "notes": null,
    "isActive": true,
    "createdAt": "2026-02-15T11:00:00Z",
    "updatedAt": "2026-02-15T11:00:00Z"
  }
]
```

**Note:** Only returns buy orders from contacts who:
1. Have accepted your contact request
2. Have granted you `for_sale_browse` permission
3. Have active buy orders (`is_active = true`)

**Example:**

```bash
curl -X GET http://localhost/api/v1/buy-orders/demand \
  -H "USER-ID: 12345" \
  -H "BACKEND-KEY: your-backend-key"
```

---

### 6. Search Item Types

Search for EVE Online item types by name. Used for autocomplete in buy order creation.

**Endpoint:** `GET /item-types/search`

**Query Parameters:**
- `q` - Search query (minimum 2 characters)

**Response:** `200 OK`

```json
[
  {
    "TypeID": 34,
    "TypeName": "Tritanium",
    "Volume": 0.01,
    "IconID": 22
  },
  {
    "TypeID": 11399,
    "TypeName": "Compressed Tritanium",
    "Volume": 0.15,
    "IconID": 22
  }
]
```

**Sorting:** Results are sorted by relevance:
1. Exact matches first (e.g., "Tritanium" for query "tritanium")
2. Items starting with query (e.g., "Tritanium" for query "trit")
3. Items containing query (e.g., "Compressed Tritanium" for query "trit")
4. Alphabetically within each group

**Limits:**
- Minimum query length: 2 characters
- Maximum results: 20 items

**Example:**

```bash
curl -X GET 'http://localhost/api/v1/item-types/search?q=trit' \
  -H "USER-ID: 12345" \
  -H "BACKEND-KEY: your-backend-key"
```

---

## Frontend API Routes

The Next.js frontend provides proxy routes to the backend:

### Frontend Routes

| Frontend Route | Backend Endpoint | Methods |
|---------------|------------------|---------|
| `/api/buy-orders` | `/v1/buy-orders` | GET, POST |
| `/api/buy-orders/[id]` | `/v1/buy-orders/{id}` | PUT, DELETE |
| `/api/buy-orders/demand` | `/v1/buy-orders/demand` | GET |
| `/api/item-types/search` | `/v1/item-types/search` | GET |

### Usage in Frontend

```typescript
// Get my buy orders
const response = await fetch('/api/buy-orders');
const orders = await response.json();

// Create a buy order
const response = await fetch('/api/buy-orders', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    typeId: 34,
    quantityDesired: 1000000,
    maxPricePerUnit: 6,
    notes: 'Urgent need'
  })
});

// Update a buy order
const response = await fetch(`/api/buy-orders/${orderId}`, {
  method: 'PUT',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    quantityDesired: 1500000,
    maxPricePerUnit: 7,
    isActive: true
  })
});

// Delete a buy order
const response = await fetch(`/api/buy-orders/${orderId}`, {
  method: 'DELETE'
});

// Get demand from contacts
const response = await fetch('/api/buy-orders/demand');
const demand = await response.json();

// Search for items
const response = await fetch(`/api/item-types/search?q=${query}`);
const items = await response.json();
```

---

## Error Handling

All endpoints follow consistent error response format:

```json
{
  "error": "Error message describing what went wrong"
}
```

### Common HTTP Status Codes

| Code | Meaning | When Used |
|------|---------|-----------|
| 200 | OK | Successful operation |
| 400 | Bad Request | Invalid input, validation failure |
| 401 | Unauthorized | Missing or invalid authentication |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 500 | Internal Server Error | Server-side error |

---

## Data Types

### BuyOrder

```typescript
type BuyOrder = {
  id: number;
  buyerUserId: number;
  typeId: number;
  typeName: string;
  quantityDesired: number;
  maxPricePerUnit: number;
  notes?: string;
  isActive: boolean;
  createdAt: string;  // ISO 8601 timestamp
  updatedAt: string;  // ISO 8601 timestamp
};
```

### ItemType

```typescript
type ItemType = {
  TypeID: number;
  TypeName: string;
  Volume: number;
  IconID?: number;
};
```

---

## Rate Limiting

Currently, there are no rate limits enforced. For production use, consider:
- Rate limiting per user
- Query result caching
- Debouncing autocomplete searches (frontend already implements 300ms debounce)

---

## Best Practices

1. **Autocomplete Debouncing**: Always debounce item search requests (300ms recommended)
2. **Validation**: Validate inputs on frontend before sending to API
3. **Error Handling**: Always handle error responses gracefully
4. **Permission Checks**: Verify permissions before attempting operations
5. **Optimistic UI**: Update UI optimistically, rollback on error
