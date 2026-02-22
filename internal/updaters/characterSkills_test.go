package updaters_test

import (
	"context"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/stretchr/testify/assert"
)

// --- Mocks ---

type mockSkillsRepo struct {
	upsertedSkills map[int64][]*models.CharacterSkill // keyed by characterID
	upsertErr      error
}

func (m *mockSkillsRepo) UpsertSkills(ctx context.Context, characterID, userID int64, skills []*models.CharacterSkill) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	if m.upsertedSkills == nil {
		m.upsertedSkills = map[int64][]*models.CharacterSkill{}
	}
	m.upsertedSkills[characterID] = skills
	return nil
}

type mockSkillsEsiClient struct {
	skillsByChar   map[int64]*client.EsiSkillsResponse
	skillsErr      error
	refreshedToken *client.RefreshedToken
	refreshErr     error
}

func (m *mockSkillsEsiClient) GetCharacterSkills(ctx context.Context, characterID int64, token string) (*client.EsiSkillsResponse, error) {
	if m.skillsErr != nil {
		return nil, m.skillsErr
	}
	return m.skillsByChar[characterID], nil
}

func (m *mockSkillsEsiClient) RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error) {
	if m.refreshErr != nil {
		return nil, m.refreshErr
	}
	return m.refreshedToken, nil
}

type mockSkillsUserRepo struct {
	userIDs []int64
	err     error
}

func (m *mockSkillsUserRepo) GetAllIDs(ctx context.Context) ([]int64, error) {
	return m.userIDs, m.err
}

type mockSkillsCharRepo struct {
	charactersByUser map[int64][]*repositories.Character
	getErr           error
	tokenUpdateErr   error
}

func (m *mockSkillsCharRepo) GetAll(ctx context.Context, baseUserID int64) ([]*repositories.Character, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.charactersByUser[baseUserID], nil
}

func (m *mockSkillsCharRepo) UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error {
	return m.tokenUpdateErr
}

// --- Tests ---

func Test_CharacterSkillsUpdater_UpdatesSkills(t *testing.T) {
	skillsRepo := &mockSkillsRepo{}
	esiClient := &mockSkillsEsiClient{
		skillsByChar: map[int64]*client.EsiSkillsResponse{
			1001: {
				Skills: []client.EsiSkillEntry{
					{SkillID: 3380, TrainedSkillLevel: 5, ActiveSkillLevel: 5, SkillpointsInSkill: 256000},
					{SkillID: 3388, TrainedSkillLevel: 4, ActiveSkillLevel: 4, SkillpointsInSkill: 45255},
				},
				TotalSP: 5000000,
			},
		},
	}
	userRepo := &mockSkillsUserRepo{userIDs: []int64{100}}
	charRepo := &mockSkillsCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {
				{ID: 1001, Name: "Test Char", UserID: 100, EsiToken: "valid-token", EsiScopes: "esi-skills.read_skills.v1", EsiTokenExpiresOn: time.Now().Add(time.Hour)},
			},
		},
	}

	updater := updaters.NewCharacterSkillsUpdater(userRepo, charRepo, skillsRepo, esiClient)
	err := updater.UpdateAllUsers(context.Background())
	assert.NoError(t, err)

	assert.Len(t, skillsRepo.upsertedSkills[1001], 2)
	assert.Equal(t, int64(3380), skillsRepo.upsertedSkills[1001][0].SkillID)
	assert.Equal(t, 5, skillsRepo.upsertedSkills[1001][0].TrainedLevel)
}

func Test_CharacterSkillsUpdater_SkipsWithoutScope(t *testing.T) {
	skillsRepo := &mockSkillsRepo{}
	esiClient := &mockSkillsEsiClient{
		skillsByChar: map[int64]*client.EsiSkillsResponse{},
	}
	userRepo := &mockSkillsUserRepo{userIDs: []int64{100}}
	charRepo := &mockSkillsCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {
				{ID: 1001, Name: "No Scope Char", UserID: 100, EsiToken: "valid-token", EsiScopes: "esi-assets.read_assets.v1", EsiTokenExpiresOn: time.Now().Add(time.Hour)},
			},
		},
	}

	updater := updaters.NewCharacterSkillsUpdater(userRepo, charRepo, skillsRepo, esiClient)
	err := updater.UpdateAllUsers(context.Background())
	assert.NoError(t, err)

	// No skills should have been upserted
	assert.Empty(t, skillsRepo.upsertedSkills)
}

func Test_CharacterSkillsUpdater_RefreshesExpiredToken(t *testing.T) {
	skillsRepo := &mockSkillsRepo{}
	esiClient := &mockSkillsEsiClient{
		skillsByChar: map[int64]*client.EsiSkillsResponse{
			1001: {Skills: []client.EsiSkillEntry{{SkillID: 3380, TrainedSkillLevel: 3, ActiveSkillLevel: 3, SkillpointsInSkill: 16000}}},
		},
		refreshedToken: &client.RefreshedToken{
			AccessToken:  "new-token",
			RefreshToken: "new-refresh",
			Expiry:       time.Now().Add(20 * time.Minute),
		},
	}
	userRepo := &mockSkillsUserRepo{userIDs: []int64{100}}
	charRepo := &mockSkillsCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {
				{ID: 1001, Name: "Expired Token Char", UserID: 100, EsiToken: "expired-token", EsiRefreshToken: "refresh-token", EsiScopes: "esi-skills.read_skills.v1", EsiTokenExpiresOn: time.Now().Add(-time.Hour)},
			},
		},
	}

	updater := updaters.NewCharacterSkillsUpdater(userRepo, charRepo, skillsRepo, esiClient)
	err := updater.UpdateAllUsers(context.Background())
	assert.NoError(t, err)

	assert.Len(t, skillsRepo.upsertedSkills[1001], 1)
}

func Test_CharacterSkillsUpdater_HandlesEsiError(t *testing.T) {
	skillsRepo := &mockSkillsRepo{}
	esiClient := &mockSkillsEsiClient{
		skillsErr: assert.AnError,
	}
	userRepo := &mockSkillsUserRepo{userIDs: []int64{100}}
	charRepo := &mockSkillsCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {
				{ID: 1001, Name: "Error Char", UserID: 100, EsiToken: "valid-token", EsiScopes: "esi-skills.read_skills.v1", EsiTokenExpiresOn: time.Now().Add(time.Hour)},
			},
		},
	}

	updater := updaters.NewCharacterSkillsUpdater(userRepo, charRepo, skillsRepo, esiClient)
	// Should not return error (logs it instead, continues to next user)
	err := updater.UpdateAllUsers(context.Background())
	assert.NoError(t, err)

	// No skills should have been upserted
	assert.Empty(t, skillsRepo.upsertedSkills)
}
