package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type CharacterSkills struct {
	db *sql.DB
}

func NewCharacterSkills(db *sql.DB) *CharacterSkills {
	return &CharacterSkills{db: db}
}

func (r *CharacterSkills) UpsertSkills(ctx context.Context, characterID, userID int64, skills []*models.CharacterSkill) error {
	if len(skills) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction for character skills upsert")
	}
	defer tx.Rollback()

	upsertQuery := `
		INSERT INTO character_skills
			(character_id, user_id, skill_id, trained_level, active_level, skillpoints, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		ON CONFLICT (character_id, skill_id)
		DO UPDATE SET
			user_id = EXCLUDED.user_id,
			trained_level = EXCLUDED.trained_level,
			active_level = EXCLUDED.active_level,
			skillpoints = EXCLUDED.skillpoints,
			updated_at = now()
	`

	stmt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return errors.Wrap(err, "failed to prepare character skill upsert")
	}

	for _, skill := range skills {
		_, err = stmt.ExecContext(ctx,
			characterID,
			userID,
			skill.SkillID,
			skill.TrainedLevel,
			skill.ActiveLevel,
			skill.Skillpoints,
		)
		if err != nil {
			return errors.Wrap(err, "failed to execute character skill upsert")
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit character skills transaction")
	}
	return nil
}

func (r *CharacterSkills) GetSkills(ctx context.Context, characterID int64) ([]*models.CharacterSkill, error) {
	query := `
		SELECT character_id, user_id, skill_id, trained_level, active_level, skillpoints, updated_at
		FROM character_skills
		WHERE character_id = $1
		ORDER BY skill_id
	`

	rows, err := r.db.QueryContext(ctx, query, characterID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query character skills")
	}
	defer rows.Close()

	skills := []*models.CharacterSkill{}
	for rows.Next() {
		var skill models.CharacterSkill
		err = rows.Scan(
			&skill.CharacterID,
			&skill.UserID,
			&skill.SkillID,
			&skill.TrainedLevel,
			&skill.ActiveLevel,
			&skill.Skillpoints,
			&skill.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan character skill")
		}
		skills = append(skills, &skill)
	}

	return skills, nil
}

// GetIndustrySkills returns a map of skill_type_id → active_level for industry-relevant skills
// for a given character. This is used by the manufacturing calculator.
func (r *CharacterSkills) GetIndustrySkills(ctx context.Context, characterID int64, skillTypeIDs []int64) (map[int64]int, error) {
	if len(skillTypeIDs) == 0 {
		return map[int64]int{}, nil
	}

	query := `
		SELECT skill_id, active_level
		FROM character_skills
		WHERE character_id = $1
		  AND skill_id = ANY($2::bigint[])
	`

	rows, err := r.db.QueryContext(ctx, query, characterID, pq.Array(skillTypeIDs))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query industry skills")
	}
	defer rows.Close()

	result := map[int64]int{}
	for rows.Next() {
		var skillID int64
		var level int
		err = rows.Scan(&skillID, &level)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan industry skill")
		}
		result[skillID] = level
	}

	return result, nil
}

// GetSkillsForUser returns all skills for all characters belonging to a user.
func (r *CharacterSkills) GetSkillsForUser(ctx context.Context, userID int64) ([]*models.CharacterSkill, error) {
	query := `
		SELECT character_id, user_id, skill_id, trained_level, active_level, skillpoints, updated_at
		FROM character_skills
		WHERE user_id = $1
		ORDER BY character_id, skill_id
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query user skills")
	}
	defer rows.Close()

	skills := []*models.CharacterSkill{}
	for rows.Next() {
		var skill models.CharacterSkill
		err = rows.Scan(
			&skill.CharacterID,
			&skill.UserID,
			&skill.SkillID,
			&skill.TrainedLevel,
			&skill.ActiveLevel,
			&skill.Skillpoints,
			&skill.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan user skill")
		}
		skills = append(skills, &skill)
	}

	return skills, nil
}
