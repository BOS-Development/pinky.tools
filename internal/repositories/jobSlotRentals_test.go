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
