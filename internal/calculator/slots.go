package calculator

import "sort"

// Skill ID constants for industry-related skills.
const (
	SkillIndustry                int64 = 3380  // 4% mfg time reduction per level
	SkillAdvIndustry             int64 = 3388  // 3% mfg time reduction per level
	SkillMassProduction          int64 = 3387  // +1 mfg slot per level
	SkillAdvMassProduction       int64 = 24625 // +1 mfg slot per level
	SkillReactions               int64 = 45746 // 4% reaction time reduction per level; enables reactions
	SkillMassReactions           int64 = 45748 // +1 reaction slot per level
	SkillAdvMassReactions        int64 = 45749 // +1 reaction slot per level
	SkillScience                 int64 = 3402  // Enables science activities
	SkillLaboratoryOperation     int64 = 3406  // +1 science slot per level
	SkillAdvLaboratoryOperation  int64 = 24624 // +1 science slot per level
)

// IndustrySkillIDs is a convenience slice for fetching all industry-related skills at once.
var IndustrySkillIDs = []int64{
	SkillIndustry, SkillAdvIndustry, SkillMassProduction, SkillAdvMassProduction,
	SkillReactions, SkillMassReactions, SkillAdvMassReactions,
	SkillScience, SkillLaboratoryOperation, SkillAdvLaboratoryOperation,
}

// CharacterCapacity holds the manufacturing and reaction slot capacity for a character.
type CharacterCapacity struct {
	CharacterID      int64
	CharacterName    string
	MfgSlotsMax      int
	MfgSlotsUsed     int
	ReactSlotsMax    int
	ReactSlotsUsed   int
	SciSlotsMax      int
	SciSlotsUsed     int
	IndustrySkill    int // 0-5
	AdvIndustrySkill int // 0-5
	ReactionsSkill   int // 0-5
	ScienceSkill     int // 0-5
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

// CalculateScienceSlots returns the maximum number of science slots available
// to a character based on their skills.
// Requires at least level 1 of Science to use any science slots.
// Base is 1 slot (when Science >= 1), +1 per level of Laboratory Operation,
// +1 per level of Advanced Laboratory Operation.
func CalculateScienceSlots(skills map[int64]int) int {
	if skills[SkillScience] < 1 {
		return 0
	}
	return 1 + skills[SkillLaboratoryOperation] + skills[SkillAdvLaboratoryOperation]
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

// SciSlotsAvailable returns the number of free science slots for a character.
// Never returns a negative value.
func SciSlotsAvailable(c *CharacterCapacity) int {
	avail := c.SciSlotsMax - c.SciSlotsUsed
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
		sciSlots := CalculateScienceSlots(skills)

		industrySkill := skills[SkillIndustry]
		advIndustrySkill := skills[SkillAdvIndustry]
		reactionsSkill := skills[SkillReactions]
		scienceSkill := skills[SkillScience]

		// Only include characters that have earned at least one slot in any activity.
		// Industry skill is not required for the base manufacturing slot (everyone starts
		// with 1), but we gate on Industry >= 1 to exclude characters that haven't trained
		// any industry skills at all.
		hasMfg := industrySkill >= 1
		hasReact := reactionsSkill >= 1
		hasSci := scienceSkill >= 1
		if !hasMfg && !hasReact && !hasSci {
			continue
		}

		usage := slotUsage[charID]
		mfgUsed := 0
		reactUsed := 0
		sciUsed := 0
		if usage != nil {
			mfgUsed = usage["manufacturing"]
			reactUsed = usage["reaction"]
			sciUsed = usage["te_research"] + usage["me_research"] + usage["copying"] + usage["invention"]
		}

		capacities = append(capacities, &CharacterCapacity{
			CharacterID:      charID,
			CharacterName:    name,
			MfgSlotsMax:      mfgSlots,
			MfgSlotsUsed:     mfgUsed,
			ReactSlotsMax:    reactSlots,
			ReactSlotsUsed:   reactUsed,
			SciSlotsMax:      sciSlots,
			SciSlotsUsed:     sciUsed,
			IndustrySkill:    industrySkill,
			AdvIndustrySkill: advIndustrySkill,
			ReactionsSkill:   reactionsSkill,
			ScienceSkill:     scienceSkill,
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
