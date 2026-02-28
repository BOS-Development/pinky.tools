package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/controllers"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock JobSlotRentalsRepository
type MockJobSlotRentalsRepository struct {
	mock.Mock
}

func (m *MockJobSlotRentalsRepository) CalculateSlotInventory(ctx context.Context, userID int64) ([]*models.CharacterSlotInventory, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CharacterSlotInventory), args.Error(1)
}

func (m *MockJobSlotRentalsRepository) GetByUser(ctx context.Context, userID int64) ([]*models.JobSlotRentalListing, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.JobSlotRentalListing), args.Error(1)
}

func (m *MockJobSlotRentalsRepository) GetBrowsableListings(ctx context.Context, viewerUserID int64, sellerUserIDs []int64) ([]*models.JobSlotRentalListing, error) {
	args := m.Called(ctx, viewerUserID, sellerUserIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.JobSlotRentalListing), args.Error(1)
}

func (m *MockJobSlotRentalsRepository) Create(ctx context.Context, listing *models.JobSlotRentalListing) error {
	args := m.Called(ctx, listing)
	return args.Error(0)
}

func (m *MockJobSlotRentalsRepository) Update(ctx context.Context, listing *models.JobSlotRentalListing) error {
	args := m.Called(ctx, listing)
	return args.Error(0)
}

func (m *MockJobSlotRentalsRepository) Delete(ctx context.Context, listingID int64, userID int64) error {
	args := m.Called(ctx, listingID, userID)
	return args.Error(0)
}

func (m *MockJobSlotRentalsRepository) GetByID(ctx context.Context, listingID int64) (*models.JobSlotRentalListing, error) {
	args := m.Called(ctx, listingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.JobSlotRentalListing), args.Error(1)
}

func (m *MockJobSlotRentalsRepository) CreateInterest(ctx context.Context, interest *models.JobSlotInterestRequest) error {
	args := m.Called(ctx, interest)
	return args.Error(0)
}

func (m *MockJobSlotRentalsRepository) GetInterestsByListing(ctx context.Context, listingID int64, sellerUserID int64) ([]*models.JobSlotInterestRequest, error) {
	args := m.Called(ctx, listingID, sellerUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.JobSlotInterestRequest), args.Error(1)
}

func (m *MockJobSlotRentalsRepository) GetInterestsByRequester(ctx context.Context, requesterUserID int64) ([]*models.JobSlotInterestRequest, error) {
	args := m.Called(ctx, requesterUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.JobSlotInterestRequest), args.Error(1)
}

func (m *MockJobSlotRentalsRepository) UpdateInterestStatus(ctx context.Context, interestID int64, userID int64, status string) error {
	args := m.Called(ctx, interestID, userID, status)
	return args.Error(0)
}

func (m *MockJobSlotRentalsRepository) GetReceivedInterests(ctx context.Context, userID int64) ([]*models.JobSlotInterestRequest, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.JobSlotInterestRequest), args.Error(1)
}

func Test_JobSlotRentalsController_GetSlotInventory_Success(t *testing.T) {
	mockRepo := new(MockJobSlotRentalsRepository)
	mockRouter := &MockRouter{}

	userID := int64(123)

	expectedInventory := []*models.CharacterSlotInventory{
		{
			CharacterID:   456,
			CharacterName: "Test Character",
			SlotsByActivity: map[string]*models.ActivitySlotInfo{
				"manufacturing": {
					ActivityType:   "manufacturing",
					SlotsMax:       11,
					SlotsInUse:     0,
					SlotsReserved:  0,
					SlotsAvailable: 11,
					SlotsListed:    0,
				},
			},
		},
	}

	mockRepo.On("CalculateSlotInventory", mock.Anything, userID).Return(expectedInventory, nil)

	req := httptest.NewRequest("GET", "/v1/job-slots/inventory", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewJobSlotRentals(mockRouter, mockRepo, &MockContactPermissionsRepository{})
	result, httpErr := controller.GetSlotInventory(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	inventory := result.([]*models.CharacterSlotInventory)
	assert.Len(t, inventory, 1)
	assert.Equal(t, "Test Character", inventory[0].CharacterName)

	mockRepo.AssertExpectations(t)
}

func Test_JobSlotRentalsController_GetMyListings_Success(t *testing.T) {
	mockRepo := new(MockJobSlotRentalsRepository)
	mockRouter := &MockRouter{}

	userID := int64(123)

	expectedListings := []*models.JobSlotRentalListing{
		{
			ID:            1,
			UserID:        userID,
			CharacterID:   456,
			CharacterName: "Test Character",
			ActivityType:  "manufacturing",
			SlotsListed:   2,
			PriceAmount:   100000,
			PricingUnit:   "per_slot_day",
			IsActive:      true,
		},
	}

	mockRepo.On("GetByUser", mock.Anything, userID).Return(expectedListings, nil)

	req := httptest.NewRequest("GET", "/v1/job-slots/listings", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewJobSlotRentals(mockRouter, mockRepo, &MockContactPermissionsRepository{})
	result, httpErr := controller.GetMyListings(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	listings := result.([]*models.JobSlotRentalListing)
	assert.Len(t, listings, 1)
	assert.Equal(t, "manufacturing", listings[0].ActivityType)

	mockRepo.AssertExpectations(t)
}

func Test_JobSlotRentalsController_CreateListing_Success(t *testing.T) {
	mockRepo := new(MockJobSlotRentalsRepository)
	mockRouter := &MockRouter{}

	userID := int64(123)
	charID := int64(456)

	// Mock slot inventory showing available slots
	inventory := []*models.CharacterSlotInventory{
		{
			CharacterID:   charID,
			CharacterName: "Test Character",
			SlotsByActivity: map[string]*models.ActivitySlotInfo{
				"manufacturing": {
					ActivityType:   "manufacturing",
					SlotsMax:       11,
					SlotsInUse:     0,
					SlotsReserved:  0,
					SlotsAvailable: 11,
					SlotsListed:    0,
				},
			},
		},
	}

	mockRepo.On("CalculateSlotInventory", mock.Anything, userID).Return(inventory, nil)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(listing *models.JobSlotRentalListing) bool {
		return listing.UserID == userID &&
			listing.CharacterID == charID &&
			listing.ActivityType == "manufacturing" &&
			listing.SlotsListed == 2
	})).Return(nil)

	body := map[string]interface{}{
		"characterId":  charID,
		"activityType": "manufacturing",
		"slotsListed":  2,
		"priceAmount":  100000,
		"pricingUnit":  "per_slot_day",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/job-slots/listings", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewJobSlotRentals(mockRouter, mockRepo, &MockContactPermissionsRepository{})
	result, httpErr := controller.CreateListing(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	mockRepo.AssertExpectations(t)
}

func Test_JobSlotRentalsController_CreateListing_InsufficientSlots(t *testing.T) {
	mockRepo := new(MockJobSlotRentalsRepository)
	mockRouter := &MockRouter{}

	userID := int64(123)
	charID := int64(456)

	// Mock slot inventory showing only 1 slot available
	inventory := []*models.CharacterSlotInventory{
		{
			CharacterID:   charID,
			CharacterName: "Test Character",
			SlotsByActivity: map[string]*models.ActivitySlotInfo{
				"manufacturing": {
					ActivityType:   "manufacturing",
					SlotsMax:       11,
					SlotsInUse:     8,
					SlotsReserved:  2,
					SlotsAvailable: 1,
					SlotsListed:    0,
				},
			},
		},
	}

	mockRepo.On("CalculateSlotInventory", mock.Anything, userID).Return(inventory, nil)

	body := map[string]interface{}{
		"characterId":  charID,
		"activityType": "manufacturing",
		"slotsListed":  2, // Requesting 2 but only 1 available
		"priceAmount":  100000,
		"pricingUnit":  "per_slot_day",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/job-slots/listings", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewJobSlotRentals(mockRouter, mockRepo, &MockContactPermissionsRepository{})
	result, httpErr := controller.CreateListing(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "not enough slots available")

	mockRepo.AssertExpectations(t)
}

func Test_JobSlotRentalsController_UpdateListing_Success(t *testing.T) {
	mockRepo := new(MockJobSlotRentalsRepository)
	mockRouter := &MockRouter{}

	userID := int64(123)
	listingID := int64(1)

	existingListing := &models.JobSlotRentalListing{
		ID:           listingID,
		UserID:       userID,
		CharacterID:  456,
		ActivityType: "manufacturing",
		SlotsListed:  2,
		PriceAmount:  100000,
		PricingUnit:  "per_slot_day",
		IsActive:     true,
	}

	mockRepo.On("GetByID", mock.Anything, listingID).Return(existingListing, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(listing *models.JobSlotRentalListing) bool {
		return listing.ID == listingID &&
			listing.SlotsListed == 3 &&
			listing.PriceAmount == 125000
	})).Return(nil)

	body := map[string]interface{}{
		"slotsListed": 3,
		"priceAmount": 125000,
		"pricingUnit": "per_slot_day",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/v1/job-slots/listings/1", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	}

	controller := controllers.NewJobSlotRentals(mockRouter, mockRepo, &MockContactPermissionsRepository{})
	result, httpErr := controller.UpdateListing(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	mockRepo.AssertExpectations(t)
}

func Test_JobSlotRentalsController_UpdateListing_NotOwner(t *testing.T) {
	mockRepo := new(MockJobSlotRentalsRepository)
	mockRouter := &MockRouter{}

	userID := int64(123)
	otherUserID := int64(999)
	listingID := int64(1)

	existingListing := &models.JobSlotRentalListing{
		ID:           listingID,
		UserID:       otherUserID, // Different owner
		CharacterID:  456,
		ActivityType: "manufacturing",
		SlotsListed:  2,
		PriceAmount:  100000,
		PricingUnit:  "per_slot_day",
		IsActive:     true,
	}

	mockRepo.On("GetByID", mock.Anything, listingID).Return(existingListing, nil)

	body := map[string]interface{}{
		"slotsListed": 3,
		"priceAmount": 125000,
		"pricingUnit": "per_slot_day",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/v1/job-slots/listings/1", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	}

	controller := controllers.NewJobSlotRentals(mockRouter, mockRepo, &MockContactPermissionsRepository{})
	result, httpErr := controller.UpdateListing(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 403, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}

func Test_JobSlotRentalsController_DeleteListing_Success(t *testing.T) {
	mockRepo := new(MockJobSlotRentalsRepository)
	mockRouter := &MockRouter{}

	userID := int64(123)
	listingID := int64(1)

	mockRepo.On("Delete", mock.Anything, listingID, userID).Return(nil)

	req := httptest.NewRequest("DELETE", "/v1/job-slots/listings/1", nil)

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	}

	controller := controllers.NewJobSlotRentals(mockRouter, mockRepo, &MockContactPermissionsRepository{})
	result, httpErr := controller.DeleteListing(args)

	assert.Nil(t, httpErr)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

func Test_JobSlotRentalsController_BrowseListings_Success(t *testing.T) {
	mockRepo := new(MockJobSlotRentalsRepository)
	mockPermRepo := new(MockContactPermissionsRepository)
	mockRouter := &MockRouter{}

	userID := int64(123)
	sellerUserIDs := []int64{200, 300}

	expectedListings := []*models.JobSlotRentalListing{
		{
			ID:            1,
			UserID:        200,
			CharacterID:   2000,
			CharacterName: "Seller Character",
			ActivityType:  "manufacturing",
			SlotsListed:   2,
			PriceAmount:   100000,
			PricingUnit:   "per_slot_day",
			IsActive:      true,
		},
	}

	mockPermRepo.On("GetUserPermissionsForService", mock.Anything, userID, "job_slot_browse").Return(sellerUserIDs, nil)
	mockRepo.On("GetBrowsableListings", mock.Anything, userID, sellerUserIDs).Return(expectedListings, nil)

	req := httptest.NewRequest("GET", "/v1/job-slots/listings/browse", nil)

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewJobSlotRentals(mockRouter, mockRepo, mockPermRepo)
	result, httpErr := controller.BrowseListings(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	listings := result.([]*models.JobSlotRentalListing)
	assert.Len(t, listings, 1)

	mockRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
}

func Test_JobSlotRentalsController_CreateInterest_Success(t *testing.T) {
	mockRepo := new(MockJobSlotRentalsRepository)
	mockRouter := &MockRouter{}

	userID := int64(123)
	listingID := int64(1)

	// Listing owned by a different user
	listing := &models.JobSlotRentalListing{
		ID:            listingID,
		UserID:        999,
		CharacterID:   9990,
		CharacterName: "Seller Character",
		ActivityType:  "manufacturing",
		SlotsListed:   3,
		PriceAmount:   100000,
		PricingUnit:   "per_slot_day",
		IsActive:      true,
	}

	mockRepo.On("GetByID", mock.Anything, listingID).Return(listing, nil)
	mockRepo.On("CreateInterest", mock.Anything, mock.MatchedBy(func(interest *models.JobSlotInterestRequest) bool {
		return interest.ListingID == listingID &&
			interest.RequesterUserID == userID &&
			interest.SlotsRequested == 2 &&
			interest.Status == "pending"
	})).Return(nil)

	body := map[string]interface{}{
		"listingId":      listingID,
		"slotsRequested": 2,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/job-slots/interest", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewJobSlotRentals(mockRouter, mockRepo, &MockContactPermissionsRepository{})
	result, httpErr := controller.CreateInterest(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	mockRepo.AssertExpectations(t)
}

func Test_JobSlotRentalsController_CreateInterest_SelfInterest(t *testing.T) {
	mockRepo := new(MockJobSlotRentalsRepository)
	mockRouter := &MockRouter{}

	userID := int64(123)
	listingID := int64(1)

	// Listing owned by the same user
	listing := &models.JobSlotRentalListing{
		ID:            listingID,
		UserID:        userID,
		CharacterID:   456,
		CharacterName: "My Character",
		ActivityType:  "manufacturing",
		SlotsListed:   3,
		PriceAmount:   100000,
		PricingUnit:   "per_slot_day",
		IsActive:      true,
	}

	mockRepo.On("GetByID", mock.Anything, listingID).Return(listing, nil)

	body := map[string]interface{}{
		"listingId":      listingID,
		"slotsRequested": 2,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/job-slots/interest", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewJobSlotRentals(mockRouter, mockRepo, &MockContactPermissionsRepository{})
	result, httpErr := controller.CreateInterest(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "cannot express interest in your own listing")

	mockRepo.AssertExpectations(t)
}

func Test_JobSlotRentalsController_GetSentInterests_Success(t *testing.T) {
	mockRepo := new(MockJobSlotRentalsRepository)
	mockRouter := &MockRouter{}

	userID := int64(123)

	expectedInterests := []*models.JobSlotInterestRequest{
		{
			ID:                   1,
			ListingID:            10,
			RequesterUserID:      userID,
			RequesterName:        "Test User",
			SlotsRequested:       2,
			Status:               "pending",
			ListingActivityType:  "manufacturing",
			ListingCharacterName: "Seller Character",
			ListingOwnerName:     "Seller User",
		},
	}

	mockRepo.On("GetInterestsByRequester", mock.Anything, userID).Return(expectedInterests, nil)

	req := httptest.NewRequest("GET", "/v1/job-slots/interest/sent", nil)

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewJobSlotRentals(mockRouter, mockRepo, &MockContactPermissionsRepository{})
	result, httpErr := controller.GetSentInterests(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	interests := result.([]*models.JobSlotInterestRequest)
	assert.Len(t, interests, 1)

	mockRepo.AssertExpectations(t)
}

func Test_JobSlotRentalsController_GetReceivedInterests_Success(t *testing.T) {
	mockRepo := new(MockJobSlotRentalsRepository)
	mockRouter := &MockRouter{}

	userID := int64(123)

	expectedInterests := []*models.JobSlotInterestRequest{
		{
			ID:                   1,
			ListingID:            10,
			RequesterUserID:      999,
			RequesterName:        "Buyer User",
			SlotsRequested:       2,
			Status:               "pending",
			ListingActivityType:  "manufacturing",
			ListingCharacterName: "My Character",
		},
	}

	mockRepo.On("GetReceivedInterests", mock.Anything, userID).Return(expectedInterests, nil)

	req := httptest.NewRequest("GET", "/v1/job-slots/interest/received", nil)

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewJobSlotRentals(mockRouter, mockRepo, &MockContactPermissionsRepository{})
	result, httpErr := controller.GetReceivedInterests(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	interests := result.([]*models.JobSlotInterestRequest)
	assert.Len(t, interests, 1)

	mockRepo.AssertExpectations(t)
}

func Test_JobSlotRentalsController_UpdateInterestStatus_Success(t *testing.T) {
	mockRepo := new(MockJobSlotRentalsRepository)
	mockRouter := &MockRouter{}

	userID := int64(123)
	interestID := int64(1)

	mockRepo.On("UpdateInterestStatus", mock.Anything, interestID, userID, "accepted").Return(nil)

	body := map[string]interface{}{
		"status": "accepted",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/v1/job-slots/interest/1/status", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	}

	controller := controllers.NewJobSlotRentals(mockRouter, mockRepo, &MockContactPermissionsRepository{})
	result, httpErr := controller.UpdateInterestStatus(args)

	assert.Nil(t, httpErr)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

func Test_JobSlotRentalsController_UpdateInterestStatus_InvalidStatus(t *testing.T) {
	mockRepo := new(MockJobSlotRentalsRepository)
	mockRouter := &MockRouter{}

	userID := int64(123)

	body := map[string]interface{}{
		"status": "invalid_status",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/v1/job-slots/interest/1/status", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	}

	controller := controllers.NewJobSlotRentals(mockRouter, mockRepo, &MockContactPermissionsRepository{})
	result, httpErr := controller.UpdateInterestStatus(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "invalid status")
}

func Test_JobSlotRentalsController_UpdateInterestStatus_NotFound(t *testing.T) {
	mockRepo := new(MockJobSlotRentalsRepository)
	mockRouter := &MockRouter{}

	userID := int64(123)
	interestID := int64(999)

	mockRepo.On("UpdateInterestStatus", mock.Anything, interestID, userID, "accepted").Return(errors.New("interest request not found or user not authorized"))

	body := map[string]interface{}{
		"status": "accepted",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/v1/job-slots/interest/999/status", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "999"},
	}

	controller := controllers.NewJobSlotRentals(mockRouter, mockRepo, &MockContactPermissionsRepository{})
	result, httpErr := controller.UpdateInterestStatus(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}
