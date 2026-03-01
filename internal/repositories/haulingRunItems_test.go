package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

// createHaulingRunForItems is a helper to create a user and run for item tests.
func createHaulingRunForItems(t *testing.T, userID int64, userName string) (*models.HaulingRun, *repositories.HaulingRunItems, func()) {
	t.Helper()
	db, err := setupDatabase(t)
	if err != nil {
		t.Fatalf("failed to setup db: %v", err)
	}

	userRepo := repositories.NewUserRepository(db)
	user := &repositories.User{ID: userID, Name: userName}
	if err := userRepo.Add(context.Background(), user); err != nil {
		t.Fatalf("failed to add user: %v", err)
	}

	runsRepo := repositories.NewHaulingRuns(db)
	run, err := runsRepo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Item Test Run",
		Status:       "PLANNING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	})
	if err != nil {
		t.Fatalf("failed to create run: %v", err)
	}

	itemsRepo := repositories.NewHaulingRunItems(db)
	return run, itemsRepo, func() {}
}

func Test_HaulingRunItems_AddItem(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9100)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Item User 1"})
	assert.NoError(t, err)

	runsRepo := repositories.NewHaulingRuns(db)
	run, err := runsRepo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Item Add Run",
		Status:       "PLANNING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	})
	assert.NoError(t, err)

	itemsRepo := repositories.NewHaulingRunItems(db)

	buyPrice := 1000.0
	sellPrice := 1500.0
	volume := 0.5

	item := &models.HaulingRunItem{
		RunID:           run.ID,
		TypeID:          int64(34),
		TypeName:        "Tritanium",
		QuantityPlanned: int64(100),
		BuyPriceISK:     &buyPrice,
		SellPriceISK:    &sellPrice,
		VolumeM3:        &volume,
	}

	created, err := itemsRepo.AddItem(context.Background(), item)
	assert.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, run.ID, created.RunID)
	assert.Equal(t, int64(34), created.TypeID)
	assert.Equal(t, "Tritanium", created.TypeName)
	assert.Equal(t, int64(100), created.QuantityPlanned)
	assert.NotEmpty(t, created.CreatedAt)
}

func Test_HaulingRunItems_GetItemsByRunID(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9110)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Item User 2"})
	assert.NoError(t, err)

	runsRepo := repositories.NewHaulingRuns(db)
	run, err := runsRepo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Get Items Run",
		Status:       "PLANNING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	})
	assert.NoError(t, err)

	itemsRepo := repositories.NewHaulingRunItems(db)

	buyPrice := 1000.0
	sellPrice := 1500.0

	// Add two items
	_, err = itemsRepo.AddItem(context.Background(), &models.HaulingRunItem{
		RunID:           run.ID,
		TypeID:          int64(34),
		TypeName:        "Tritanium",
		QuantityPlanned: int64(100),
		BuyPriceISK:     &buyPrice,
		SellPriceISK:    &sellPrice,
	})
	assert.NoError(t, err)

	_, err = itemsRepo.AddItem(context.Background(), &models.HaulingRunItem{
		RunID:           run.ID,
		TypeID:          int64(35),
		TypeName:        "Pyerite",
		QuantityPlanned: int64(50),
	})
	assert.NoError(t, err)

	items, err := itemsRepo.GetItemsByRunID(context.Background(), run.ID)
	assert.NoError(t, err)
	assert.Len(t, items, 2)

	// Verify computed fields for item with prices
	for _, it := range items {
		if it.TypeID == int64(34) {
			// FillPercent should be 0 (0 acquired / 100 planned * 100)
			assert.Equal(t, 0.0, it.FillPercent)
			// NetProfitISK = (1500 - 1000) * 100 = 50000
			assert.NotNil(t, it.NetProfitISK)
			assert.Equal(t, 50000.0, *it.NetProfitISK)
		}
		if it.TypeID == int64(35) {
			assert.Nil(t, it.NetProfitISK)
		}
	}
}

func Test_HaulingRunItems_GetItemsByRunID_Empty(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9120)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Item User 3"})
	assert.NoError(t, err)

	runsRepo := repositories.NewHaulingRuns(db)
	run, err := runsRepo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Empty Items Run",
		Status:       "PLANNING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	})
	assert.NoError(t, err)

	itemsRepo := repositories.NewHaulingRunItems(db)
	items, err := itemsRepo.GetItemsByRunID(context.Background(), run.ID)
	assert.NoError(t, err)
	assert.NotNil(t, items)
	assert.Len(t, items, 0)
}

func Test_HaulingRunItems_FillPercent(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9130)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Item User 4"})
	assert.NoError(t, err)

	runsRepo := repositories.NewHaulingRuns(db)
	run, err := runsRepo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Fill Test Run",
		Status:       "ACCUMULATING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	})
	assert.NoError(t, err)

	itemsRepo := repositories.NewHaulingRunItems(db)
	added, err := itemsRepo.AddItem(context.Background(), &models.HaulingRunItem{
		RunID:             run.ID,
		TypeID:            int64(36),
		TypeName:          "Mexallon",
		QuantityPlanned:   int64(200),
		QuantityAcquired:  int64(50),
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(50), added.QuantityAcquired)

	items, err := itemsRepo.GetItemsByRunID(context.Background(), run.ID)
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	// 50 / 200 * 100 = 25%
	assert.Equal(t, 25.0, items[0].FillPercent)
}

func Test_HaulingRunItems_UpdateItemAcquired(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9140)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Item User 5"})
	assert.NoError(t, err)

	runsRepo := repositories.NewHaulingRuns(db)
	run, err := runsRepo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Update Acquired Run",
		Status:       "ACCUMULATING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	})
	assert.NoError(t, err)

	itemsRepo := repositories.NewHaulingRunItems(db)
	added, err := itemsRepo.AddItem(context.Background(), &models.HaulingRunItem{
		RunID:           run.ID,
		TypeID:          int64(37),
		TypeName:        "Isogen",
		QuantityPlanned: int64(100),
	})
	assert.NoError(t, err)

	err = itemsRepo.UpdateItemAcquired(context.Background(), added.ID, run.ID, int64(75))
	assert.NoError(t, err)

	items, err := itemsRepo.GetItemsByRunID(context.Background(), run.ID)
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, int64(75), items[0].QuantityAcquired)
	assert.Equal(t, 75.0, items[0].FillPercent)
}

func Test_HaulingRunItems_UpdateItemAcquired_NotFound(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	itemsRepo := repositories.NewHaulingRunItems(db)
	err = itemsRepo.UpdateItemAcquired(context.Background(), int64(999999999), int64(1), int64(10))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func Test_HaulingRunItems_RemoveItem(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9150)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Item User 6"})
	assert.NoError(t, err)

	runsRepo := repositories.NewHaulingRuns(db)
	run, err := runsRepo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Remove Item Run",
		Status:       "PLANNING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	})
	assert.NoError(t, err)

	itemsRepo := repositories.NewHaulingRunItems(db)
	added, err := itemsRepo.AddItem(context.Background(), &models.HaulingRunItem{
		RunID:           run.ID,
		TypeID:          int64(38),
		TypeName:        "Nocxium",
		QuantityPlanned: int64(10),
	})
	assert.NoError(t, err)

	err = itemsRepo.RemoveItem(context.Background(), added.ID, run.ID)
	assert.NoError(t, err)

	items, err := itemsRepo.GetItemsByRunID(context.Background(), run.ID)
	assert.NoError(t, err)
	assert.Len(t, items, 0)
}

func Test_HaulingRunItems_RemoveItem_NotFound(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	itemsRepo := repositories.NewHaulingRunItems(db)
	err = itemsRepo.RemoveItem(context.Background(), int64(999999999), int64(1))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func Test_HaulingRunItems_CascadeDeleteWithRun(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9160)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Item User 7"})
	assert.NoError(t, err)

	runsRepo := repositories.NewHaulingRuns(db)
	run, err := runsRepo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Cascade Test Run",
		Status:       "PLANNING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	})
	assert.NoError(t, err)

	itemsRepo := repositories.NewHaulingRunItems(db)
	_, err = itemsRepo.AddItem(context.Background(), &models.HaulingRunItem{
		RunID:           run.ID,
		TypeID:          int64(39),
		TypeName:        "Zydrine",
		QuantityPlanned: int64(5),
	})
	assert.NoError(t, err)

	// Delete the run - should cascade
	err = runsRepo.DeleteRun(context.Background(), run.ID, userID)
	assert.NoError(t, err)

	// Items should be gone
	items, err := itemsRepo.GetItemsByRunID(context.Background(), run.ID)
	assert.NoError(t, err)
	assert.Len(t, items, 0)
}
