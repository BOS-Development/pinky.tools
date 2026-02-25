package calculator

import "sort"

// Skill ID constants for industry-related skills.
const (
	SkillIndustry          int64 = 3380  // 4% mfg time reduction per level
	SkillAdvIndustry       int64 = 3388  // 3% mfg time reduction per level
	SkillMassProduction    int64 = 3387  // +1 mfg slot per level
	SkillAdvMassProduction int64 = 24625 // +1 mfg slot per level
	SkillReactions         int64 = 45746 // 4% reaction time reduction per level; enables reactions
	SkillMassReactions     int64 = 45748 // +1 reaction slot per level
	SkillAdvMassReactions  int64 = 45749 // +1 reaction slot per level
)

// IndustrySkillIDs is a convenience slice for fetching all industry-related skills at once.
var IndustrySkillIDs = []int64{
	SkillIndustry, SkillAdvIndustry, SkillMassProduction, SkillAdvMassProduction,
	SkillReactions, SkillMassReactions, SkillAdvMassReactions,
}

// CharacterCapacity holds the manufacturing and reaction slot capacity for a character.
type CharacterCapacity struct {
	CharacterID      int64
	CharacterName    string
	MfgSlotsMax      int
	MfgSlotsUsed     int
	ReactSlotsMax    int
	ReactSlotsUsed   int
	IndustrySkill    int // 0-5
	AdvIndustrySkill int // 0-5
	ReactionsSkill   int // 0-5
}

// CalculateManufacturingSlots returns the maximum number of manufacturing slots
// available to a character based on their skills.
// Base is 1 slot, +1 per level of Mass Production, +1 per level of Advanced Mass Production.
func CalculateManufacturingSlots(skills map[int64]int) int {
	return 1 + skills[SkillMassProduction] + skills[SkillAdvMassProduction]
}

// CalculateReactionSlots returns the maximum number of reaction slots available
// to a character based on their skills.
// Requires at least level 1 of Reactions to use any reaction slots.
// Base is 1 slot (when Reactions >= 1), +1 per level of Mass Reactions,
// +1 per level of Advanced Mass Reactions.
func CalculateReactionSlots(skills map[int64]int) int {
	if skills[SkillReactions] < 1 {
		return 0
	}
	return 1 + skills[SkillMassReactions] + skills[SkillAdvMassReactions]
}

// MfgSlotsAvailable returns the number of free manufacturing slots for a character.
// Never returns a negative value.
func MfgSlotsAvailable(c *CharacterCapacity) int {
	avail := c.MfgSlotsMax - c.MfgSlotsUsed
	if avail < 0 {
		return 0
	}
	return avail
}

// ReactSlotsAvailable returns the number of free reaction slots for a character.
// Never returns a negative value.
func ReactSlotsAvailable(c *CharacterCapacity) int {
	avail := c.ReactSlotsMax - c.ReactSlotsUsed
	if avail < 0 {
		return 0
	}
	return avail
}

// BuildCharacterCapacities constructs a slice of CharacterCapacity values from
// character name, skill, and slot-usage data.
//
// Parameters:
//   - characterNames: map of characterID -> display name
//   - skillsByCharacter: map of characterID -> skillID -> trained level
//   - slotUsage: map of characterID -> activity -> count
//     (activity keys are "manufacturing" and "reaction")
//
// Only characters that have at least one manufacturing or reaction slot are included
// (i.e., characters with Industry >= 1 OR Reactions >= 1).
//
// The result is sorted by skill efficiency descending:
//   - Manufacturing characters (Industry > 0) first, ordered by
//     (IndustrySkill + AdvIndustrySkill) descending.
//   - Then reaction-only characters, ordered by ReactionsSkill descending.
func BuildCharacterCapacities(
	characterNames map[int64]string,
	skillsByCharacter map[int64]map[int64]int,
	slotUsage map[int64]map[string]int,
) []*CharacterCapacity {
	capacities := []*CharacterCapacity{}

	for charID, name := range characterNames {
		skills := skillsByCharacter[charID]
		if skills == nil {
			skills = map[int64]int{}
		}

		mfgSlots := CalculateManufacturingSlots(skills)
		reactSlots := CalculateReactionSlots(skills)

		industrySkill := skills[SkillIndustry]
		advIndustrySkill := skills[SkillAdvIndustry]
		reactionsSkill := skills[SkillReactions]

		// Only include characters that have earned at least one slot in either activity.
		// Industry skill is not required for the base manufacturing slot (everyone starts
		// with 1), but we gate on Industry >= 1 to exclude characters that haven't trained
		// any industry skills at all.
		hasMfg := industrySkill >= 1
		hasReact := reactionsSkill >= 1
		if !hasMfg && !hasReact {
			continue
		}

		usage := slotUsage[charID]
		mfgUsed := 0
		reactUsed := 0
		if usage != nil {
			mfgUsed = usage["manufacturing"]
			reactUsed = usage["reaction"]
		}

		capacities = append(capacities, &CharacterCapacity{
			CharacterID:      charID,
			CharacterName:    name,
			MfgSlotsMax:      mfgSlots,
			MfgSlotsUsed:     mfgUsed,
			ReactSlotsMax:    reactSlots,
			ReactSlotsUsed:   reactUsed,
			IndustrySkill:    industrySkill,
			AdvIndustrySkill: advIndustrySkill,
			ReactionsSkill:   reactionsSkill,
		})
	}

	sort.Slice(capacities, func(i, j int) bool {
		ci := capacities[i]
		cj := capacities[j]

		iMfgScore := ci.IndustrySkill + ci.AdvIndustrySkill
		jMfgScore := cj.IndustrySkill + cj.AdvIndustrySkill

		// Manufacturing characters come before reaction-only characters.
		iHasMfg := ci.IndustrySkill >= 1
		jHasMfg := cj.IndustrySkill >= 1
		if iHasMfg != jHasMfg {
			return iHasMfg
		}

		// Among manufacturing characters, sort by combined mfg skill descending.
		if iMfgScore != jMfgScore {
			return iMfgScore > jMfgScore
		}

		// Break ties with reactions skill descending.
		return ci.ReactionsSkill > cj.ReactionsSkill
	})

	return capacities
}
