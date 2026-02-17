package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_ContactsShouldCreateAndGet(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	contactsRepo := repositories.NewContacts(db)

	// Create test users
	user1 := &repositories.User{ID: 100, Name: "User 1"}
	user2 := &repositories.User{ID: 101, Name: "User 2"}

	err = userRepo.Add(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Add(context.Background(), user2)
	assert.NoError(t, err)

	// Create test characters
	char1 := &repositories.Character{ID: 1000, Name: "Character 1", UserID: user1.ID}
	char2 := &repositories.Character{ID: 1001, Name: "Character 2", UserID: user2.ID}

	err = characterRepo.Add(context.Background(), char1)
	assert.NoError(t, err)
	err = characterRepo.Add(context.Background(), char2)
	assert.NoError(t, err)

	// Create contact request
	contact, err := contactsRepo.Create(context.Background(), user1.ID, user2.ID)
	assert.NoError(t, err)
	assert.NotNil(t, contact)
	assert.Equal(t, user1.ID, contact.RequesterUserID)
	assert.Equal(t, user2.ID, contact.RecipientUserID)
	assert.Equal(t, "pending", contact.Status)

	// Get contacts for user1
	contacts, err := contactsRepo.GetByUser(context.Background(), user1.ID)
	assert.NoError(t, err)
	assert.Len(t, contacts, 1)
	assert.Equal(t, contact.ID, contacts[0].ID)
	assert.Equal(t, "User 1", contacts[0].RequesterName)
	assert.Equal(t, "User 2", contacts[0].RecipientName)
}

func Test_ContactsShouldUpdateStatus(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	contactsRepo := repositories.NewContacts(db)

	user1 := &repositories.User{ID: 200, Name: "User 1"}
	user2 := &repositories.User{ID: 201, Name: "User 2"}

	err = userRepo.Add(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Add(context.Background(), user2)
	assert.NoError(t, err)

	char1 := &repositories.Character{ID: 2000, Name: "Character 1", UserID: user1.ID}
	char2 := &repositories.Character{ID: 2001, Name: "Character 2", UserID: user2.ID}

	err = characterRepo.Add(context.Background(), char1)
	assert.NoError(t, err)
	err = characterRepo.Add(context.Background(), char2)
	assert.NoError(t, err)

	contact, err := contactsRepo.Create(context.Background(), user1.ID, user2.ID)
	assert.NoError(t, err)

	// Accept contact as recipient (user2)
	updatedContact, err := contactsRepo.UpdateStatus(context.Background(), contact.ID, user2.ID, "accepted")
	assert.NoError(t, err)
	assert.Equal(t, "accepted", updatedContact.Status)
	assert.NotNil(t, updatedContact.RespondedAt)
}

func Test_ContactsShouldRejectStatus(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	contactsRepo := repositories.NewContacts(db)

	user1 := &repositories.User{ID: 300, Name: "User 1"}
	user2 := &repositories.User{ID: 301, Name: "User 2"}

	err = userRepo.Add(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Add(context.Background(), user2)
	assert.NoError(t, err)

	char1 := &repositories.Character{ID: 3000, Name: "Character 1", UserID: user1.ID}
	char2 := &repositories.Character{ID: 3001, Name: "Character 2", UserID: user2.ID}

	err = characterRepo.Add(context.Background(), char1)
	assert.NoError(t, err)
	err = characterRepo.Add(context.Background(), char2)
	assert.NoError(t, err)

	contact, err := contactsRepo.Create(context.Background(), user1.ID, user2.ID)
	assert.NoError(t, err)

	// Reject contact as recipient
	updatedContact, err := contactsRepo.UpdateStatus(context.Background(), contact.ID, user2.ID, "rejected")
	assert.NoError(t, err)
	assert.Equal(t, "rejected", updatedContact.Status)
	assert.NotNil(t, updatedContact.RespondedAt)
}

func Test_ContactsShouldDelete(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	contactsRepo := repositories.NewContacts(db)

	user1 := &repositories.User{ID: 400, Name: "User 1"}
	user2 := &repositories.User{ID: 401, Name: "User 2"}

	err = userRepo.Add(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Add(context.Background(), user2)
	assert.NoError(t, err)

	char1 := &repositories.Character{ID: 4000, Name: "Character 1", UserID: user1.ID}
	char2 := &repositories.Character{ID: 4001, Name: "Character 2", UserID: user2.ID}

	err = characterRepo.Add(context.Background(), char1)
	assert.NoError(t, err)
	err = characterRepo.Add(context.Background(), char2)
	assert.NoError(t, err)

	contact, err := contactsRepo.Create(context.Background(), user1.ID, user2.ID)
	assert.NoError(t, err)

	// Delete as requester
	err = contactsRepo.Delete(context.Background(), contact.ID, user1.ID)
	assert.NoError(t, err)

	// Verify deletion
	contacts, err := contactsRepo.GetByUser(context.Background(), user1.ID)
	assert.NoError(t, err)
	assert.Len(t, contacts, 0)
}

func Test_ContactsShouldPreventDuplicates(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	contactsRepo := repositories.NewContacts(db)

	user1 := &repositories.User{ID: 500, Name: "User 1"}
	user2 := &repositories.User{ID: 501, Name: "User 2"}

	err = userRepo.Add(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Add(context.Background(), user2)
	assert.NoError(t, err)

	char1 := &repositories.Character{ID: 5000, Name: "Character 1", UserID: user1.ID}
	char2 := &repositories.Character{ID: 5001, Name: "Character 2", UserID: user2.ID}

	err = characterRepo.Add(context.Background(), char1)
	assert.NoError(t, err)
	err = characterRepo.Add(context.Background(), char2)
	assert.NoError(t, err)

	// Create first contact
	_, err = contactsRepo.Create(context.Background(), user1.ID, user2.ID)
	assert.NoError(t, err)

	// Try to create duplicate
	_, err = contactsRepo.Create(context.Background(), user1.ID, user2.ID)
	assert.Error(t, err)
}

func Test_ContactsShouldGetByUserBidirectional(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	contactsRepo := repositories.NewContacts(db)

	user1 := &repositories.User{ID: 600, Name: "User 1"}
	user2 := &repositories.User{ID: 601, Name: "User 2"}
	user3 := &repositories.User{ID: 602, Name: "User 3"}

	err = userRepo.Add(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Add(context.Background(), user2)
	assert.NoError(t, err)
	err = userRepo.Add(context.Background(), user3)
	assert.NoError(t, err)

	char1 := &repositories.Character{ID: 6000, Name: "Character 1", UserID: user1.ID}
	char2 := &repositories.Character{ID: 6001, Name: "Character 2", UserID: user2.ID}
	char3 := &repositories.Character{ID: 6002, Name: "Character 3", UserID: user3.ID}

	err = characterRepo.Add(context.Background(), char1)
	assert.NoError(t, err)
	err = characterRepo.Add(context.Background(), char2)
	assert.NoError(t, err)
	err = characterRepo.Add(context.Background(), char3)
	assert.NoError(t, err)

	// User1 requests User2
	_, err = contactsRepo.Create(context.Background(), user1.ID, user2.ID)
	assert.NoError(t, err)

	// User3 requests User1
	_, err = contactsRepo.Create(context.Background(), user3.ID, user1.ID)
	assert.NoError(t, err)

	// User1 should see both contacts (one as requester, one as recipient)
	contacts, err := contactsRepo.GetByUser(context.Background(), user1.ID)
	assert.NoError(t, err)
	assert.Len(t, contacts, 2)
}

func Test_ContactsShouldGetUserIDByCharacterName(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	contactsRepo := repositories.NewContacts(db)

	user := &repositories.User{ID: 700, Name: "Test User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 7000, Name: "Test Character", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	// Get user ID by character name
	userID, err := contactsRepo.GetUserIDByCharacterName(context.Background(), "Test Character")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, userID)
}

func Test_ContactsShouldGetUserIDByCharacterNameCaseInsensitive(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	contactsRepo := repositories.NewContacts(db)

	user := &repositories.User{ID: 800, Name: "Test User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 8000, Name: "Test Character", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	// Get user ID with different case
	userID, err := contactsRepo.GetUserIDByCharacterName(context.Background(), "test character")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, userID)
}

func Test_ContactsShouldGetUserIDByCharacterID(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	contactsRepo := repositories.NewContacts(db)

	user := &repositories.User{ID: 900, Name: "Test User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 9000, Name: "Test Character", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	// Get user ID by character ID
	userID, err := contactsRepo.GetUserIDByCharacterID(context.Background(), char.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, userID)
}

func Test_ContactsShouldNotUpdateStatusIfNotRecipient(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	contactsRepo := repositories.NewContacts(db)

	user1 := &repositories.User{ID: 1000, Name: "User 1"}
	user2 := &repositories.User{ID: 1001, Name: "User 2"}
	user3 := &repositories.User{ID: 1002, Name: "User 3"}

	err = userRepo.Add(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Add(context.Background(), user2)
	assert.NoError(t, err)
	err = userRepo.Add(context.Background(), user3)
	assert.NoError(t, err)

	char1 := &repositories.Character{ID: 10000, Name: "Character 1", UserID: user1.ID}
	char2 := &repositories.Character{ID: 10001, Name: "Character 2", UserID: user2.ID}
	char3 := &repositories.Character{ID: 10002, Name: "Character 3", UserID: user3.ID}

	err = characterRepo.Add(context.Background(), char1)
	assert.NoError(t, err)
	err = characterRepo.Add(context.Background(), char2)
	assert.NoError(t, err)
	err = characterRepo.Add(context.Background(), char3)
	assert.NoError(t, err)

	contact, err := contactsRepo.Create(context.Background(), user1.ID, user2.ID)
	assert.NoError(t, err)

	// Try to update as user3 (not the recipient)
	_, err = contactsRepo.UpdateStatus(context.Background(), contact.ID, user3.ID, "accepted")
	assert.Error(t, err)
}
