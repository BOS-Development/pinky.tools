package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_UserShouldAddAndGet(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewUserRepository(db)

	user := &repositories.User{
		ID:   42,
		Name: "Test User",
	}

	err = repo.Add(context.Background(), user)
	assert.NoError(t, err)

	retrieved, err := repo.Get(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Name, retrieved.Name)
}

func Test_UserGetReturnsNilForNonExistent(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewUserRepository(db)

	retrieved, err := repo.Get(context.Background(), 99999)
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func Test_UserGetAllIDs(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewUserRepository(db)

	// Empty database should return empty slice
	ids, err := repo.GetAllIDs(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, ids)

	// Add multiple users
	users := []*repositories.User{
		{ID: 1001, Name: "Alice"},
		{ID: 1002, Name: "Bob"},
		{ID: 1003, Name: "Charlie"},
	}

	for _, u := range users {
		err = repo.Add(context.Background(), u)
		assert.NoError(t, err)
	}

	ids, err = repo.GetAllIDs(context.Background())
	assert.NoError(t, err)
	assert.Len(t, ids, 3)

	// All IDs should be present
	idSet := make(map[int64]bool)
	for _, id := range ids {
		idSet[id] = true
	}
	assert.True(t, idSet[1001])
	assert.True(t, idSet[1002])
	assert.True(t, idSet[1003])
}

func Test_UserUpdateAssetsLastUpdated(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewUserRepository(db)

	user := &repositories.User{ID: 42, Name: "Test User"}
	err = repo.Add(context.Background(), user)
	assert.NoError(t, err)

	// Before update, should be nil
	lastUpdated, err := repo.GetAssetsLastUpdated(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Nil(t, lastUpdated)

	// Update the timestamp
	before := time.Now().Add(-1 * time.Second)
	err = repo.UpdateAssetsLastUpdated(context.Background(), user.ID)
	assert.NoError(t, err)
	after := time.Now().Add(1 * time.Second)

	// Should now return a timestamp between before and after
	lastUpdated, err = repo.GetAssetsLastUpdated(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, lastUpdated)
	assert.True(t, lastUpdated.After(before), "lastUpdated should be after test start")
	assert.True(t, lastUpdated.Before(after), "lastUpdated should be before test end")
}

func Test_UserUpdateAssetsLastUpdatedOverwritesPrevious(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewUserRepository(db)

	user := &repositories.User{ID: 42, Name: "Test User"}
	err = repo.Add(context.Background(), user)
	assert.NoError(t, err)

	// First update
	err = repo.UpdateAssetsLastUpdated(context.Background(), user.ID)
	assert.NoError(t, err)

	firstUpdated, err := repo.GetAssetsLastUpdated(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, firstUpdated)

	// Small delay to ensure different timestamp
	time.Sleep(10 * time.Millisecond)

	// Second update
	err = repo.UpdateAssetsLastUpdated(context.Background(), user.ID)
	assert.NoError(t, err)

	secondUpdated, err := repo.GetAssetsLastUpdated(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, secondUpdated)

	assert.True(t, secondUpdated.After(*firstUpdated), "second update should be after first")
}

func Test_UserGetAssetsLastUpdatedIsolatedPerUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewUserRepository(db)

	user1 := &repositories.User{ID: 1001, Name: "Alice"}
	user2 := &repositories.User{ID: 1002, Name: "Bob"}

	err = repo.Add(context.Background(), user1)
	assert.NoError(t, err)
	err = repo.Add(context.Background(), user2)
	assert.NoError(t, err)

	// Update only user1's timestamp
	err = repo.UpdateAssetsLastUpdated(context.Background(), user1.ID)
	assert.NoError(t, err)

	// User1 should have a timestamp
	lastUpdated1, err := repo.GetAssetsLastUpdated(context.Background(), user1.ID)
	assert.NoError(t, err)
	assert.NotNil(t, lastUpdated1)

	// User2 should still be nil
	lastUpdated2, err := repo.GetAssetsLastUpdated(context.Background(), user2.ID)
	assert.NoError(t, err)
	assert.Nil(t, lastUpdated2)
}

func Test_UserGetAssetsLastUpdatedNonExistentUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewUserRepository(db)

	// Non-existent user should return an error (sql.ErrNoRows wrapped)
	_, err = repo.GetAssetsLastUpdated(context.Background(), 99999)
	assert.Error(t, err)
}
