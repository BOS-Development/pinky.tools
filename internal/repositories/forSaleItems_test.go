package repositories_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func setupForSaleTestData(t *testing.T, db *sql.DB, userID, charID, typeID, systemID int64) {
	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)

	user := &repositories.User{ID: userID, Name: "Test User"}
	err := userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: charID, Name: "Test Character", UserID: userID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	itemTypes := []models.EveInventoryType{
		{TypeID: typeID, TypeName: "Tritanium", Volume: 0.01},
	}
	err = itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes)
	assert.NoError(t, err)

	// Create region, constellation, and solar system with proper foreign key relationships
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
		systemID, "Jita", 20000020, 0.9)
	assert.NoError(t, err)
}

func Test_ForSaleItemsShouldCreate(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupForSaleTestData(t, db, 2000, 20000, 34, 30000142)
	forSaleRepo := repositories.NewForSaleItems(db)

	item := &models.ForSaleItem{
		UserID:            2000,
		TypeID:            34,
		OwnerType:         "character",
		OwnerID:           20000,
		LocationID:        30000142,
		QuantityAvailable: 1000,
		PricePerUnit:      50,
		IsActive:          true,
	}

	err = forSaleRepo.Upsert(context.Background(), item)
	assert.NoError(t, err)
	assert.NotZero(t, item.ID)
}

func Test_ForSaleItemsShouldGetByUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupForSaleTestData(t, db, 2100, 21000, 35, 30000143)
	forSaleRepo := repositories.NewForSaleItems(db)

	item := &models.ForSaleItem{
		UserID:            2100,
		TypeID:            35,
		OwnerType:         "character",
		OwnerID:           21000,
		LocationID:        30000143,
		QuantityAvailable: 2000,
		PricePerUnit:      100,
		IsActive:          true,
	}

	err = forSaleRepo.Upsert(context.Background(), item)
	assert.NoError(t, err)

	items, err := forSaleRepo.GetByUser(context.Background(), 2100)
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "Tritanium", items[0].TypeName)
	assert.Equal(t, "Test Character", items[0].OwnerName)
	assert.Equal(t, "Jita", items[0].LocationName)
	assert.Equal(t, int64(2000), items[0].QuantityAvailable)
	assert.Equal(t, int64(100), items[0].PricePerUnit)
}

func Test_ForSaleItemsShouldDelete(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupForSaleTestData(t, db, 2300, 23000, 37, 30000145)
	forSaleRepo := repositories.NewForSaleItems(db)

	item := &models.ForSaleItem{
		UserID:            2300,
		TypeID:            37,
		OwnerType:         "character",
		OwnerID:           23000,
		LocationID:        30000145,
		QuantityAvailable: 300,
		PricePerUnit:      150,
		IsActive:          true,
	}

	err = forSaleRepo.Upsert(context.Background(), item)
	assert.NoError(t, err)

	err = forSaleRepo.Delete(context.Background(), item.ID, 2300)
	assert.NoError(t, err)

	items, err := forSaleRepo.GetByUser(context.Background(), 2300)
	assert.NoError(t, err)
	assert.Len(t, items, 0)
}

func Test_ForSaleItemsShouldGetUserIDByCharacterID(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	forSaleRepo := repositories.NewForSaleItems(db)

	user := &repositories.User{ID: 2900, Name: "Test User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 29000, Name: "Test Character", UserID: 2900}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	userID, err := forSaleRepo.GetUserIDByCharacterID(context.Background(), 29000)
	assert.NoError(t, err)
	assert.Equal(t, int64(2900), userID)
}

func Test_ForSaleItemsUpdateQuantityToZero_ShouldMarkInactive(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupForSaleTestData(t, db, 2400, 24000, 38, 30000146)
	forSaleRepo := repositories.NewForSaleItems(db)

	item := &models.ForSaleItem{
		UserID:            2400,
		TypeID:            38,
		OwnerType:         "character",
		OwnerID:           24000,
		LocationID:        30000146,
		QuantityAvailable: 100,
		PricePerUnit:      75,
		IsActive:          true,
	}

	err = forSaleRepo.Upsert(context.Background(), item)
	assert.NoError(t, err)

	// Begin transaction and update quantity to 0
	tx, err := db.BeginTx(context.Background(), nil)
	assert.NoError(t, err)
	defer tx.Rollback()

	err = forSaleRepo.UpdateQuantity(context.Background(), tx, item.ID, 0)
	assert.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	// Verify item is inactive but quantity is preserved
	updated, err := forSaleRepo.GetByID(context.Background(), item.ID)
	assert.NoError(t, err)
	assert.False(t, updated.IsActive)
	assert.Equal(t, int64(100), updated.QuantityAvailable) // Quantity preserved
}

func Test_ForSaleItemsUpdateQuantityPartial_ShouldUpdateQuantity(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupForSaleTestData(t, db, 2500, 25000, 39, 30000147)
	forSaleRepo := repositories.NewForSaleItems(db)

	item := &models.ForSaleItem{
		UserID:            2500,
		TypeID:            39,
		OwnerType:         "character",
		OwnerID:           25000,
		LocationID:        30000147,
		QuantityAvailable: 500,
		PricePerUnit:      85,
		IsActive:          true,
	}

	err = forSaleRepo.Upsert(context.Background(), item)
	assert.NoError(t, err)

	// Begin transaction and update quantity to partial
	tx, err := db.BeginTx(context.Background(), nil)
	assert.NoError(t, err)
	defer tx.Rollback()

	err = forSaleRepo.UpdateQuantity(context.Background(), tx, item.ID, 250)
	assert.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	// Verify item is still active with updated quantity
	updated, err := forSaleRepo.GetByID(context.Background(), item.ID)
	assert.NoError(t, err)
	assert.True(t, updated.IsActive)
	assert.Equal(t, int64(250), updated.QuantityAvailable)
}
