package controllers_test

import (
	"bytes"
	"context"
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

func Test_BuyOrders_CreateOrder_Success(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	userID := int64(6000)

	userRepo := repositories.NewUserRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	buyOrdersRepo := repositories.NewBuyOrders(db)
	permRepo := &MockContactPermissionsRepository{}

	user := &repositories.User{ID: userID, Name: "Test User"}
	assert.NoError(t, userRepo.Add(context.Background(), user))

	itemTypes := []models.EveInventoryType{
		{TypeID: 70, TypeName: "Tritanium", Volume: 0.01},
	}
	assert.NoError(t, itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes))

	controller := controllers.NewBuyOrders(&MockRouter{}, buyOrdersRepo, permRepo)

	reqBody := map[string]interface{}{
		"typeId":          70,
		"quantityDesired": 100000,
		"maxPricePerUnit": 6,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/buy-orders", bytes.NewReader(body))

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.CreateOrder(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	order := result.(*models.BuyOrder)
	assert.NotZero(t, order.ID)
	assert.Equal(t, userID, order.BuyerUserID)
	assert.Equal(t, int64(70), order.TypeID)
	assert.Equal(t, int64(100000), order.QuantityDesired)
	assert.Equal(t, float64(6), order.MaxPricePerUnit)
	assert.True(t, order.IsActive)
}

func Test_BuyOrders_CreateOrder_InvalidQuantity(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	userID := int64(6010)
	buyOrdersRepo := repositories.NewBuyOrders(db)
	permRepo := &MockContactPermissionsRepository{}

	controller := controllers.NewBuyOrders(&MockRouter{}, buyOrdersRepo, permRepo)

	reqBody := map[string]interface{}{
		"typeId":          70,
		"quantityDesired": -100,
		"maxPricePerUnit": 6,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/buy-orders", bytes.NewReader(body))

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.CreateOrder(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "quantityDesired must be positive")
}

func Test_BuyOrders_GetMyOrders(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	userID := int64(6020)

	userRepo := repositories.NewUserRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	buyOrdersRepo := repositories.NewBuyOrders(db)
	permRepo := &MockContactPermissionsRepository{}

	user := &repositories.User{ID: userID, Name: "Test User"}
	assert.NoError(t, userRepo.Add(context.Background(), user))

	itemTypes := []models.EveInventoryType{
		{TypeID: 71, TypeName: "Pyerite", Volume: 0.01},
		{TypeID: 72, TypeName: "Mexallon", Volume: 0.01},
	}
	assert.NoError(t, itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes))

	// Create some buy orders
	for i := 0; i < 3; i++ {
		order := &models.BuyOrder{
			BuyerUserID:     userID,
			TypeID:          71 + int64(i%2),
			QuantityDesired: int64(10000 * (i + 1)),
			MaxPricePerUnit: float64(10 + i),
			IsActive:        true,
		}
		assert.NoError(t, buyOrdersRepo.Create(context.Background(), order))
	}

	controller := controllers.NewBuyOrders(&MockRouter{}, buyOrdersRepo, permRepo)

	req := httptest.NewRequest("GET", "/v1/buy-orders", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.GetMyOrders(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	orders := result.([]*models.BuyOrder)
	assert.Len(t, orders, 3)
	assert.NotEmpty(t, orders[0].TypeName)
}

func Test_BuyOrders_UpdateOrder_Success(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	userID := int64(6030)

	userRepo := repositories.NewUserRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	buyOrdersRepo := repositories.NewBuyOrders(db)
	permRepo := &MockContactPermissionsRepository{}

	user := &repositories.User{ID: userID, Name: "Test User"}
	assert.NoError(t, userRepo.Add(context.Background(), user))

	itemTypes := []models.EveInventoryType{
		{TypeID: 73, TypeName: "Isogen", Volume: 0.01},
	}
	assert.NoError(t, itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes))

	// Create order
	order := &models.BuyOrder{
		BuyerUserID:     userID,
		TypeID:          73,
		QuantityDesired: 50000,
		MaxPricePerUnit: 20,
		IsActive:        true,
	}
	assert.NoError(t, buyOrdersRepo.Create(context.Background(), order))

	controller := controllers.NewBuyOrders(&MockRouter{}, buyOrdersRepo, permRepo)

	// Update order
	reqBody := map[string]interface{}{
		"quantityDesired": 75000,
		"maxPricePerUnit": 25,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/v1/buy-orders/"+strconv.FormatInt(order.ID, 10), bytes.NewReader(body))

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": strconv.FormatInt(order.ID, 10)},
	}

	result, httpErr := controller.UpdateOrder(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	updated := result.(*models.BuyOrder)
	assert.Equal(t, int64(75000), updated.QuantityDesired)
	assert.Equal(t, float64(25), updated.MaxPricePerUnit)
}

func Test_BuyOrders_UpdateOrder_NotOwner(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	ownerID := int64(6040)
	otherUserID := int64(6041)

	userRepo := repositories.NewUserRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	buyOrdersRepo := repositories.NewBuyOrders(db)
	permRepo := &MockContactPermissionsRepository{}

	// Create both users
	owner := &repositories.User{ID: ownerID, Name: "Owner"}
	assert.NoError(t, userRepo.Add(context.Background(), owner))

	otherUser := &repositories.User{ID: otherUserID, Name: "Other User"}
	assert.NoError(t, userRepo.Add(context.Background(), otherUser))

	itemTypes := []models.EveInventoryType{
		{TypeID: 74, TypeName: "Nocxium", Volume: 0.01},
	}
	assert.NoError(t, itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes))

	// Create order as owner
	order := &models.BuyOrder{
		BuyerUserID:     ownerID,
		TypeID:          74,
		QuantityDesired: 25000,
		MaxPricePerUnit: 30,
		IsActive:        true,
	}
	assert.NoError(t, buyOrdersRepo.Create(context.Background(), order))

	controller := controllers.NewBuyOrders(&MockRouter{}, buyOrdersRepo, permRepo)

	// Try to update as different user
	reqBody := map[string]interface{}{
		"quantityDesired": 50000,
		"maxPricePerUnit": 35,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/v1/buy-orders/"+strconv.FormatInt(order.ID, 10), bytes.NewReader(body))

	args := &web.HandlerArgs{
		Request: req,
		User:    &otherUserID,
		Params:  map[string]string{"id": strconv.FormatInt(order.ID, 10)},
	}

	result, httpErr := controller.UpdateOrder(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 403, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "you do not own this buy order")
}

func Test_BuyOrders_DeleteOrder_Success(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	userID := int64(6050)

	userRepo := repositories.NewUserRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	buyOrdersRepo := repositories.NewBuyOrders(db)
	permRepo := &MockContactPermissionsRepository{}

	user := &repositories.User{ID: userID, Name: "Test User"}
	assert.NoError(t, userRepo.Add(context.Background(), user))

	itemTypes := []models.EveInventoryType{
		{TypeID: 75, TypeName: "Zydrine", Volume: 0.01},
	}
	assert.NoError(t, itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes))

	// Create order
	order := &models.BuyOrder{
		BuyerUserID:     userID,
		TypeID:          75,
		QuantityDesired: 15000,
		MaxPricePerUnit: 40,
		IsActive:        true,
	}
	assert.NoError(t, buyOrdersRepo.Create(context.Background(), order))

	controller := controllers.NewBuyOrders(&MockRouter{}, buyOrdersRepo, permRepo)

	req := httptest.NewRequest("DELETE", "/v1/buy-orders/"+strconv.FormatInt(order.ID, 10), nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": strconv.FormatInt(order.ID, 10)},
	}

	result, httpErr := controller.DeleteOrder(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	// Verify soft delete
	retrieved, err := buyOrdersRepo.GetByID(context.Background(), order.ID)
	assert.NoError(t, err)
	assert.False(t, retrieved.IsActive)
}

func Test_BuyOrders_GetDemand(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	buyerID := int64(6060)
	sellerID := int64(6061)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	contactsRepo := repositories.NewContacts(db)
	permRepo := repositories.NewContactPermissions(db)
	buyOrdersRepo := repositories.NewBuyOrders(db)

	// Create users
	buyer := &repositories.User{ID: buyerID, Name: "Buyer"}
	assert.NoError(t, userRepo.Add(context.Background(), buyer))

	seller := &repositories.User{ID: sellerID, Name: "Seller"}
	assert.NoError(t, userRepo.Add(context.Background(), seller))

	// Create characters
	buyerChar := &repositories.Character{ID: 60600, Name: "Buyer Character", UserID: buyerID}
	assert.NoError(t, charRepo.Add(context.Background(), buyerChar))

	sellerChar := &repositories.Character{ID: 60610, Name: "Seller Character", UserID: sellerID}
	assert.NoError(t, charRepo.Add(context.Background(), sellerChar))

	// Create contact relationship
	contact, err := contactsRepo.Create(context.Background(), buyerID, sellerID)
	assert.NoError(t, err)

	// Accept contact
	_, err = contactsRepo.UpdateStatus(context.Background(), contact.ID, sellerID, "accepted")
	assert.NoError(t, err)

	// Grant permission
	perm := &models.ContactPermission{
		ContactID:       contact.ID,
		GrantingUserID:  buyerID,
		ReceivingUserID: sellerID,
		ServiceType:     "for_sale_browse",
		CanAccess:       true,
	}
	assert.NoError(t, permRepo.Upsert(context.Background(), perm))

	// Create item types
	itemTypes := []models.EveInventoryType{
		{TypeID: 76, TypeName: "Megacyte", Volume: 0.01},
	}
	assert.NoError(t, itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes))

	// Create buy orders
	for i := 0; i < 2; i++ {
		order := &models.BuyOrder{
			BuyerUserID:     buyerID,
			TypeID:          76,
			QuantityDesired: int64(5000 * (i + 1)),
			MaxPricePerUnit: float64(50 + i*10),
			IsActive:        true,
		}
		assert.NoError(t, buyOrdersRepo.Create(context.Background(), order))
	}

	controller := controllers.NewBuyOrders(&MockRouter{}, buyOrdersRepo, permRepo)

	req := httptest.NewRequest("GET", "/v1/buy-orders/demand", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &sellerID,
	}

	result, httpErr := controller.GetDemand(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	demand := result.([]*models.BuyOrder)
	assert.Len(t, demand, 2)

	for _, order := range demand {
		assert.Equal(t, buyerID, order.BuyerUserID)
		assert.True(t, order.IsActive)
		assert.NotEmpty(t, order.TypeName)
	}
}
