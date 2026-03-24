package controllers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_ListManufacturingProfiles_Returns403_WhenNotEnabled(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(200)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(false, nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/manufacturing-profiles", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.ListManufacturingProfiles(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 403, httpErr.StatusCode)
	mocks.settings.AssertExpectations(t)
}

func Test_ListManufacturingProfiles_ReturnsProfiles(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(200)

	profiles := []*models.ArbiterManufacturingProfile{
		{ID: 1, UserID: userID, Name: "Profile A", FinalStructure: "raitaru", FinalRig: "t2"},
		{ID: 2, UserID: userID, Name: "Profile B", FinalStructure: "azbel", FinalRig: "t1"},
	}

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.mfgProfile.On("ListManufacturingProfiles", mock.Anything, userID).Return(profiles, nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/manufacturing-profiles", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.ListManufacturingProfiles(args)
	assert.Nil(t, httpErr)
	assert.Equal(t, profiles, result)
	mocks.settings.AssertExpectations(t)
	mocks.mfgProfile.AssertExpectations(t)
}

func Test_ListManufacturingProfiles_Returns500_OnRepoError(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(200)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.mfgProfile.On("ListManufacturingProfiles", mock.Anything, userID).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/arbiter/manufacturing-profiles", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.ListManufacturingProfiles(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

func Test_CreateManufacturingProfile_Returns403_WhenNotEnabled(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(200)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(false, nil)

	body, _ := json.Marshal(map[string]interface{}{"name": "My Profile"})
	req := httptest.NewRequest("POST", "/v1/arbiter/manufacturing-profiles", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.CreateManufacturingProfile(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 403, httpErr.StatusCode)
}

func Test_CreateManufacturingProfile_Returns400_WhenNameMissing(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(200)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)

	body, _ := json.Marshal(map[string]interface{}{"final_structure": "raitaru"})
	req := httptest.NewRequest("POST", "/v1/arbiter/manufacturing-profiles", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.CreateManufacturingProfile(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_CreateManufacturingProfile_Success(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(200)

	reqBody := map[string]interface{}{
		"name":                 "Jita Raitaru",
		"reaction_structure":   "athanor",
		"reaction_rig":         "t1",
		"reaction_facility_tax": 0.3,
		"invention_structure":  "raitaru",
		"invention_rig":        "t1",
		"component_structure":  "raitaru",
		"component_rig":        "t2",
		"final_structure":      "raitaru",
		"final_rig":            "t2",
		"final_facility_tax":   0.5,
	}

	createdProfile := &models.ArbiterManufacturingProfile{
		ID:                  42,
		UserID:              userID,
		Name:                "Jita Raitaru",
		ReactionStructure:   "athanor",
		ReactionRig:         "t1",
		ReactionFacilityTax: 0.3,
		InventionStructure:  "raitaru",
		InventionRig:        "t1",
		ComponentStructure:  "raitaru",
		ComponentRig:        "t2",
		FinalStructure:      "raitaru",
		FinalRig:            "t2",
		FinalFacilityTax:    0.5,
	}

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.mfgProfile.On("CreateManufacturingProfile", mock.Anything, mock.MatchedBy(func(p *models.ArbiterManufacturingProfile) bool {
		return p.UserID == userID && p.Name == "Jita Raitaru" && p.FinalStructure == "raitaru"
	})).Return(createdProfile, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/arbiter/manufacturing-profiles", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.CreateManufacturingProfile(args)
	assert.Nil(t, httpErr)
	assert.Equal(t, createdProfile, result)
	mocks.mfgProfile.AssertExpectations(t)
}

func Test_UpdateManufacturingProfile_Returns404_WhenNotFound(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(200)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.mfgProfile.On("UpdateManufacturingProfile", mock.Anything, mock.Anything).Return(nil, nil)

	body, _ := json.Marshal(map[string]interface{}{"name": "Updated"})
	req := httptest.NewRequest("PUT", "/v1/arbiter/manufacturing-profiles/99", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "99"}}

	result, httpErr := c.UpdateManufacturingProfile(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

func Test_UpdateManufacturingProfile_Returns400_WhenInvalidID(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(200)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)

	body, _ := json.Marshal(map[string]interface{}{"name": "Updated"})
	req := httptest.NewRequest("PUT", "/v1/arbiter/manufacturing-profiles/abc", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "abc"}}

	result, httpErr := c.UpdateManufacturingProfile(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_UpdateManufacturingProfile_Success(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(200)

	updatedProfile := &models.ArbiterManufacturingProfile{
		ID:             10,
		UserID:         userID,
		Name:           "New Name",
		FinalStructure: "azbel",
		FinalRig:       "t2",
	}

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.mfgProfile.On("UpdateManufacturingProfile", mock.Anything, mock.MatchedBy(func(p *models.ArbiterManufacturingProfile) bool {
		return p.ID == int64(10) && p.UserID == userID && p.Name == "New Name"
	})).Return(updatedProfile, nil)

	body, _ := json.Marshal(map[string]interface{}{
		"name":            "New Name",
		"final_structure": "azbel",
		"final_rig":       "t2",
	})
	req := httptest.NewRequest("PUT", "/v1/arbiter/manufacturing-profiles/10", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "10"}}

	result, httpErr := c.UpdateManufacturingProfile(args)
	assert.Nil(t, httpErr)
	assert.Equal(t, updatedProfile, result)
	mocks.mfgProfile.AssertExpectations(t)
}

func Test_DeleteManufacturingProfile_Returns403_WhenNotEnabled(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(200)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(false, nil)

	req := httptest.NewRequest("DELETE", "/v1/arbiter/manufacturing-profiles/5", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := c.DeleteManufacturingProfile(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 403, httpErr.StatusCode)
}

func Test_DeleteManufacturingProfile_Success(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(200)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.mfgProfile.On("DeleteManufacturingProfile", mock.Anything, int64(5), userID).Return(nil)

	req := httptest.NewRequest("DELETE", "/v1/arbiter/manufacturing-profiles/5", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := c.DeleteManufacturingProfile(args)
	assert.Nil(t, result)
	assert.Nil(t, httpErr)
	mocks.mfgProfile.AssertExpectations(t)
}

func Test_DeleteManufacturingProfile_Returns400_WhenInvalidID(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(200)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)

	req := httptest.NewRequest("DELETE", "/v1/arbiter/manufacturing-profiles/notanid", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "notanid"}}

	result, httpErr := c.DeleteManufacturingProfile(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_ApplyManufacturingProfile_Returns404_WhenProfileNotFound(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(200)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.mfgProfile.On("GetManufacturingProfile", mock.Anything, int64(99), userID).Return(nil, nil)

	req := httptest.NewRequest("POST", "/v1/arbiter/manufacturing-profiles/99/apply", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "99"}}

	result, httpErr := c.ApplyManufacturingProfile(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

func Test_ApplyManufacturingProfile_UpdatesSettingsWithProfileValues(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(200)

	profile := &models.ArbiterManufacturingProfile{
		ID:                   7,
		UserID:               userID,
		Name:                 "My Profile",
		ReactionStructure:    "tatara",
		ReactionRig:          "t2",
		ReactionFacilityTax:  0.3,
		InventionStructure:   "raitaru",
		InventionRig:         "t1",
		InventionFacilityTax: 0.0,
		ComponentStructure:   "azbel",
		ComponentRig:         "t2",
		ComponentFacilityTax: 0.5,
		FinalStructure:       "sotiyo",
		FinalRig:             "t2",
		FinalFacilityTax:     1.0,
	}

	existingSettings := &models.ArbiterSettings{
		UserID:          userID,
		UseWhitelist:    true,
		UseBlacklist:    false,
		ReactionStructure: "athanor",
		FinalStructure:  "raitaru",
	}

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.mfgProfile.On("GetManufacturingProfile", mock.Anything, int64(7), userID).Return(profile, nil)
	mocks.settings.On("GetArbiterSettings", mock.Anything, userID).Return(existingSettings, nil)
	mocks.settings.On("UpsertArbiterSettings", mock.Anything, mock.MatchedBy(func(s *models.ArbiterSettings) bool {
		// Verify profile fields were applied
		return s.ReactionStructure == "tatara" &&
			s.ReactionRig == "t2" &&
			s.FinalStructure == "sotiyo" &&
			s.FinalFacilityTax == 1.0 &&
			// Verify existing non-structure fields are preserved
			s.UseWhitelist == true &&
			s.UseBlacklist == false
	})).Return(nil)

	req := httptest.NewRequest("POST", "/v1/arbiter/manufacturing-profiles/7/apply", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "7"}}

	result, httpErr := c.ApplyManufacturingProfile(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	settings := result.(*models.ArbiterSettings)
	assert.Equal(t, "tatara", settings.ReactionStructure)
	assert.Equal(t, "sotiyo", settings.FinalStructure)
	assert.True(t, settings.UseWhitelist)
	mocks.settings.AssertExpectations(t)
	mocks.mfgProfile.AssertExpectations(t)
}

func Test_ApplyManufacturingProfile_Returns400_WhenInvalidID(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(200)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)

	req := httptest.NewRequest("POST", "/v1/arbiter/manufacturing-profiles/xyz/apply", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "xyz"}}

	result, httpErr := c.ApplyManufacturingProfile(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}
