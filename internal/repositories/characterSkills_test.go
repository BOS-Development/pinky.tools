package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_CharacterSkillsShouldUpsertAndGet(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	skillsRepo := repositories.NewCharacterSkills(db)

	user := &repositories.User{ID: 5000, Name: "Skills Test User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 50001, Name: "Skills Char", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	skills := []*models.CharacterSkill{
		{SkillID: 3380, TrainedLevel: 5, ActiveLevel: 5, Skillpoints: 256000},
		{SkillID: 3388, TrainedLevel: 4, ActiveLevel: 4, Skillpoints: 45255},
		{SkillID: 22242, TrainedLevel: 3, ActiveLevel: 3, Skillpoints: 16000},
	}

	err = skillsRepo.UpsertSkills(context.Background(), char.ID, user.ID, skills)
	assert.NoError(t, err)

	result, err := skillsRepo.GetSkills(context.Background(), char.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, int64(3380), result[0].SkillID)
	assert.Equal(t, 5, result[0].TrainedLevel)
	assert.Equal(t, 5, result[0].ActiveLevel)
	assert.Equal(t, int64(256000), result[0].Skillpoints)
	assert.Equal(t, char.ID, result[0].CharacterID)
	assert.Equal(t, user.ID, result[0].UserID)
}

func Test_CharacterSkillsShouldUpsertExistingSkills(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	skillsRepo := repositories.NewCharacterSkills(db)

	user := &repositories.User{ID: 5010, Name: "Skills Upsert User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 50101, Name: "Upsert Char", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	// Insert initial skills
	skills := []*models.CharacterSkill{
		{SkillID: 3380, TrainedLevel: 4, ActiveLevel: 4, Skillpoints: 45000},
	}
	err = skillsRepo.UpsertSkills(context.Background(), char.ID, user.ID, skills)
	assert.NoError(t, err)

	// Upsert with updated levels
	updatedSkills := []*models.CharacterSkill{
		{SkillID: 3380, TrainedLevel: 5, ActiveLevel: 5, Skillpoints: 256000},
	}
	err = skillsRepo.UpsertSkills(context.Background(), char.ID, user.ID, updatedSkills)
	assert.NoError(t, err)

	result, err := skillsRepo.GetSkills(context.Background(), char.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 5, result[0].TrainedLevel)
	assert.Equal(t, int64(256000), result[0].Skillpoints)
}

func Test_CharacterSkillsShouldGetIndustrySkills(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	skillsRepo := repositories.NewCharacterSkills(db)

	user := &repositories.User{ID: 5020, Name: "Industry Skills User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 50201, Name: "Industry Char", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	skills := []*models.CharacterSkill{
		{SkillID: 3380, TrainedLevel: 5, ActiveLevel: 5, Skillpoints: 256000},  // Industry
		{SkillID: 3388, TrainedLevel: 4, ActiveLevel: 4, Skillpoints: 45255},  // Advanced Industry
		{SkillID: 22242, TrainedLevel: 3, ActiveLevel: 3, Skillpoints: 16000}, // Some other skill
		{SkillID: 11529, TrainedLevel: 2, ActiveLevel: 2, Skillpoints: 5000},  // Reactions
	}

	err = skillsRepo.UpsertSkills(context.Background(), char.ID, user.ID, skills)
	assert.NoError(t, err)

	// Query only Industry and Reactions skill IDs
	industrySkills, err := skillsRepo.GetIndustrySkills(context.Background(), char.ID, []int64{3380, 11529})
	assert.NoError(t, err)
	assert.Len(t, industrySkills, 2)
	assert.Equal(t, 5, industrySkills[3380])
	assert.Equal(t, 2, industrySkills[11529])
}

func Test_CharacterSkillsShouldReturnEmptyForNoSkills(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	skillsRepo := repositories.NewCharacterSkills(db)

	result, err := skillsRepo.GetSkills(context.Background(), 99999)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func Test_CharacterSkillsShouldHandleEmptyUpsert(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	skillsRepo := repositories.NewCharacterSkills(db)

	err = skillsRepo.UpsertSkills(context.Background(), 99999, 99999, []*models.CharacterSkill{})
	assert.NoError(t, err)
}

func Test_CharacterSkillsShouldGetEmptyIndustrySkills(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	skillsRepo := repositories.NewCharacterSkills(db)

	result, err := skillsRepo.GetIndustrySkills(context.Background(), 99999, []int64{})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func Test_CharacterSkillsShouldGetSkillsForUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	skillsRepo := repositories.NewCharacterSkills(db)

	user := &repositories.User{ID: 5030, Name: "Multi-Char User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char1 := &repositories.Character{ID: 50301, Name: "Char 1", UserID: user.ID}
	char2 := &repositories.Character{ID: 50302, Name: "Char 2", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char1)
	assert.NoError(t, err)
	err = characterRepo.Add(context.Background(), char2)
	assert.NoError(t, err)

	skills1 := []*models.CharacterSkill{
		{SkillID: 3380, TrainedLevel: 5, ActiveLevel: 5, Skillpoints: 256000},
	}
	skills2 := []*models.CharacterSkill{
		{SkillID: 3380, TrainedLevel: 3, ActiveLevel: 3, Skillpoints: 16000},
	}

	err = skillsRepo.UpsertSkills(context.Background(), char1.ID, user.ID, skills1)
	assert.NoError(t, err)
	err = skillsRepo.UpsertSkills(context.Background(), char2.ID, user.ID, skills2)
	assert.NoError(t, err)

	result, err := skillsRepo.GetSkillsForUser(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	// Ordered by character_id, skill_id
	assert.Equal(t, char1.ID, result[0].CharacterID)
	assert.Equal(t, 5, result[0].ActiveLevel)
	assert.Equal(t, char2.ID, result[1].CharacterID)
	assert.Equal(t, 3, result[1].ActiveLevel)
}
