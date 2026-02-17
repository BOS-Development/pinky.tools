package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_ContactPermissionsShouldUpsert(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	contactsRepo := repositories.NewContacts(db)
	permissionsRepo := repositories.NewContactPermissions(db)

	// Setup users and characters
	user1 := &repositories.User{ID: 1100, Name: "User 1"}
	user2 := &repositories.User{ID: 1101, Name: "User 2"}

	err = userRepo.Add(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Add(context.Background(), user2)
	assert.NoError(t, err)

	char1 := &repositories.Character{ID: 11000, Name: "Character 1", UserID: user1.ID}
	char2 := &repositories.Character{ID: 11001, Name: "Character 2", UserID: user2.ID}

	err = characterRepo.Add(context.Background(), char1)
	assert.NoError(t, err)
	err = characterRepo.Add(context.Background(), char2)
	assert.NoError(t, err)

	// Create contact
	contact, err := contactsRepo.Create(context.Background(), user1.ID, user2.ID)
	assert.NoError(t, err)

	// Upsert permission
	perm := &models.ContactPermission{
		ContactID:       contact.ID,
		GrantingUserID:  user1.ID,
		ReceivingUserID: user2.ID,
		ServiceType:     "for_sale_browse",
		CanAccess:       true,
	}
	err = permissionsRepo.Upsert(context.Background(), perm)
	assert.NoError(t, err)

	// Verify permission
	hasPermission, err := permissionsRepo.CheckPermission(context.Background(), user1.ID, user2.ID, "for_sale_browse")
	assert.NoError(t, err)
	assert.True(t, hasPermission)
}

func Test_ContactPermissionsShouldUpdateExisting(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	contactsRepo := repositories.NewContacts(db)
	permissionsRepo := repositories.NewContactPermissions(db)

	user1 := &repositories.User{ID: 1200, Name: "User 1"}
	user2 := &repositories.User{ID: 1201, Name: "User 2"}

	err = userRepo.Add(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Add(context.Background(), user2)
	assert.NoError(t, err)

	char1 := &repositories.Character{ID: 12000, Name: "Character 1", UserID: user1.ID}
	char2 := &repositories.Character{ID: 12001, Name: "Character 2", UserID: user2.ID}

	err = characterRepo.Add(context.Background(), char1)
	assert.NoError(t, err)
	err = characterRepo.Add(context.Background(), char2)
	assert.NoError(t, err)

	contact, err := contactsRepo.Create(context.Background(), user1.ID, user2.ID)
	assert.NoError(t, err)

	// Create permission as false
	perm := &models.ContactPermission{
		ContactID:       contact.ID,
		GrantingUserID:  user1.ID,
		ReceivingUserID: user2.ID,
		ServiceType:     "for_sale_browse",
		CanAccess:       false,
	}
	err = permissionsRepo.Upsert(context.Background(), perm)
	assert.NoError(t, err)

	// Update to true
	perm.CanAccess = true
	err = permissionsRepo.Upsert(context.Background(), perm)
	assert.NoError(t, err)

	// Verify updated permission
	hasPermission, err := permissionsRepo.CheckPermission(context.Background(), user1.ID, user2.ID, "for_sale_browse")
	assert.NoError(t, err)
	assert.True(t, hasPermission)
}

func Test_ContactPermissionsShouldCheckPermission(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	contactsRepo := repositories.NewContacts(db)
	permissionsRepo := repositories.NewContactPermissions(db)

	user1 := &repositories.User{ID: 1300, Name: "User 1"}
	user2 := &repositories.User{ID: 1301, Name: "User 2"}

	err = userRepo.Add(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Add(context.Background(), user2)
	assert.NoError(t, err)

	char1 := &repositories.Character{ID: 13000, Name: "Character 1", UserID: user1.ID}
	char2 := &repositories.Character{ID: 13001, Name: "Character 2", UserID: user2.ID}

	err = characterRepo.Add(context.Background(), char1)
	assert.NoError(t, err)
	err = characterRepo.Add(context.Background(), char2)
	assert.NoError(t, err)

	contact, err := contactsRepo.Create(context.Background(), user1.ID, user2.ID)
	assert.NoError(t, err)

	// Grant permission
	perm := &models.ContactPermission{
		ContactID:       contact.ID,
		GrantingUserID:  user1.ID,
		ReceivingUserID: user2.ID,
		ServiceType:     "for_sale_browse",
		CanAccess:       true,
	}
	err = permissionsRepo.Upsert(context.Background(), perm)
	assert.NoError(t, err)

	// Check permission exists
	hasPermission, err := permissionsRepo.CheckPermission(context.Background(), user1.ID, user2.ID, "for_sale_browse")
	assert.NoError(t, err)
	assert.True(t, hasPermission)

	// Check non-existent permission
	hasPermission, err = permissionsRepo.CheckPermission(context.Background(), user2.ID, user1.ID, "for_sale_browse")
	assert.NoError(t, err)
	assert.False(t, hasPermission)
}

func Test_ContactPermissionsShouldGetUserPermissionsForService(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	contactsRepo := repositories.NewContacts(db)
	permissionsRepo := repositories.NewContactPermissions(db)

	user1 := &repositories.User{ID: 1400, Name: "User 1"}
	user2 := &repositories.User{ID: 1401, Name: "User 2"}
	user3 := &repositories.User{ID: 1402, Name: "User 3"}

	err = userRepo.Add(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Add(context.Background(), user2)
	assert.NoError(t, err)
	err = userRepo.Add(context.Background(), user3)
	assert.NoError(t, err)

	char1 := &repositories.Character{ID: 14000, Name: "Character 1", UserID: user1.ID}
	char2 := &repositories.Character{ID: 14001, Name: "Character 2", UserID: user2.ID}
	char3 := &repositories.Character{ID: 14002, Name: "Character 3", UserID: user3.ID}

	err = characterRepo.Add(context.Background(), char1)
	assert.NoError(t, err)
	err = characterRepo.Add(context.Background(), char2)
	assert.NoError(t, err)
	err = characterRepo.Add(context.Background(), char3)
	assert.NoError(t, err)

	// Create contacts
	contact1, err := contactsRepo.Create(context.Background(), user1.ID, user2.ID)
	assert.NoError(t, err)

	contact2, err := contactsRepo.Create(context.Background(), user1.ID, user3.ID)
	assert.NoError(t, err)

	// Grant permission from user2 to user1
	perm1 := &models.ContactPermission{
		ContactID:       contact1.ID,
		GrantingUserID:  user2.ID,
		ReceivingUserID: user1.ID,
		ServiceType:     "for_sale_browse",
		CanAccess:       true,
	}
	err = permissionsRepo.Upsert(context.Background(), perm1)
	assert.NoError(t, err)

	// Grant permission from user3 to user1
	perm2 := &models.ContactPermission{
		ContactID:       contact2.ID,
		GrantingUserID:  user3.ID,
		ReceivingUserID: user1.ID,
		ServiceType:     "for_sale_browse",
		CanAccess:       true,
	}
	err = permissionsRepo.Upsert(context.Background(), perm2)
	assert.NoError(t, err)

	// Get all users who granted permission to user1
	userIDs, err := permissionsRepo.GetUserPermissionsForService(context.Background(), user1.ID, "for_sale_browse")
	assert.NoError(t, err)
	assert.Len(t, userIDs, 2)
	assert.Contains(t, userIDs, user2.ID)
	assert.Contains(t, userIDs, user3.ID)
}

func Test_ContactPermissionsShouldInitializePermissionsForContact(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	contactsRepo := repositories.NewContacts(db)
	permissionsRepo := repositories.NewContactPermissions(db)

	user1 := &repositories.User{ID: 1500, Name: "User 1"}
	user2 := &repositories.User{ID: 1501, Name: "User 2"}

	err = userRepo.Add(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Add(context.Background(), user2)
	assert.NoError(t, err)

	char1 := &repositories.Character{ID: 15000, Name: "Character 1", UserID: user1.ID}
	char2 := &repositories.Character{ID: 15001, Name: "Character 2", UserID: user2.ID}

	err = characterRepo.Add(context.Background(), char1)
	assert.NoError(t, err)
	err = characterRepo.Add(context.Background(), char2)
	assert.NoError(t, err)

	contact, err := contactsRepo.Create(context.Background(), user1.ID, user2.ID)
	assert.NoError(t, err)

	// Initialize permissions
	tx, err := db.BeginTx(context.Background(), nil)
	assert.NoError(t, err)
	defer tx.Rollback()

	err = permissionsRepo.InitializePermissionsForContact(context.Background(), tx, contact.ID, user1.ID, user2.ID)
	assert.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	// Verify both users have permissions (all set to false initially)
	hasPermission1, err := permissionsRepo.CheckPermission(context.Background(), user1.ID, user2.ID, "for_sale_browse")
	assert.NoError(t, err)
	assert.False(t, hasPermission1)

	hasPermission2, err := permissionsRepo.CheckPermission(context.Background(), user2.ID, user1.ID, "for_sale_browse")
	assert.NoError(t, err)
	assert.False(t, hasPermission2)
}

func Test_ContactPermissionsShouldGetByContact(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	contactsRepo := repositories.NewContacts(db)
	permissionsRepo := repositories.NewContactPermissions(db)

	user1 := &repositories.User{ID: 1600, Name: "User 1"}
	user2 := &repositories.User{ID: 1601, Name: "User 2"}

	err = userRepo.Add(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Add(context.Background(), user2)
	assert.NoError(t, err)

	char1 := &repositories.Character{ID: 16000, Name: "Character 1", UserID: user1.ID}
	char2 := &repositories.Character{ID: 16001, Name: "Character 2", UserID: user2.ID}

	err = characterRepo.Add(context.Background(), char1)
	assert.NoError(t, err)
	err = characterRepo.Add(context.Background(), char2)
	assert.NoError(t, err)

	contact, err := contactsRepo.Create(context.Background(), user1.ID, user2.ID)
	assert.NoError(t, err)

	// Create multiple permissions
	perm1 := &models.ContactPermission{
		ContactID:       contact.ID,
		GrantingUserID:  user1.ID,
		ReceivingUserID: user2.ID,
		ServiceType:     "for_sale_browse",
		CanAccess:       true,
	}
	err = permissionsRepo.Upsert(context.Background(), perm1)
	assert.NoError(t, err)

	perm2 := &models.ContactPermission{
		ContactID:       contact.ID,
		GrantingUserID:  user2.ID,
		ReceivingUserID: user1.ID,
		ServiceType:     "for_sale_browse",
		CanAccess:       false,
	}
	err = permissionsRepo.Upsert(context.Background(), perm2)
	assert.NoError(t, err)

	// Get all permissions for the contact (as user1)
	permissions, err := permissionsRepo.GetByContact(context.Background(), contact.ID, user1.ID)
	assert.NoError(t, err)
	assert.Len(t, permissions, 2)

	// Verify permissions
	foundGranted := false
	foundNotGranted := false
	for _, perm := range permissions {
		if perm.GrantingUserID == user1.ID && perm.CanAccess {
			foundGranted = true
		}
		if perm.GrantingUserID == user2.ID && !perm.CanAccess {
			foundNotGranted = true
		}
	}
	assert.True(t, foundGranted)
	assert.True(t, foundNotGranted)
}

func Test_ContactPermissionsShouldBeUnidirectional(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	contactsRepo := repositories.NewContacts(db)
	permissionsRepo := repositories.NewContactPermissions(db)

	user1 := &repositories.User{ID: 1700, Name: "User 1"}
	user2 := &repositories.User{ID: 1701, Name: "User 2"}

	err = userRepo.Add(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Add(context.Background(), user2)
	assert.NoError(t, err)

	char1 := &repositories.Character{ID: 17000, Name: "Character 1", UserID: user1.ID}
	char2 := &repositories.Character{ID: 17001, Name: "Character 2", UserID: user2.ID}

	err = characterRepo.Add(context.Background(), char1)
	assert.NoError(t, err)
	err = characterRepo.Add(context.Background(), char2)
	assert.NoError(t, err)

	contact, err := contactsRepo.Create(context.Background(), user1.ID, user2.ID)
	assert.NoError(t, err)

	// User1 grants permission to User2
	perm1 := &models.ContactPermission{
		ContactID:       contact.ID,
		GrantingUserID:  user1.ID,
		ReceivingUserID: user2.ID,
		ServiceType:     "for_sale_browse",
		CanAccess:       true,
	}
	err = permissionsRepo.Upsert(context.Background(), perm1)
	assert.NoError(t, err)

	// User2 does NOT grant permission to User1
	perm2 := &models.ContactPermission{
		ContactID:       contact.ID,
		GrantingUserID:  user2.ID,
		ReceivingUserID: user1.ID,
		ServiceType:     "for_sale_browse",
		CanAccess:       false,
	}
	err = permissionsRepo.Upsert(context.Background(), perm2)
	assert.NoError(t, err)

	// Verify unidirectional permission
	hasPermission1to2, err := permissionsRepo.CheckPermission(context.Background(), user1.ID, user2.ID, "for_sale_browse")
	assert.NoError(t, err)
	assert.True(t, hasPermission1to2)

	hasPermission2to1, err := permissionsRepo.CheckPermission(context.Background(), user2.ID, user1.ID, "for_sale_browse")
	assert.NoError(t, err)
	assert.False(t, hasPermission2to1)
}
