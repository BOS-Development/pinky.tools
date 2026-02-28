package controllers

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type JobSlotRentalsRepository interface {
	CalculateSlotInventory(ctx context.Context, userID int64) ([]*models.CharacterSlotInventory, error)
	GetByUser(ctx context.Context, userID int64) ([]*models.JobSlotRentalListing, error)
	GetBrowsableListings(ctx context.Context, viewerUserID int64, sellerUserIDs []int64) ([]*models.JobSlotRentalListing, error)
	Create(ctx context.Context, listing *models.JobSlotRentalListing) error
	Update(ctx context.Context, listing *models.JobSlotRentalListing) error
	Delete(ctx context.Context, listingID int64, userID int64) error
	GetByID(ctx context.Context, listingID int64) (*models.JobSlotRentalListing, error)
	CreateInterest(ctx context.Context, interest *models.JobSlotInterestRequest) error
	GetInterestsByListing(ctx context.Context, listingID int64, sellerUserID int64) ([]*models.JobSlotInterestRequest, error)
	GetInterestsByRequester(ctx context.Context, requesterUserID int64) ([]*models.JobSlotInterestRequest, error)
	UpdateInterestStatus(ctx context.Context, interestID int64, userID int64, status string) error
	GetReceivedInterests(ctx context.Context, userID int64) ([]*models.JobSlotInterestRequest, error)
}

type JobSlotRentals struct {
	repository            JobSlotRentalsRepository
	permissionsRepository ContactPermissionsRepository
}

func NewJobSlotRentals(router Routerer, repository JobSlotRentalsRepository, permissionsRepository ContactPermissionsRepository) *JobSlotRentals {
	controller := &JobSlotRentals{
		repository:            repository,
		permissionsRepository: permissionsRepository,
	}

	router.RegisterRestAPIRoute("/v1/job-slots/inventory", web.AuthAccessUser, controller.GetSlotInventory, "GET")
	router.RegisterRestAPIRoute("/v1/job-slots/listings", web.AuthAccessUser, controller.GetMyListings, "GET")
	router.RegisterRestAPIRoute("/v1/job-slots/listings", web.AuthAccessUser, controller.CreateListing, "POST")
	router.RegisterRestAPIRoute("/v1/job-slots/listings/{id}", web.AuthAccessUser, controller.UpdateListing, "PUT")
	router.RegisterRestAPIRoute("/v1/job-slots/listings/{id}", web.AuthAccessUser, controller.DeleteListing, "DELETE")
	router.RegisterRestAPIRoute("/v1/job-slots/listings/browse", web.AuthAccessUser, controller.BrowseListings, "GET")
	router.RegisterRestAPIRoute("/v1/job-slots/interest", web.AuthAccessUser, controller.CreateInterest, "POST")
	router.RegisterRestAPIRoute("/v1/job-slots/interest/sent", web.AuthAccessUser, controller.GetSentInterests, "GET")
	router.RegisterRestAPIRoute("/v1/job-slots/interest/received", web.AuthAccessUser, controller.GetReceivedInterests, "GET")
	router.RegisterRestAPIRoute("/v1/job-slots/interest/{id}/status", web.AuthAccessUser, controller.UpdateInterestStatus, "PUT")

	return controller
}

// GetSlotInventory returns slot inventory for all characters
func (c *JobSlotRentals) GetSlotInventory(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID := *args.User

	inventory, err := c.repository.CalculateSlotInventory(args.Request.Context(), userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to calculate slot inventory")}
	}

	return inventory, nil
}

// GetMyListings returns all active listings owned by the authenticated user
func (c *JobSlotRentals) GetMyListings(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID := *args.User

	listings, err := c.repository.GetByUser(args.Request.Context(), userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get job slot listings")}
	}

	return listings, nil
}

// CreateListing creates a new job slot rental listing
func (c *JobSlotRentals) CreateListing(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID := *args.User

	var req models.CreateJobSlotListingRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	// Validate required fields
	if req.CharacterID == 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("characterId is required")}
	}
	if req.ActivityType == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("activityType is required")}
	}
	if req.SlotsListed <= 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("slotsListed must be greater than 0")}
	}
	if req.PriceAmount < 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("priceAmount must be non-negative")}
	}
	if req.PricingUnit == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("pricingUnit is required")}
	}

	// Validate activity type
	validActivities := []string{"manufacturing", "reaction", "te_research", "me_research", "copying", "invention"}
	validActivity := false
	for _, activity := range validActivities {
		if req.ActivityType == activity {
			validActivity = true
			break
		}
	}
	if !validActivity {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid activityType")}
	}

	// Check slot availability
	inventory, err := c.repository.CalculateSlotInventory(args.Request.Context(), userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to check slot availability")}
	}

	// Find the character and verify slot availability
	var charInventory *models.CharacterSlotInventory
	for _, inv := range inventory {
		if inv.CharacterID == req.CharacterID {
			charInventory = inv
			break
		}
	}

	if charInventory == nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("character not found or does not belong to user")}
	}

	activityInfo, ok := charInventory.SlotsByActivity[req.ActivityType]
	if !ok {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("activity type not available for character")}
	}

	if activityInfo.SlotsAvailable < req.SlotsListed {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("not enough slots available")}
	}

	listing := &models.JobSlotRentalListing{
		UserID:       userID,
		CharacterID:  req.CharacterID,
		ActivityType: req.ActivityType,
		SlotsListed:  req.SlotsListed,
		PriceAmount:  req.PriceAmount,
		PricingUnit:  req.PricingUnit,
		LocationID:   req.LocationID,
		Notes:        req.Notes,
		IsActive:     true,
	}

	if err := c.repository.Create(args.Request.Context(), listing); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to create listing")}
	}

	return listing, nil
}

// UpdateListing updates an existing listing
func (c *JobSlotRentals) UpdateListing(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID := *args.User

	listingIDStr, ok := args.Params["id"]
	if !ok {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("listing ID is required")}
	}

	listingID, err := strconv.ParseInt(listingIDStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid listing ID")}
	}

	// Get existing listing to verify ownership
	existingListing, err := c.repository.GetByID(args.Request.Context(), listingID)
	if err != nil {
		if errors.Cause(err).Error() == "job slot listing not found" {
			return nil, &web.HttpError{StatusCode: 404, Error: errors.New("job slot listing not found")}
		}
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get job slot listing")}
	}

	if existingListing.UserID != userID {
		return nil, &web.HttpError{StatusCode: 403, Error: errors.New("not authorized to update this listing")}
	}

	var req models.UpdateJobSlotListingRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	// Validate updates
	if req.SlotsListed <= 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("slotsListed must be greater than 0")}
	}
	if req.PriceAmount < 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("priceAmount must be non-negative")}
	}
	if req.PricingUnit == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("pricingUnit is required")}
	}

	// Update fields
	existingListing.SlotsListed = req.SlotsListed
	existingListing.PriceAmount = req.PriceAmount
	existingListing.PricingUnit = req.PricingUnit
	existingListing.LocationID = req.LocationID
	existingListing.Notes = req.Notes
	if req.IsActive != nil {
		existingListing.IsActive = *req.IsActive
	}

	if err := c.repository.Update(args.Request.Context(), existingListing); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to update listing")}
	}

	return existingListing, nil
}

// DeleteListing soft-deletes a listing
func (c *JobSlotRentals) DeleteListing(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID := *args.User

	listingIDStr, ok := args.Params["id"]
	if !ok {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("listing ID is required")}
	}

	listingID, err := strconv.ParseInt(listingIDStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid listing ID")}
	}

	if err := c.repository.Delete(args.Request.Context(), listingID, userID); err != nil {
		if err.Error() == "job slot listing not found or user is not the owner" {
			return nil, &web.HttpError{StatusCode: 404, Error: err}
		}
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to delete listing")}
	}

	return nil, nil
}

// BrowseListings returns listings from contacts who granted browse permission
func (c *JobSlotRentals) BrowseListings(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID := *args.User

	// Get list of users who granted this user permission to browse their job slot listings
	sellerUserIDs, err := c.permissionsRepository.GetUserPermissionsForService(args.Request.Context(), userID, "job_slot_browse")
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get permissions")}
	}

	// Get browsable listings from those sellers
	listings, err := c.repository.GetBrowsableListings(args.Request.Context(), userID, sellerUserIDs)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get browsable listings")}
	}

	return listings, nil
}

// CreateInterest creates a new interest request
func (c *JobSlotRentals) CreateInterest(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID := *args.User

	var req models.CreateInterestRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	// Validate required fields
	if req.ListingID == 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("listingId is required")}
	}
	if req.SlotsRequested <= 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("slotsRequested must be greater than 0")}
	}

	// Get listing to verify it exists and isn't self-interest
	listing, err := c.repository.GetByID(args.Request.Context(), req.ListingID)
	if err != nil {
		if errors.Cause(err).Error() == "job slot listing not found" {
			return nil, &web.HttpError{StatusCode: 404, Error: errors.New("job slot listing not found")}
		}
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get listing")}
	}

	// Prevent self-interest
	if listing.UserID == userID {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("cannot express interest in your own listing")}
	}

	interest := &models.JobSlotInterestRequest{
		ListingID:       req.ListingID,
		RequesterUserID: userID,
		SlotsRequested:  req.SlotsRequested,
		DurationDays:    req.DurationDays,
		Message:         req.Message,
		Status:          "pending",
	}

	if err := c.repository.CreateInterest(args.Request.Context(), interest); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to create interest request")}
	}

	return interest, nil
}

// GetSentInterests returns all interest requests sent by the user
func (c *JobSlotRentals) GetSentInterests(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID := *args.User

	interests, err := c.repository.GetInterestsByRequester(args.Request.Context(), userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get sent interests")}
	}

	return interests, nil
}

// GetReceivedInterests returns all interest requests received for user's listings
func (c *JobSlotRentals) GetReceivedInterests(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID := *args.User

	interests, err := c.repository.GetReceivedInterests(args.Request.Context(), userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get received interests")}
	}

	return interests, nil
}

// UpdateInterestStatus updates the status of an interest request
func (c *JobSlotRentals) UpdateInterestStatus(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID := *args.User

	interestIDStr, ok := args.Params["id"]
	if !ok {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("interest ID is required")}
	}

	interestID, err := strconv.ParseInt(interestIDStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid interest ID")}
	}

	var req models.UpdateInterestStatusRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	if req.Status == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("status is required")}
	}

	// Validate status
	validStatuses := []string{"pending", "accepted", "declined", "withdrawn"}
	validStatus := false
	for _, status := range validStatuses {
		if req.Status == status {
			validStatus = true
			break
		}
	}
	if !validStatus {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid status")}
	}

	if err := c.repository.UpdateInterestStatus(args.Request.Context(), interestID, userID, req.Status); err != nil {
		if err.Error() == "interest request not found or user not authorized" {
			return nil, &web.HttpError{StatusCode: 404, Error: err}
		}
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to update interest status")}
	}

	return nil, nil
}
