package controllers_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/controllers"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/stretchr/testify/assert"
)

func setupPurchasesTestDB(t *testing.T) *sql.DB {
	db, err := setupDatabase()
	assert.NoError(t, err)

	// Create base data
	_, err = db.ExecContext(context.Background(),
		"INSERT INTO regions (region_id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		10000002, "The Forge")
	assert.NoError(t, err)

	_, err = db.ExecContext(context.Background(),
		"INSERT INTO constellations (constellation_id, name, region_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
		20000020, "Kimotoro", 10000002)
	assert.NoError(t, err)

	_, err = db.ExecContext(context.Background(),
		"INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING",
		30000142, "Jita", 20000020, 0.9)
	assert.NoError(t, err)

	return db
}

func Test_PurchaseItem_Success(t *testing.T) {
	db := setupPurchasesTestDB(t)

	buyerID := int64(4000)
	sellerID := int64(4001)
	typeID := int64(50)

	// Setup users and data
	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	forSaleRepo := repositories.NewForSaleItems(db)
	permRepo := repositories.NewContactPermissions(db)
	contactsRepo := repositories.NewContacts(db)

	buyer := &repositories.User{ID: buyerID, Name: "Buyer"}
	seller := &repositories.User{ID: sellerID, Name: "Seller"}
	assert.NoError(t, userRepo.Add(context.Background(), buyer))
	assert.NoError(t, userRepo.Add(context.Background(), seller))

	buyerChar := &repositories.Character{ID: buyerID * 10, Name: "Buyer Char", UserID: buyerID}
	sellerChar := &repositories.Character{ID: sellerID * 10, Name: "Seller Char", UserID: sellerID}
	assert.NoError(t, charRepo.Add(context.Background(), buyerChar))
	assert.NoError(t, charRepo.Add(context.Background(), sellerChar))

	itemTypes := []models.EveInventoryType{{TypeID: typeID, TypeName: "Tritanium", Volume: 0.01}}
	assert.NoError(t, itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes))

	// Create contact with permission
	contact, err := contactsRepo.Create(context.Background(), buyerID, sellerID)
	assert.NoError(t, err)
	_, err = contactsRepo.UpdateStatus(context.Background(), contact.ID, sellerID, "accepted")
	assert.NoError(t, err)

	perm := &models.ContactPermission{
		ContactID:       contact.ID,
		GrantingUserID:  sellerID,
		ReceivingUserID: buyerID,
		ServiceType:     "for_sale_browse",
		CanAccess:       true,
	}
	assert.NoError(t, permRepo.Upsert(context.Background(), perm))

	// Create for-sale item
	item := &models.ForSaleItem{
		UserID:            sellerID,
		TypeID:            typeID,
		OwnerType:         "character",
		OwnerID:           sellerID * 10,
		LocationID:        30000142,
		QuantityAvailable: 1000,
		PricePerUnit:      100,
		IsActive:          true,
	}
	assert.NoError(t, forSaleRepo.Upsert(context.Background(), item))

	// Setup controller
	purchaseRepo := repositories.NewPurchaseTransactions(db)
	controller := controllers.NewPurchases(&MockRouter{}, db, purchaseRepo, forSaleRepo, permRepo, userRepo, nil, nil)

	// Make purchase request
	reqBody := map[string]interface{}{
		"forSaleItemId":     item.ID,
		"quantityPurchased": 250,
		"notes":             "Test purchase",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/purchases", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		Params:  map[string]string{},
		User:    &buyerID,
	}

	// Execute
	result, httpErr := controller.PurchaseItem(args)

	// Verify
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	purchase := result.(*models.PurchaseTransaction)
	assert.Equal(t, item.ID, purchase.ForSaleItemID)
	assert.Equal(t, buyerID, purchase.BuyerUserID)
	assert.Equal(t, sellerID, purchase.SellerUserID)
	assert.Equal(t, int64(250), purchase.QuantityPurchased)
	assert.Equal(t, float64(25000), purchase.TotalPrice)
	assert.Equal(t, "pending", purchase.Status)

	// Verify quantity was reduced
	updatedItem, err := forSaleRepo.GetByID(context.Background(), item.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(750), updatedItem.QuantityAvailable)
}

func Test_PurchaseItem_EntireQuantity_MarksInactive(t *testing.T) {
	db := setupPurchasesTestDB(t)

	buyerID := int64(4010)
	sellerID := int64(4011)
	typeID := int64(51)

	// Setup (similar to previous test)
	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	forSaleRepo := repositories.NewForSaleItems(db)
	permRepo := repositories.NewContactPermissions(db)
	contactsRepo := repositories.NewContacts(db)

	buyer := &repositories.User{ID: buyerID, Name: "Buyer"}
	seller := &repositories.User{ID: sellerID, Name: "Seller"}
	assert.NoError(t, userRepo.Add(context.Background(), buyer))
	assert.NoError(t, userRepo.Add(context.Background(), seller))

	buyerChar := &repositories.Character{ID: buyerID * 10, Name: "Buyer Char", UserID: buyerID}
	sellerChar := &repositories.Character{ID: sellerID * 10, Name: "Seller Char", UserID: sellerID}
	assert.NoError(t, charRepo.Add(context.Background(), buyerChar))
	assert.NoError(t, charRepo.Add(context.Background(), sellerChar))

	itemTypes := []models.EveInventoryType{{TypeID: typeID, TypeName: "Pyerite", Volume: 0.01}}
	assert.NoError(t, itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes))

	contact, err := contactsRepo.Create(context.Background(), buyerID, sellerID)
	assert.NoError(t, err)
	_, err = contactsRepo.UpdateStatus(context.Background(), contact.ID, sellerID, "accepted")
	assert.NoError(t, err)

	perm := &models.ContactPermission{
		ContactID:       contact.ID,
		GrantingUserID:  sellerID,
		ReceivingUserID: buyerID,
		ServiceType:     "for_sale_browse",
		CanAccess:       true,
	}
	assert.NoError(t, permRepo.Upsert(context.Background(), perm))

	item := &models.ForSaleItem{
		UserID:            sellerID,
		TypeID:            typeID,
		OwnerType:         "character",
		OwnerID:           sellerID * 10,
		LocationID:        30000142,
		QuantityAvailable: 100,
		PricePerUnit:      200,
		IsActive:          true,
	}
	assert.NoError(t, forSaleRepo.Upsert(context.Background(), item))

	purchaseRepo := repositories.NewPurchaseTransactions(db)
	controller := controllers.NewPurchases(&MockRouter{}, db, purchaseRepo, forSaleRepo, permRepo, userRepo, nil, nil)

	// Purchase entire quantity
	reqBody := map[string]interface{}{
		"forSaleItemId":     item.ID,
		"quantityPurchased": 100,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/purchases", bytes.NewReader(body))

	args := &web.HandlerArgs{
		Request: req,
		User:    &buyerID,
	}

	result, httpErr := controller.PurchaseItem(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	// Verify item is now inactive
	updatedItem, err := forSaleRepo.GetByID(context.Background(), item.ID)
	assert.NoError(t, err)
	assert.False(t, updatedItem.IsActive)
}

func Test_PurchaseItem_NoPermission(t *testing.T) {
	db := setupPurchasesTestDB(t)

	buyerID := int64(4020)
	sellerID := int64(4021)
	typeID := int64(52)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	forSaleRepo := repositories.NewForSaleItems(db)
	permRepo := repositories.NewContactPermissions(db)

	buyer := &repositories.User{ID: buyerID, Name: "Buyer"}
	seller := &repositories.User{ID: sellerID, Name: "Seller"}
	assert.NoError(t, userRepo.Add(context.Background(), buyer))
	assert.NoError(t, userRepo.Add(context.Background(), seller))

	buyerChar := &repositories.Character{ID: buyerID * 10, Name: "Buyer Char", UserID: buyerID}
	sellerChar := &repositories.Character{ID: sellerID * 10, Name: "Seller Char", UserID: sellerID}
	assert.NoError(t, charRepo.Add(context.Background(), buyerChar))
	assert.NoError(t, charRepo.Add(context.Background(), sellerChar))

	itemTypes := []models.EveInventoryType{{TypeID: typeID, TypeName: "Mexallon", Volume: 0.01}}
	assert.NoError(t, itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes))

	item := &models.ForSaleItem{
		UserID:            sellerID,
		TypeID:            typeID,
		OwnerType:         "character",
		OwnerID:           sellerID * 10,
		LocationID:        30000142,
		QuantityAvailable: 500,
		PricePerUnit:      150,
		IsActive:          true,
	}
	assert.NoError(t, forSaleRepo.Upsert(context.Background(), item))

	purchaseRepo := repositories.NewPurchaseTransactions(db)
	controller := controllers.NewPurchases(&MockRouter{}, db, purchaseRepo, forSaleRepo, permRepo, userRepo, nil, nil)

	reqBody := map[string]interface{}{
		"forSaleItemId":     item.ID,
		"quantityPurchased": 50,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/purchases", bytes.NewReader(body))

	args := &web.HandlerArgs{
		Request: req,
		User:    &buyerID,
	}

	result, httpErr := controller.PurchaseItem(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 403, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "permission")
}

func Test_PurchaseItem_SelfPurchase_Rejected(t *testing.T) {
	db := setupPurchasesTestDB(t)

	userID := int64(4030)
	typeID := int64(53)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	forSaleRepo := repositories.NewForSaleItems(db)
	permRepo := repositories.NewContactPermissions(db)

	user := &repositories.User{ID: userID, Name: "User"}
	assert.NoError(t, userRepo.Add(context.Background(), user))

	char := &repositories.Character{ID: userID * 10, Name: "Char", UserID: userID}
	assert.NoError(t, charRepo.Add(context.Background(), char))

	itemTypes := []models.EveInventoryType{{TypeID: typeID, TypeName: "Isogen", Volume: 0.01}}
	assert.NoError(t, itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes))

	item := &models.ForSaleItem{
		UserID:            userID,
		TypeID:            typeID,
		OwnerType:         "character",
		OwnerID:           userID * 10,
		LocationID:        30000142,
		QuantityAvailable: 200,
		PricePerUnit:      120,
		IsActive:          true,
	}
	assert.NoError(t, forSaleRepo.Upsert(context.Background(), item))

	purchaseRepo := repositories.NewPurchaseTransactions(db)
	controller := controllers.NewPurchases(&MockRouter{}, db, purchaseRepo, forSaleRepo, permRepo, userRepo, nil, nil)

	reqBody := map[string]interface{}{
		"forSaleItemId":     item.ID,
		"quantityPurchased": 50,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/purchases", bytes.NewReader(body))

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.PurchaseItem(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	// Note: Self-purchase is blocked by permission check (403) before the self-purchase validation (400)
	// since a user cannot create a contact with themselves
	assert.Equal(t, 403, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "permission")
}

func Test_PurchaseItem_QuantityExceeded(t *testing.T) {
	db := setupPurchasesTestDB(t)

	buyerID := int64(4040)
	sellerID := int64(4041)
	typeID := int64(54)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	forSaleRepo := repositories.NewForSaleItems(db)
	permRepo := repositories.NewContactPermissions(db)
	contactsRepo := repositories.NewContacts(db)

	buyer := &repositories.User{ID: buyerID, Name: "Buyer"}
	seller := &repositories.User{ID: sellerID, Name: "Seller"}
	assert.NoError(t, userRepo.Add(context.Background(), buyer))
	assert.NoError(t, userRepo.Add(context.Background(), seller))

	buyerChar := &repositories.Character{ID: buyerID * 10, Name: "Buyer Char", UserID: buyerID}
	sellerChar := &repositories.Character{ID: sellerID * 10, Name: "Seller Char", UserID: sellerID}
	assert.NoError(t, charRepo.Add(context.Background(), buyerChar))
	assert.NoError(t, charRepo.Add(context.Background(), sellerChar))

	itemTypes := []models.EveInventoryType{{TypeID: typeID, TypeName: "Nocxium", Volume: 0.01}}
	assert.NoError(t, itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes))

	contact, err := contactsRepo.Create(context.Background(), buyerID, sellerID)
	assert.NoError(t, err)
	_, err = contactsRepo.UpdateStatus(context.Background(), contact.ID, sellerID, "accepted")
	assert.NoError(t, err)

	perm := &models.ContactPermission{
		ContactID:       contact.ID,
		GrantingUserID:  sellerID,
		ReceivingUserID: buyerID,
		ServiceType:     "for_sale_browse",
		CanAccess:       true,
	}
	assert.NoError(t, permRepo.Upsert(context.Background(), perm))

	item := &models.ForSaleItem{
		UserID:            sellerID,
		TypeID:            typeID,
		OwnerType:         "character",
		OwnerID:           sellerID * 10,
		LocationID:        30000142,
		QuantityAvailable: 100,
		PricePerUnit:      300,
		IsActive:          true,
	}
	assert.NoError(t, forSaleRepo.Upsert(context.Background(), item))

	purchaseRepo := repositories.NewPurchaseTransactions(db)
	controller := controllers.NewPurchases(&MockRouter{}, db, purchaseRepo, forSaleRepo, permRepo, userRepo, nil, nil)

	reqBody := map[string]interface{}{
		"forSaleItemId":     item.ID,
		"quantityPurchased": 150, // More than available
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/purchases", bytes.NewReader(body))

	args := &web.HandlerArgs{
		Request: req,
		User:    &buyerID,
	}

	result, httpErr := controller.PurchaseItem(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "exceeds available quantity")
}

func Test_MarkContractCreated_Success(t *testing.T) {
	db := setupPurchasesTestDB(t)

	buyerID := int64(4050)
	sellerID := int64(4051)
	typeID := int64(55)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	forSaleRepo := repositories.NewForSaleItems(db)
	permRepo := repositories.NewContactPermissions(db)
	purchaseRepo := repositories.NewPurchaseTransactions(db)

	buyer := &repositories.User{ID: buyerID, Name: "Buyer"}
	seller := &repositories.User{ID: sellerID, Name: "Seller"}
	assert.NoError(t, userRepo.Add(context.Background(), buyer))
	assert.NoError(t, userRepo.Add(context.Background(), seller))

	buyerChar := &repositories.Character{ID: buyerID * 10, Name: "Buyer Char", UserID: buyerID}
	sellerChar := &repositories.Character{ID: sellerID * 10, Name: "Seller Char", UserID: sellerID}
	assert.NoError(t, charRepo.Add(context.Background(), buyerChar))
	assert.NoError(t, charRepo.Add(context.Background(), sellerChar))

	itemTypes := []models.EveInventoryType{{TypeID: typeID, TypeName: "Zydrine", Volume: 0.01}}
	assert.NoError(t, itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes))

	item := &models.ForSaleItem{
		UserID:            sellerID,
		TypeID:            typeID,
		OwnerType:         "character",
		OwnerID:           sellerID * 10,
		LocationID:        30000142,
		QuantityAvailable: 500,
		PricePerUnit:      400,
		IsActive:          true,
	}
	assert.NoError(t, forSaleRepo.Upsert(context.Background(), item))

	// Create a purchase
	tx, err := db.BeginTx(context.Background(), nil)
	assert.NoError(t, err)
	defer tx.Rollback()

	purchase := &models.PurchaseTransaction{
		ForSaleItemID:     item.ID,
		BuyerUserID:       buyerID,
		SellerUserID:      sellerID,
		TypeID:            typeID,
		QuantityPurchased: 100,
		PricePerUnit:      400,
		TotalPrice:        40000,
		Status:            "pending",
	}
	assert.NoError(t, purchaseRepo.Create(context.Background(), tx, purchase))
	assert.NoError(t, tx.Commit())

	controller := controllers.NewPurchases(&MockRouter{}, db, purchaseRepo, forSaleRepo, permRepo, userRepo, nil, nil)

	// Mark as contract created
	contractKey := "PT-4050-30000142-1234567890"
	reqBody := map[string]interface{}{
		"contractKey": contractKey,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/purchases/"+strconv.FormatInt(purchase.ID, 10)+"/mark-contract-created", bytes.NewReader(body))

	args := &web.HandlerArgs{
		Request: req,
		Params:  map[string]string{"id": strconv.FormatInt(purchase.ID, 10)},
		User:    &sellerID, // Seller is marking
	}

	result, httpErr := controller.MarkContractCreated(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	// Verify status changed
	updated, err := purchaseRepo.GetByID(context.Background(), purchase.ID)
	assert.NoError(t, err)
	assert.Equal(t, "contract_created", updated.Status)
	assert.Equal(t, contractKey, *updated.ContractKey)
}

func Test_CompletePurchase_Success(t *testing.T) {
	db := setupPurchasesTestDB(t)

	buyerID := int64(4060)
	sellerID := int64(4061)
	typeID := int64(56)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	forSaleRepo := repositories.NewForSaleItems(db)
	permRepo := repositories.NewContactPermissions(db)
	purchaseRepo := repositories.NewPurchaseTransactions(db)

	buyer := &repositories.User{ID: buyerID, Name: "Buyer"}
	seller := &repositories.User{ID: sellerID, Name: "Seller"}
	assert.NoError(t, userRepo.Add(context.Background(), buyer))
	assert.NoError(t, userRepo.Add(context.Background(), seller))

	buyerChar := &repositories.Character{ID: buyerID * 10, Name: "Buyer Char", UserID: buyerID}
	sellerChar := &repositories.Character{ID: sellerID * 10, Name: "Seller Char", UserID: sellerID}
	assert.NoError(t, charRepo.Add(context.Background(), buyerChar))
	assert.NoError(t, charRepo.Add(context.Background(), sellerChar))

	itemTypes := []models.EveInventoryType{{TypeID: typeID, TypeName: "Megacyte", Volume: 0.01}}
	assert.NoError(t, itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes))

	item := &models.ForSaleItem{
		UserID:            sellerID,
		TypeID:            typeID,
		OwnerType:         "character",
		OwnerID:           sellerID * 10,
		LocationID:        30000142,
		QuantityAvailable: 600,
		PricePerUnit:      500,
		IsActive:          true,
	}
	assert.NoError(t, forSaleRepo.Upsert(context.Background(), item))

	tx, err := db.BeginTx(context.Background(), nil)
	assert.NoError(t, err)
	defer tx.Rollback()

	purchase := &models.PurchaseTransaction{
		ForSaleItemID:     item.ID,
		BuyerUserID:       buyerID,
		SellerUserID:      sellerID,
		TypeID:            typeID,
		QuantityPurchased: 150,
		PricePerUnit:      500,
		TotalPrice:        75000,
		Status:            "contract_created",
	}
	assert.NoError(t, purchaseRepo.Create(context.Background(), tx, purchase))
	assert.NoError(t, tx.Commit())

	controller := controllers.NewPurchases(&MockRouter{}, db, purchaseRepo, forSaleRepo, permRepo, userRepo, nil, nil)

	req := httptest.NewRequest("POST", "/v1/purchases/"+strconv.FormatInt(purchase.ID, 10)+"/complete", nil)
	args := &web.HandlerArgs{
		Request: req,
		Params:  map[string]string{"id": strconv.FormatInt(purchase.ID, 10)},
		User:    &buyerID, // Buyer is completing
	}

	result, httpErr := controller.CompletePurchase(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	// Verify status changed to completed
	updated, err := purchaseRepo.GetByID(context.Background(), purchase.ID)
	assert.NoError(t, err)
	assert.Equal(t, "completed", updated.Status)
}

func Test_CancelPurchase_RestoresQuantity(t *testing.T) {
	db := setupPurchasesTestDB(t)

	buyerID := int64(4070)
	sellerID := int64(4071)
	typeID := int64(57)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	forSaleRepo := repositories.NewForSaleItems(db)
	permRepo := repositories.NewContactPermissions(db)
	purchaseRepo := repositories.NewPurchaseTransactions(db)
	contactsRepo := repositories.NewContacts(db)

	buyer := &repositories.User{ID: buyerID, Name: "Buyer"}
	seller := &repositories.User{ID: sellerID, Name: "Seller"}
	assert.NoError(t, userRepo.Add(context.Background(), buyer))
	assert.NoError(t, userRepo.Add(context.Background(), seller))

	buyerChar := &repositories.Character{ID: buyerID * 10, Name: "Buyer Char", UserID: buyerID}
	sellerChar := &repositories.Character{ID: sellerID * 10, Name: "Seller Char", UserID: sellerID}
	assert.NoError(t, charRepo.Add(context.Background(), buyerChar))
	assert.NoError(t, charRepo.Add(context.Background(), sellerChar))

	itemTypes := []models.EveInventoryType{{TypeID: typeID, TypeName: "Morphite", Volume: 0.01}}
	assert.NoError(t, itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes))

	contact, err := contactsRepo.Create(context.Background(), buyerID, sellerID)
	assert.NoError(t, err)
	_, err = contactsRepo.UpdateStatus(context.Background(), contact.ID, sellerID, "accepted")
	assert.NoError(t, err)

	perm := &models.ContactPermission{
		ContactID:       contact.ID,
		GrantingUserID:  sellerID,
		ReceivingUserID: buyerID,
		ServiceType:     "for_sale_browse",
		CanAccess:       true,
	}
	assert.NoError(t, permRepo.Upsert(context.Background(), perm))

	item := &models.ForSaleItem{
		UserID:            sellerID,
		TypeID:            typeID,
		OwnerType:         "character",
		OwnerID:           sellerID * 10,
		LocationID:        30000142,
		QuantityAvailable: 1000,
		PricePerUnit:      600,
		IsActive:          true,
	}
	assert.NoError(t, forSaleRepo.Upsert(context.Background(), item))

	// Make a purchase first
	controller := controllers.NewPurchases(&MockRouter{}, db, purchaseRepo, forSaleRepo, permRepo, userRepo, nil, nil)

	reqBody := map[string]interface{}{
		"forSaleItemId":     item.ID,
		"quantityPurchased": 200,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/purchases", bytes.NewReader(body))
	args := &web.HandlerArgs{
		Request: req,
		User:    &buyerID,
	}

	result, httpErr := controller.PurchaseItem(args)
	assert.Nil(t, httpErr)
	purchase := result.(*models.PurchaseTransaction)

	// Verify quantity reduced
	updatedItem, err := forSaleRepo.GetByID(context.Background(), item.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(800), updatedItem.QuantityAvailable)

	// Cancel the purchase
	req = httptest.NewRequest("POST", "/v1/purchases/"+strconv.FormatInt(purchase.ID, 10)+"/cancel", nil)
	args = &web.HandlerArgs{
		Request: req,
		Params:  map[string]string{"id": strconv.FormatInt(purchase.ID, 10)},
		User:    &buyerID,
	}

	result, httpErr = controller.CancelPurchase(args)
	assert.Nil(t, httpErr)

	// Verify quantity restored
	restoredItem, err := forSaleRepo.GetByID(context.Background(), item.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1000), restoredItem.QuantityAvailable)

	// Verify purchase status
	cancelledPurchase, err := purchaseRepo.GetByID(context.Background(), purchase.ID)
	assert.NoError(t, err)
	assert.Equal(t, "cancelled", cancelledPurchase.Status)
}

func Test_MarkContractCreated_AutoGeneratesKey(t *testing.T) {
	db := setupPurchasesTestDB(t)

	buyerID := int64(4080)
	sellerID := int64(4081)
	typeID := int64(58)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	forSaleRepo := repositories.NewForSaleItems(db)
	permRepo := repositories.NewContactPermissions(db)
	purchaseRepo := repositories.NewPurchaseTransactions(db)

	buyer := &repositories.User{ID: buyerID, Name: "Buyer"}
	seller := &repositories.User{ID: sellerID, Name: "Seller"}
	assert.NoError(t, userRepo.Add(context.Background(), buyer))
	assert.NoError(t, userRepo.Add(context.Background(), seller))

	buyerChar := &repositories.Character{ID: buyerID * 10, Name: "Buyer Char", UserID: buyerID}
	sellerChar := &repositories.Character{ID: sellerID * 10, Name: "Seller Char", UserID: sellerID}
	assert.NoError(t, charRepo.Add(context.Background(), buyerChar))
	assert.NoError(t, charRepo.Add(context.Background(), sellerChar))

	itemTypes := []models.EveInventoryType{{TypeID: typeID, TypeName: "Veldspar", Volume: 0.1}}
	assert.NoError(t, itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes))

	item := &models.ForSaleItem{
		UserID:            sellerID,
		TypeID:            typeID,
		OwnerType:         "character",
		OwnerID:           sellerID * 10,
		LocationID:        30000142,
		QuantityAvailable: 500,
		PricePerUnit:      10,
		IsActive:          true,
	}
	assert.NoError(t, forSaleRepo.Upsert(context.Background(), item))

	tx, err := db.BeginTx(context.Background(), nil)
	assert.NoError(t, err)
	defer tx.Rollback()

	purchase := &models.PurchaseTransaction{
		ForSaleItemID:     item.ID,
		BuyerUserID:       buyerID,
		SellerUserID:      sellerID,
		TypeID:            typeID,
		QuantityPurchased: 100,
		PricePerUnit:      10,
		TotalPrice:        1000,
		Status:            "pending",
	}
	assert.NoError(t, purchaseRepo.Create(context.Background(), tx, purchase))
	assert.NoError(t, tx.Commit())

	controller := controllers.NewPurchases(&MockRouter{}, db, purchaseRepo, forSaleRepo, permRepo, userRepo, nil, nil)

	// Mark as contract created WITHOUT providing a contractKey (empty body)
	req := httptest.NewRequest("POST", "/v1/purchases/"+strconv.FormatInt(purchase.ID, 10)+"/mark-contract-created", nil)
	args := &web.HandlerArgs{
		Request: req,
		Params:  map[string]string{"id": strconv.FormatInt(purchase.ID, 10)},
		User:    &sellerID,
	}

	result, httpErr := controller.MarkContractCreated(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	// Verify auto-generated key format is PT-{purchaseID}
	resultMap := result.(map[string]string)
	expectedKey := "PT-" + strconv.FormatInt(purchase.ID, 10)
	assert.Equal(t, expectedKey, resultMap["contractKey"])
	assert.Equal(t, "contract_created", resultMap["status"])

	// Verify key was persisted
	updated, err := purchaseRepo.GetByID(context.Background(), purchase.ID)
	assert.NoError(t, err)
	assert.Equal(t, "contract_created", updated.Status)
	assert.Equal(t, expectedKey, *updated.ContractKey)
}

func Test_MarkContractCreated_UsesProvidedKey(t *testing.T) {
	db := setupPurchasesTestDB(t)

	buyerID := int64(4090)
	sellerID := int64(4091)
	typeID := int64(59)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	forSaleRepo := repositories.NewForSaleItems(db)
	permRepo := repositories.NewContactPermissions(db)
	purchaseRepo := repositories.NewPurchaseTransactions(db)

	buyer := &repositories.User{ID: buyerID, Name: "Buyer"}
	seller := &repositories.User{ID: sellerID, Name: "Seller"}
	assert.NoError(t, userRepo.Add(context.Background(), buyer))
	assert.NoError(t, userRepo.Add(context.Background(), seller))

	buyerChar := &repositories.Character{ID: buyerID * 10, Name: "Buyer Char", UserID: buyerID}
	sellerChar := &repositories.Character{ID: sellerID * 10, Name: "Seller Char", UserID: sellerID}
	assert.NoError(t, charRepo.Add(context.Background(), buyerChar))
	assert.NoError(t, charRepo.Add(context.Background(), sellerChar))

	itemTypes := []models.EveInventoryType{{TypeID: typeID, TypeName: "Scordite", Volume: 0.15}}
	assert.NoError(t, itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes))

	item := &models.ForSaleItem{
		UserID:            sellerID,
		TypeID:            typeID,
		OwnerType:         "character",
		OwnerID:           sellerID * 10,
		LocationID:        30000142,
		QuantityAvailable: 800,
		PricePerUnit:      15,
		IsActive:          true,
	}
	assert.NoError(t, forSaleRepo.Upsert(context.Background(), item))

	tx, err := db.BeginTx(context.Background(), nil)
	assert.NoError(t, err)
	defer tx.Rollback()

	purchase := &models.PurchaseTransaction{
		ForSaleItemID:     item.ID,
		BuyerUserID:       buyerID,
		SellerUserID:      sellerID,
		TypeID:            typeID,
		QuantityPurchased: 200,
		PricePerUnit:      15,
		TotalPrice:        3000,
		Status:            "pending",
	}
	assert.NoError(t, purchaseRepo.Create(context.Background(), tx, purchase))
	assert.NoError(t, tx.Commit())

	controller := controllers.NewPurchases(&MockRouter{}, db, purchaseRepo, forSaleRepo, permRepo, userRepo, nil, nil)

	// Mark as contract created WITH a custom contractKey
	customKey := "CUSTOM-BATCH-99"
	reqBody := map[string]interface{}{
		"contractKey": customKey,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/purchases/"+strconv.FormatInt(purchase.ID, 10)+"/mark-contract-created", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		Params:  map[string]string{"id": strconv.FormatInt(purchase.ID, 10)},
		User:    &sellerID,
	}

	result, httpErr := controller.MarkContractCreated(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	// Verify the custom key was used instead of auto-generated
	resultMap := result.(map[string]string)
	assert.Equal(t, customKey, resultMap["contractKey"])

	// Verify persisted
	updated, err := purchaseRepo.GetByID(context.Background(), purchase.ID)
	assert.NoError(t, err)
	assert.Equal(t, customKey, *updated.ContractKey)
}
