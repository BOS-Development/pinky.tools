package updaters

import (
	"context"
	"strings"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"

	"github.com/pkg/errors"
)

type SkillsCharacterSkillsRepository interface {
	UpsertSkills(ctx context.Context, characterID, userID int64, skills []*models.CharacterSkill) error
}

type SkillsEsiClient interface {
	GetCharacterSkills(ctx context.Context, characterID int64, token string) (*client.EsiSkillsResponse, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error)
}

type SkillsUserRepository interface {
	GetAllIDs(ctx context.Context) ([]int64, error)
}

type SkillsCharacterRepository interface {
	GetAll(ctx context.Context, baseUserID int64) ([]*repositories.Character, error)
	UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error
}

type CharacterSkillsUpdater struct {
	userRepo      SkillsUserRepository
	characterRepo SkillsCharacterRepository
	skillsRepo    SkillsCharacterSkillsRepository
	esiClient     SkillsEsiClient
}

func NewCharacterSkillsUpdater(
	userRepo SkillsUserRepository,
	characterRepo SkillsCharacterRepository,
	skillsRepo SkillsCharacterSkillsRepository,
	esiClient SkillsEsiClient,
) *CharacterSkillsUpdater {
	return &CharacterSkillsUpdater{
		userRepo:      userRepo,
		characterRepo: characterRepo,
		skillsRepo:    skillsRepo,
		esiClient:     esiClient,
	}
}

func (u *CharacterSkillsUpdater) UpdateAllUsers(ctx context.Context) error {
	userIDs, err := u.userRepo.GetAllIDs(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get user IDs for skills update")
	}

	for _, userID := range userIDs {
		if err := u.UpdateUserSkills(ctx, userID); err != nil {
			log.Error("failed to update skills for user", "userID", userID, "error", err)
		}
	}

	return nil
}

func (u *CharacterSkillsUpdater) UpdateUserSkills(ctx context.Context, userID int64) error {
	characters, err := u.characterRepo.GetAll(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get user characters for skills update")
	}

	for _, char := range characters {
		if !strings.Contains(char.EsiScopes, "esi-skills.read_skills.v1") {
			continue
		}

		if err := u.updateCharacterSkills(ctx, char, userID); err != nil {
			log.Error("failed to update skills for character", "characterID", char.ID, "error", err)
		}
	}

	return nil
}

func (u *CharacterSkillsUpdater) updateCharacterSkills(ctx context.Context, char *repositories.Character, userID int64) error {
	token := char.EsiToken

	if time.Now().After(char.EsiTokenExpiresOn) {
		refreshed, err := u.esiClient.RefreshAccessToken(ctx, char.EsiRefreshToken)
		if err != nil {
			return errors.Wrapf(err, "failed to refresh token for character %d", char.ID)
		}
		token = refreshed.AccessToken

		err = u.characterRepo.UpdateTokens(ctx, char.ID, char.UserID, refreshed.AccessToken, refreshed.RefreshToken, refreshed.Expiry)
		if err != nil {
			return errors.Wrapf(err, "failed to persist refreshed token for character %d", char.ID)
		}
		log.Info("refreshed ESI token for character (skills)", "characterID", char.ID)
	}

	esiSkills, err := u.esiClient.GetCharacterSkills(ctx, char.ID, token)
	if err != nil {
		return errors.Wrap(err, "failed to get character skills from ESI")
	}

	skills := []*models.CharacterSkill{}
	for _, s := range esiSkills.Skills {
		skills = append(skills, &models.CharacterSkill{
			CharacterID:  char.ID,
			UserID:       userID,
			SkillID:      s.SkillID,
			TrainedLevel: s.TrainedSkillLevel,
			ActiveLevel:  s.ActiveSkillLevel,
			Skillpoints:  s.SkillpointsInSkill,
		})
	}

	err = u.skillsRepo.UpsertSkills(ctx, char.ID, userID, skills)
	if err != nil {
		return errors.Wrap(err, "failed to upsert character skills")
	}

	log.Info("updated character skills", "characterID", char.ID, "count", len(skills))
	return nil
}
