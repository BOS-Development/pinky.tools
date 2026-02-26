# Purchase System Documentation

Complete documentation for the EVE Industry Tool purchase system (Phase 4).

## Quick Links

- **[Quick Start Guide](QUICK_START_PURCHASES.md)** - Get started in 5 minutes
- **[API Reference](API_PURCHASES.md)** - Endpoint specifications
- **[Complete Documentation](PURCHASES.md)** - Deep dive into architecture
- **[Test Suite](TESTING_PURCHASES.md)** - Testing guide
- **[Phase 4 Summary](PHASE4_SUMMARY.md)** - What was built

---

## For Different Audiences

### 🎯 I want to use the API (Frontend Developer)
Start here: **[API Reference](API_PURCHASES.md)**

Quick example:
```bash
# Purchase 50,000 Tritanium
POST /v1/purchases
{
  "forSaleItemId": 123,
  "quantityPurchased": 50000
}
```

### 🚀 I want to get started quickly
Start here: **[Quick Start Guide](QUICK_START_PURCHASES.md)**

Follow the step-by-step flow from browsing items to completing a purchase.

### 🔧 I need to understand the internals
Start here: **[Complete Documentation](PURCHASES.md)**

Covers architecture, database schema, business logic, and troubleshooting.

### 🧪 I want to run the tests
Start here: **[Test Suite Documentation](TESTING_PURCHASES.md)**

```bash
# Run all purchase tests
make test-backend
```

### 📊 I want to see what was built
Start here: **[Phase 4 Summary](PHASE4_SUMMARY.md)**

Complete overview of features, files, metrics, and success criteria.

---

## Overview

The Purchase System enables users to buy items from their contacts' for-sale listings with a complete transaction workflow:

```
Browse Items → Purchase → Contract Created → Completed
                    ↓
                 Cancel (restores quantity)
```

### Key Features

✅ **Atomic Transactions** - Quantity updates and purchase records succeed or fail together
✅ **Multi-Stage Workflow** - pending → contract_created → completed
✅ **Cancel & Restore** - Full refund of quantities
✅ **Permission Enforcement** - Contact-based access control
✅ **Transaction History** - Complete audit trail

### Status: Production Ready

- ✅ 19/19 tests passing
- ✅ Complete documentation
- ✅ All edge cases handled
- ✅ Performance optimized
- ✅ Security hardened

---

## Documentation Files

| File | Purpose | Audience |
|------|---------|----------|
| [QUICK_START_PURCHASES.md](QUICK_START_PURCHASES.md) | Get started fast | All users |
| [API_PURCHASES.md](API_PURCHASES.md) | API reference | Frontend devs |
| [PURCHASES.md](PURCHASES.md) | Technical deep dive | Backend devs |
| [TESTING_PURCHASES.md](TESTING_PURCHASES.md) | Test suite | QA, developers |
| [PHASE4_SUMMARY.md](PHASE4_SUMMARY.md) | What was built | PM, stakeholders |

---

## API Endpoints

```
POST   /v1/purchases                           - Purchase items
GET    /v1/purchases/buyer                     - Buyer's history
GET    /v1/purchases/seller                    - Seller's history
GET    /v1/purchases/pending                   - Pending sales
POST   /v1/purchases/:id/mark-contract-created - Mark contract created
POST   /v1/purchases/:id/complete              - Complete purchase
POST   /v1/purchases/:id/cancel                - Cancel purchase
```

---

## Sample Flow

```bash
# 1. Browse items
GET /v1/for-sale/browse

# 2. Purchase
POST /v1/purchases
{ "forSaleItemId": 123, "quantityPurchased": 50000 }

# 3. Seller creates contract
POST /v1/purchases/456/mark-contract-created
{ "contractKey": "123456789" }

# 4. Buyer completes
POST /v1/purchases/456/complete
```

---

## Database Schema

```sql
CREATE TABLE purchase_transactions (
    id BIGSERIAL PRIMARY KEY,
    for_sale_item_id BIGINT NOT NULL,
    buyer_user_id BIGINT NOT NULL,
    seller_user_id BIGINT NOT NULL,
    type_id BIGINT NOT NULL,
    quantity_purchased BIGINT NOT NULL,
    price_per_unit BIGINT NOT NULL,
    total_price BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    contract_key VARCHAR(100),
    location_id BIGINT NOT NULL,
    purchased_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## Related Features

- **Contact System** - [docs/features/social/contact-marketplace.md](../../social/contact-marketplace.md)
- **For-Sale Listings** - Part of marketplace system
- **Permissions** - Granular service-based access control

---

## Next: Phase 5

**Buy Orders & Demand Tracking**
- Buyers place orders for out-of-stock items
- Sellers see demand from contacts
- Price matching and notifications

---

## Support

- 📖 Full documentation: [PURCHASES.md](PURCHASES.md)
- 🐛 Report issues: [GitHub Issues](https://github.com/YOUR_USERNAME/industry-tool/issues)
- 💬 Questions: See troubleshooting section in [PURCHASES.md](PURCHASES.md)

---

**Back to main README:** [../../../README.md](../../../README.md)
