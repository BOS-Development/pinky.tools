package repositories_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func setupJobSlotRentalTestData(t *testing.T, db *sql.DB, userID, charID int64) {
	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)

	user := &repositories.User{ID: userID, Name: "Test User"}
	err := userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: charID, Name: "Test Character", UserID: userID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	// Add some test skills for slot calculation
	_, err = db.ExecContext(context.Background(),
		`INSERT INTO character_skills (character_id, user_id, skill_id, trained_level, active_level, skillpoints, updated_at)
		VALUES ($1, $2, 3387, 5, 5, 100000, NOW()),
		       ($1, $2, 24625, 5, 5, 100000, NOW()),
		       ($1, $2, 45746, 4, 4, 80000, NOW()),
		       ($1, $2, 45748, 3, 3, 60000, NOW()),
		       ($1, $2, 3402, 3, 3, 60000, NOW()),
		       ($1, $2, 3406, 2, 2, 40000, NOW())`,
		charID, userID)
	assert.NoError(t, err)

	// Create region, constellation, and solar system
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
}

func Test_JobSlotRentalsCreateListing(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userID := int64(8000)
	charID := int64(80000)
	setupJobSlotRentalTestData(t, db, userID, charID)

	repo := repositories.NewJobSlotRentals(db)

	locationID := int64(30000142)
	listing := &models.JobSlotRentalListing{
		UserID:       userID,
		CharacterID:  charID,
		ActivityType: "manufacturing",
		SlotsListed:  2,
		PriceAmount:  100000,
		PricingUnit:  "per_slot_day",
		LocationID:   &locationID,
		IsActive:     true,
	}

	err = repo.Create(context.Background(), listing)
	assert.NoError(t, err)
	assert.NotZero(t, listing.ID)
	assert.NotZero(t, listing.CreatedAt)
}

func Test_JobSlotRentalsGetByUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userID := int64(8100)
	charID := int64(81000)
	setupJobSlotRentalTestData(t, db, userID, charID)

	repo := repositories.NewJobSlotRentals(db)

	locationID := int64(30000142)
	listing := &models.JobSlotRentalListing{
		UserID:       userID,
		CharacterID:  charID,
		ActivityType: "reaction",
		SlotsListed:  1,
		PriceAmount:  50000,
		PricingUnit:  "per_job",
		LocationID:   &locationID,
		IsActive:     true,
	}

	err = repo.Create(context.Background(), listing)
	assert.NoError(t, err)

	listings, err := repo.GetByUser(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, listings, 1)
	assert.Equal(t, "reaction", listings[0].ActivityType)
	assert.Equal(t, 1, listings[0].SlotsListed)
	assert.Equal(t, "Test Character", listings[0].CharacterName)
}

func Test_JobSlotRentalsUpdateListing(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userID := int64(8200)
	charID := int64(82000)
	setupJobSlotRentalTestData(t, db, userID, charID)

	repo := repositories.NewJobSlotRentals(db)

	locationID := int64(30000142)
	listing := &models.JobSlotRentalListing{
		UserID:       userID,
		CharacterID:  charID,
		ActivityType: "invention",
		SlotsListed:  2,
		PriceAmount:  75000,
		PricingUnit:  "per_slot_day",
		LocationID:   &locationID,
		IsActive:     true,
	}

	err = repo.Create(context.Background(), listing)
	assert.NoError(t, err)

	// Update
	listing.SlotsListed = 3
	listing.PriceAmount = 90000
	err = repo.Update(context.Background(), listing)
	assert.NoError(t, err)

	// Verify
	updated, err := repo.GetByID(context.Background(), listing.ID)
	assert.NoError(t, err)
	assert.Equal(t, 3, updated.SlotsListed)
	assert.Equal(t, float64(90000), updated.PriceAmount)
}

func Test_JobSlotRentalsDeleteListing(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userID := int64(8300)
	charID := int64(83000)
	setupJobSlotRentalTestData(t, db, userID, charID)

	repo := repositories.NewJobSlotRentals(db)

	locationID := int64(30000142)
	listing := &models.JobSlotRentalListing{
		UserID:       userID,
		CharacterID:  charID,
		ActivityType: "copying",
		SlotsListed:  1,
		PriceAmount:  25000,
		PricingUnit:  "per_slot_day",
		LocationID:   &locationID,
		IsActive:     true,
	}

	err = repo.Create(context.Background(), listing)
	assert.NoError(t, err)

	// Delete
	err = repo.Delete(context.Background(), listing.ID, userID)
	assert.NoError(t, err)

	// Verify no longer active
	listings, err := repo.GetByUser(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, listings, 0)
}

func Test_JobSlotRentalsCalculateSlotInventory(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userID := int64(8400)
	charID := int64(84000)
	setupJobSlotRentalTestData(t, db, userID, charID)

	repo := repositories.NewJobSlotRentals(db)

	// Calculate inventory
	inventory, err := repo.CalculateSlotInventory(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, inventory, 1)
	assert.Equal(t, charID, inventory[0].CharacterID)
	assert.Equal(t, "Test Character", inventory[0].CharacterName)

	// Verify manufacturing slots (1 + MassProd(5) + AdvMassProd(5) = 11)
	mfg := inventory[0].SlotsByActivity["manufacturing"]
	assert.NotNil(t, mfg)
	assert.Equal(t, 11, mfg.SlotsMax)
	assert.Equal(t, 0, mfg.SlotsInUse)
	assert.Equal(t, 11, mfg.SlotsAvailable)

	// Verify reaction slots (1 + MassReactions(3) = 4)
	react := inventory[0].SlotsByActivity["reaction"]
	assert.NotNil(t, react)
	assert.Equal(t, 4, react.SlotsMax)

	// Verify science slots (1 + LabOp(2) = 3)
	invention := inventory[0].SlotsByActivity["invention"]
	assert.NotNil(t, invention)
	assert.Equal(t, 3, invention.SlotsMax)
}

func Test_JobSlotRentalsCalculateSlotInventoryWithListings(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userID := int64(8500)
	charID := int64(85000)
	setupJobSlotRentalTestData(t, db, userID, charID)

	repo := repositories.NewJobSlotRentals(db)

	// Create a listing for 2 manufacturing slots
	locationID := int64(30000142)
	listing := &models.JobSlotRentalListing{
		UserID:       userID,
		CharacterID:  charID,
		ActivityType: "manufacturing",
		SlotsListed:  2,
		PriceAmount:  100000,
		PricingUnit:  "per_slot_day",
		LocationID:   &locationID,
		IsActive:     true,
	}

	err = repo.Create(context.Background(), listing)
	assert.NoError(t, err)

	// Calculate inventory
	inventory, err := repo.CalculateSlotInventory(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, inventory, 1)

	// Verify manufacturing slots show 2 listed
	mfg := inventory[0].SlotsByActivity["manufacturing"]
	assert.Equal(t, 2, mfg.SlotsListed)
	assert.Equal(t, 9, mfg.SlotsAvailable) // 11 max - 2 listed = 9 available
}

func Test_JobSlotRentalsCreateInterest(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	sellerUserID := int64(8600)
	sellerCharID := int64(86000)
	buyerUserID := int64(8601)
	setupJobSlotRentalTestData(t, db, sellerUserID, sellerCharID)

	// Create buyer user
	userRepo := repositories.NewUserRepository(db)
	buyer := &repositories.User{ID: buyerUserID, Name: "Buyer User"}
	err = userRepo.Add(context.Background(), buyer)
	assert.NoError(t, err)

	repo := repositories.NewJobSlotRentals(db)

	// Create listing
	locationID := int64(30000142)
	listing := &models.JobSlotRentalListing{
		UserID:       sellerUserID,
		CharacterID:  sellerCharID,
		ActivityType: "manufacturing",
		SlotsListed:  3,
		PriceAmount:  100000,
		PricingUnit:  "per_slot_day",
		LocationID:   &locationID,
		IsActive:     true,
	}

	err = repo.Create(context.Background(), listing)
	assert.NoError(t, err)

	// Create interest
	durationDays := 7
	interest := &models.JobSlotInterestRequest{
		ListingID:       listing.ID,
		RequesterUserID: buyerUserID,
		SlotsRequested:  2,
		DurationDays:    &durationDays,
		Status:          "pending",
	}

	err = repo.CreateInterest(context.Background(), interest)
	assert.NoError(t, err)
	assert.NotZero(t, interest.ID)
}

func Test_JobSlotRentalsGetInterestsByListing(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	sellerUserID := int64(8700)
	sellerCharID := int64(87000)
	buyerUserID := int64(8701)
	setupJobSlotRentalTestData(t, db, sellerUserID, sellerCharID)

	// Create buyer user
	userRepo := repositories.NewUserRepository(db)
	buyer := &repositories.User{ID: buyerUserID, Name: "Buyer User"}
	err = userRepo.Add(context.Background(), buyer)
	assert.NoError(t, err)

	repo := repositories.NewJobSlotRentals(db)

	// Create listing
	locationID := int64(30000142)
	listing := &models.JobSlotRentalListing{
		UserID:       sellerUserID,
		CharacterID:  sellerCharID,
		ActivityType: "reaction",
		SlotsListed:  2,
		PriceAmount:  50000,
		PricingUnit:  "per_job",
		LocationID:   &locationID,
		IsActive:     true,
	}

	err = repo.Create(context.Background(), listing)
	assert.NoError(t, err)

	// Create interest
	interest := &models.JobSlotInterestRequest{
		ListingID:       listing.ID,
		RequesterUserID: buyerUserID,
		SlotsRequested:  1,
		Status:          "pending",
	}

	err = repo.CreateInterest(context.Background(), interest)
	assert.NoError(t, err)

	// Get interests
	interests, err := repo.GetInterestsByListing(context.Background(), listing.ID, sellerUserID)
	assert.NoError(t, err)
	assert.Len(t, interests, 1)
	assert.Equal(t, "Buyer User", interests[0].RequesterName)
	assert.Equal(t, 1, interests[0].SlotsRequested)
}

func Test_JobSlotRentalsGetInterestsByRequester(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	sellerUserID := int64(8800)
	sellerCharID := int64(88000)
	buyerUserID := int64(8801)
	setupJobSlotRentalTestData(t, db, sellerUserID, sellerCharID)

	// Create buyer user
	userRepo := repositories.NewUserRepository(db)
	buyer := &repositories.User{ID: buyerUserID, Name: "Buyer User"}
	err = userRepo.Add(context.Background(), buyer)
	assert.NoError(t, err)

	repo := repositories.NewJobSlotRentals(db)

	// Create listing
	locationID := int64(30000142)
	listing := &models.JobSlotRentalListing{
		UserID:       sellerUserID,
		CharacterID:  sellerCharID,
		ActivityType: "invention",
		SlotsListed:  2,
		PriceAmount:  75000,
		PricingUnit:  "per_slot_day",
		LocationID:   &locationID,
		IsActive:     true,
	}

	err = repo.Create(context.Background(), listing)
	assert.NoError(t, err)

	// Create interest
	interest := &models.JobSlotInterestRequest{
		ListingID:       listing.ID,
		RequesterUserID: buyerUserID,
		SlotsRequested:  2,
		Status:          "pending",
	}

	err = repo.CreateInterest(context.Background(), interest)
	assert.NoError(t, err)

	// Get interests by requester
	interests, err := repo.GetInterestsByRequester(context.Background(), buyerUserID)
	assert.NoError(t, err)
	assert.Len(t, interests, 1)
	assert.Equal(t, "invention", interests[0].ListingActivityType)
	assert.Equal(t, "Test Character", interests[0].ListingCharacterName)
	assert.Equal(t, "Test User", interests[0].ListingOwnerName)
}

func Test_JobSlotRentalsGetInterestByID(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	sellerUserID := int64(9000)
	sellerCharID := int64(90000)
	buyerUserID := int64(9001)
	setupJobSlotRentalTestData(t, db, sellerUserID, sellerCharID)

	// Create buyer user
	userRepo := repositories.NewUserRepository(db)
	buyer := &repositories.User{ID: buyerUserID, Name: "Buyer User"}
	err = userRepo.Add(context.Background(), buyer)
	assert.NoError(t, err)

	repo := repositories.NewJobSlotRentals(db)

	// Create listing
	locationID := int64(30000142)
	listing := &models.JobSlotRentalListing{
		UserID:       sellerUserID,
		CharacterID:  sellerCharID,
		ActivityType: "manufacturing",
		SlotsListed:  3,
		PriceAmount:  100000,
		PricingUnit:  "per_slot_day",
		LocationID:   &locationID,
		IsActive:     true,
	}

	err = repo.Create(context.Background(), listing)
	assert.NoError(t, err)

	// Create interest
	durationDays := 14
	msg := "Looking for manufacturing slots"
	interest := &models.JobSlotInterestRequest{
		ListingID:       listing.ID,
		RequesterUserID: buyerUserID,
		SlotsRequested:  2,
		DurationDays:    &durationDays,
		Message:         &msg,
		Status:          "pending",
	}

	err = repo.CreateInterest(context.Background(), interest)
	assert.NoError(t, err)
	assert.NotZero(t, interest.ID)

	// GetInterestByID should return enriched interest
	enriched, err := repo.GetInterestByID(context.Background(), interest.ID)
	assert.NoError(t, err)
	assert.NotNil(t, enriched)

	assert.Equal(t, interest.ID, enriched.ID)
	assert.Equal(t, listing.ID, enriched.ListingID)
	assert.Equal(t, buyerUserID, enriched.RequesterUserID)
	assert.Equal(t, "Buyer User", enriched.RequesterName)
	assert.Equal(t, 2, enriched.SlotsRequested)
	assert.Equal(t, "pending", enriched.Status)

	// Enriched listing fields
	assert.Equal(t, "manufacturing", enriched.ListingActivityType)
	assert.Equal(t, "Test Character", enriched.ListingCharacterName)
	assert.Equal(t, "Test User", enriched.ListingOwnerName)
	assert.Equal(t, float64(100000), enriched.ListingPriceAmount)
	assert.Equal(t, "per_slot_day", enriched.ListingPricingUnit)
}

func Test_JobSlotRentalsGetInterestByID_NotFound(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewJobSlotRentals(db)

	_, err = repo.GetInterestByID(context.Background(), 999999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interest request not found")
}

func Test_JobSlotRentalsUpdateInterestStatus(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	sellerUserID := int64(8900)
	sellerCharID := int64(89000)
	buyerUserID := int64(8901)
	setupJobSlotRentalTestData(t, db, sellerUserID, sellerCharID)

	// Create buyer user
	userRepo := repositories.NewUserRepository(db)
	buyer := &repositories.User{ID: buyerUserID, Name: "Buyer User"}
	err = userRepo.Add(context.Background(), buyer)
	assert.NoError(t, err)

	repo := repositories.NewJobSlotRentals(db)

	// Create listing
	locationID := int64(30000142)
	listing := &models.JobSlotRentalListing{
		UserID:       sellerUserID,
		CharacterID:  sellerCharID,
		ActivityType: "manufacturing",
		SlotsListed:  3,
		PriceAmount:  100000,
		PricingUnit:  "per_slot_day",
		LocationID:   &locationID,
		IsActive:     true,
	}

	err = repo.Create(context.Background(), listing)
	assert.NoError(t, err)

	// Create interest
	interest := &models.JobSlotInterestRequest{
		ListingID:       listing.ID,
		RequesterUserID: buyerUserID,
		SlotsRequested:  2,
		Status:          "pending",
	}

	err = repo.CreateInterest(context.Background(), interest)
	assert.NoError(t, err)

	// Seller accepts
	err = repo.UpdateInterestStatus(context.Background(), interest.ID, sellerUserID, "accepted")
	assert.NoError(t, err)

	// Verify status
	interests, err := repo.GetInterestsByListing(context.Background(), listing.ID, sellerUserID)
	assert.NoError(t, err)
	assert.Len(t, interests, 1)
	assert.Equal(t, "accepted", interests[0].Status)
}

// --- Phase 3: Agreement repository tests ---

func setupJobSlotAgreementTestData(t *testing.T, db *sql.DB, sellerUserID, sellerCharID, renterUserID int64) (*models.JobSlotRentalListing, *models.JobSlotInterestRequest) {
	setupJobSlotRentalTestData(t, db, sellerUserID, sellerCharID)

	userRepo := repositories.NewUserRepository(db)
	renter := &repositories.User{ID: renterUserID, Name: "Renter User"}
	err := userRepo.Add(context.Background(), renter)
	assert.NoError(t, err)

	repo := repositories.NewJobSlotRentals(db)

	locationID := int64(30000142)
	listing := &models.JobSlotRentalListing{
		UserID:       sellerUserID,
		CharacterID:  sellerCharID,
		ActivityType: "manufacturing",
		SlotsListed:  3,
		PriceAmount:  100000,
		PricingUnit:  "per_slot_day",
		LocationID:   &locationID,
		IsActive:     true,
	}

	err = repo.Create(context.Background(), listing)
	assert.NoError(t, err)

	interest := &models.JobSlotInterestRequest{
		ListingID:       listing.ID,
		RequesterUserID: renterUserID,
		SlotsRequested:  2,
		Status:          "pending",
	}

	err = repo.CreateInterest(context.Background(), interest)
	assert.NoError(t, err)

	return listing, interest
}

func Test_JobSlotRentalsAcceptInterestWithAgreement_HappyPath(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	sellerUserID := int64(9100)
	sellerCharID := int64(91000)
	renterUserID := int64(9101)
	listing, interest := setupJobSlotAgreementTestData(t, db, sellerUserID, sellerCharID, renterUserID)

	repo := repositories.NewJobSlotRentals(db)

	agreement, err := repo.AcceptInterestWithAgreement(context.Background(), interest.ID, sellerUserID)
	assert.NoError(t, err)
	assert.NotNil(t, agreement)

	assert.NotZero(t, agreement.ID)
	assert.Equal(t, interest.ID, agreement.InterestRequestID)
	assert.Equal(t, listing.ID, agreement.ListingID)
	assert.Equal(t, sellerUserID, agreement.SellerUserID)
	assert.Equal(t, renterUserID, agreement.RenterUserID)
	assert.Equal(t, 2, agreement.SlotsAgreed)
	assert.Equal(t, float64(100000), agreement.PriceAmount)
	assert.Equal(t, "per_slot_day", agreement.PricingUnit)
	assert.Equal(t, "active", agreement.Status)
	assert.Equal(t, "Test User", agreement.SellerName)
	assert.Equal(t, "Renter User", agreement.RenterName)
	assert.Equal(t, "manufacturing", agreement.ActivityType)
	assert.Equal(t, "Test Character", agreement.CharacterName)
	assert.Equal(t, sellerCharID, agreement.CharacterID)

	// Interest status should now be "accepted"
	interests, err := repo.GetInterestsByListing(context.Background(), listing.ID, sellerUserID)
	assert.NoError(t, err)
	assert.Len(t, interests, 1)
	assert.Equal(t, "accepted", interests[0].Status)
}

func Test_JobSlotRentalsAcceptInterestWithAgreement_NotOwner(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	sellerUserID := int64(9200)
	sellerCharID := int64(92000)
	renterUserID := int64(9201)
	otherUserID := int64(9202)
	_, interest := setupJobSlotAgreementTestData(t, db, sellerUserID, sellerCharID, renterUserID)

	// Create unrelated user
	userRepo := repositories.NewUserRepository(db)
	other := &repositories.User{ID: otherUserID, Name: "Other User"}
	err = userRepo.Add(context.Background(), other)
	assert.NoError(t, err)

	repo := repositories.NewJobSlotRentals(db)

	_, err = repo.AcceptInterestWithAgreement(context.Background(), interest.ID, otherUserID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interest request not found or user not authorized")
}

func Test_JobSlotRentalsAcceptInterestWithAgreement_NotPending(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	sellerUserID := int64(9300)
	sellerCharID := int64(93000)
	renterUserID := int64(9301)
	_, interest := setupJobSlotAgreementTestData(t, db, sellerUserID, sellerCharID, renterUserID)

	// Withdraw the interest first
	repo := repositories.NewJobSlotRentals(db)
	err = repo.UpdateInterestStatus(context.Background(), interest.ID, renterUserID, "withdrawn")
	assert.NoError(t, err)

	// Now try to accept the withdrawn interest
	_, err = repo.AcceptInterestWithAgreement(context.Background(), interest.ID, sellerUserID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interest request is not pending")
}

func Test_JobSlotRentalsGetAgreements_BothRoles(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	sellerUserID := int64(9400)
	sellerCharID := int64(94000)
	renterUserID := int64(9401)
	_, interest := setupJobSlotAgreementTestData(t, db, sellerUserID, sellerCharID, renterUserID)

	repo := repositories.NewJobSlotRentals(db)

	// Create agreement
	_, err = repo.AcceptInterestWithAgreement(context.Background(), interest.ID, sellerUserID)
	assert.NoError(t, err)

	// Both seller and renter should see it when no role filter
	sellerAgreements, err := repo.GetAgreements(context.Background(), sellerUserID, "", "")
	assert.NoError(t, err)
	assert.Len(t, sellerAgreements, 1)

	renterAgreements, err := repo.GetAgreements(context.Background(), renterUserID, "", "")
	assert.NoError(t, err)
	assert.Len(t, renterAgreements, 1)
}

func Test_JobSlotRentalsGetAgreements_RoleFilter(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	sellerUserID := int64(9500)
	sellerCharID := int64(95000)
	renterUserID := int64(9501)
	_, interest := setupJobSlotAgreementTestData(t, db, sellerUserID, sellerCharID, renterUserID)

	repo := repositories.NewJobSlotRentals(db)

	_, err = repo.AcceptInterestWithAgreement(context.Background(), interest.ID, sellerUserID)
	assert.NoError(t, err)

	// Seller role filter
	sellerOnly, err := repo.GetAgreements(context.Background(), sellerUserID, "", "seller")
	assert.NoError(t, err)
	assert.Len(t, sellerOnly, 1)

	// Renter role filter for seller should return nothing
	renterFilter, err := repo.GetAgreements(context.Background(), sellerUserID, "", "renter")
	assert.NoError(t, err)
	assert.Len(t, renterFilter, 0)
}

func Test_JobSlotRentalsGetAgreements_StatusFilter(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	sellerUserID := int64(9600)
	sellerCharID := int64(96000)
	renterUserID := int64(9601)
	_, interest := setupJobSlotAgreementTestData(t, db, sellerUserID, sellerCharID, renterUserID)

	repo := repositories.NewJobSlotRentals(db)

	agreement, err := repo.AcceptInterestWithAgreement(context.Background(), interest.ID, sellerUserID)
	assert.NoError(t, err)

	// Filter by active
	active, err := repo.GetAgreements(context.Background(), sellerUserID, "active", "")
	assert.NoError(t, err)
	assert.Len(t, active, 1)

	// Cancel it
	err = repo.UpdateAgreementStatus(context.Background(), agreement.ID, sellerUserID, "cancelled", nil)
	assert.NoError(t, err)

	// Filter by active should be empty now
	activeAfter, err := repo.GetAgreements(context.Background(), sellerUserID, "active", "")
	assert.NoError(t, err)
	assert.Len(t, activeAfter, 0)

	// Filter by cancelled should have one
	cancelled, err := repo.GetAgreements(context.Background(), sellerUserID, "cancelled", "")
	assert.NoError(t, err)
	assert.Len(t, cancelled, 1)
}

func Test_JobSlotRentalsUpdateAgreementStatus_Complete(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	sellerUserID := int64(9700)
	sellerCharID := int64(97000)
	renterUserID := int64(9701)
	_, interest := setupJobSlotAgreementTestData(t, db, sellerUserID, sellerCharID, renterUserID)

	repo := repositories.NewJobSlotRentals(db)

	agreement, err := repo.AcceptInterestWithAgreement(context.Background(), interest.ID, sellerUserID)
	assert.NoError(t, err)

	// Complete
	err = repo.UpdateAgreementStatus(context.Background(), agreement.ID, sellerUserID, "completed", nil)
	assert.NoError(t, err)

	// Verify
	agreements, err := repo.GetAgreements(context.Background(), sellerUserID, "completed", "")
	assert.NoError(t, err)
	assert.Len(t, agreements, 1)
	assert.Equal(t, "completed", agreements[0].Status)
}

func Test_JobSlotRentalsUpdateAgreementStatus_CancelWithReason(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	sellerUserID := int64(9800)
	sellerCharID := int64(98000)
	renterUserID := int64(9801)
	_, interest := setupJobSlotAgreementTestData(t, db, sellerUserID, sellerCharID, renterUserID)

	repo := repositories.NewJobSlotRentals(db)

	agreement, err := repo.AcceptInterestWithAgreement(context.Background(), interest.ID, sellerUserID)
	assert.NoError(t, err)

	reason := "Renter stopped responding"
	err = repo.UpdateAgreementStatus(context.Background(), agreement.ID, sellerUserID, "cancelled", &reason)
	assert.NoError(t, err)

	// Verify cancellation reason
	agreements, err := repo.GetAgreements(context.Background(), sellerUserID, "cancelled", "")
	assert.NoError(t, err)
	assert.Len(t, agreements, 1)
	assert.Equal(t, "cancelled", agreements[0].Status)
	assert.NotNil(t, agreements[0].CancellationReason)
	assert.Equal(t, reason, *agreements[0].CancellationReason)
}

func Test_JobSlotRentalsUpdateAgreementStatus_NotOwner(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	sellerUserID := int64(9900)
	sellerCharID := int64(99000)
	renterUserID := int64(9901)
	_, interest := setupJobSlotAgreementTestData(t, db, sellerUserID, sellerCharID, renterUserID)

	repo := repositories.NewJobSlotRentals(db)

	agreement, err := repo.AcceptInterestWithAgreement(context.Background(), interest.ID, sellerUserID)
	assert.NoError(t, err)

	// Renter tries to complete — should fail
	err = repo.UpdateAgreementStatus(context.Background(), agreement.ID, renterUserID, "completed", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not authorized to update this agreement")
}

func Test_JobSlotRentalsUpdateAgreementStatus_AlreadyCompleted(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	sellerUserID := int64(9910)
	sellerCharID := int64(99100)
	renterUserID := int64(9911)
	_, interest := setupJobSlotAgreementTestData(t, db, sellerUserID, sellerCharID, renterUserID)

	repo := repositories.NewJobSlotRentals(db)

	agreement, err := repo.AcceptInterestWithAgreement(context.Background(), interest.ID, sellerUserID)
	assert.NoError(t, err)

	err = repo.UpdateAgreementStatus(context.Background(), agreement.ID, sellerUserID, "completed", nil)
	assert.NoError(t, err)

	// Try to cancel an already-completed agreement
	err = repo.UpdateAgreementStatus(context.Background(), agreement.ID, sellerUserID, "cancelled", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only active agreements can be updated")
}

func Test_JobSlotRentalsGetAgreementJobsByID_NotFound(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewJobSlotRentals(db)

	_, err = repo.GetAgreementJobsByID(context.Background(), 999999, 123)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agreement not found")
}

func Test_JobSlotRentalsGetAgreementJobsByID_NotAuthorized(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	sellerUserID := int64(9920)
	sellerCharID := int64(99200)
	renterUserID := int64(9921)
	_, interest := setupJobSlotAgreementTestData(t, db, sellerUserID, sellerCharID, renterUserID)

	repo := repositories.NewJobSlotRentals(db)

	agreement, err := repo.AcceptInterestWithAgreement(context.Background(), interest.ID, sellerUserID)
	assert.NoError(t, err)

	// Renter tries to view jobs — should fail
	_, err = repo.GetAgreementJobsByID(context.Background(), agreement.ID, renterUserID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not authorized to view jobs for this agreement")
}

func Test_JobSlotRentalsGetAgreementJobsByID_EmptyJobs(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	sellerUserID := int64(9930)
	sellerCharID := int64(99300)
	renterUserID := int64(9931)
	_, interest := setupJobSlotAgreementTestData(t, db, sellerUserID, sellerCharID, renterUserID)

	repo := repositories.NewJobSlotRentals(db)

	agreement, err := repo.AcceptInterestWithAgreement(context.Background(), interest.ID, sellerUserID)
	assert.NoError(t, err)

	// No ESI jobs for character — should return empty slice
	jobs, err := repo.GetAgreementJobsByID(context.Background(), agreement.ID, sellerUserID)
	assert.NoError(t, err)
	assert.NotNil(t, jobs)
	assert.Len(t, jobs, 0)
}
