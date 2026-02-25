package calculator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// CalculateManufacturingSlots
// ---------------------------------------------------------------------------

func TestCalculateManufacturingSlots(t *testing.T) {
	tests := []struct {
		name     string
		skills   map[int64]int
		expected int
	}{
		{
			name:     "no skills trained — base slot only",
			skills:   map[int64]int{},
			expected: 1,
		},
		{
			name: "Mass Production 5 only",
			skills: map[int64]int{
				SkillMassProduction: 5,
			},
			expected: 6, // 1 + 5
		},
		{
			name: "Advanced Mass Production 5 only",
			skills: map[int64]int{
				SkillAdvMassProduction: 5,
			},
			expected: 6, // 1 + 5
		},
		{
			name: "Mass Production 5 and Advanced Mass Production 5",
			skills: map[int64]int{
				SkillMassProduction:    5,
				SkillAdvMassProduction: 5,
			},
			expected: 11, // 1 + 5 + 5
		},
		{
			name: "partial skills — Mass Production 3, Advanced Mass Production 2",
			skills: map[int64]int{
				SkillMassProduction:    3,
				SkillAdvMassProduction: 2,
			},
			expected: 6, // 1 + 3 + 2
		},
		{
			name: "unrelated skills have no effect",
			skills: map[int64]int{
				SkillIndustry:  5,
				SkillReactions: 5,
			},
			expected: 1, // only base slot
		},
		{
			name: "Mass Production 1 — minimum trained slot gain",
			skills: map[int64]int{
				SkillMassProduction: 1,
			},
			expected: 2, // 1 + 1
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := CalculateManufacturingSlots(tc.skills)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// ---------------------------------------------------------------------------
// CalculateReactionSlots
// ---------------------------------------------------------------------------

func TestCalculateReactionSlots(t *testing.T) {
	tests := []struct {
		name     string
		skills   map[int64]int
		expected int
	}{
		{
			name:     "no skills trained — no reaction slots",
			skills:   map[int64]int{},
			expected: 0,
		},
		{
			name: "Reactions 0 explicitly — no reaction slots",
			skills: map[int64]int{
				SkillReactions: 0,
			},
			expected: 0,
		},
		{
			name: "Reactions 1 — base slot unlocked",
			skills: map[int64]int{
				SkillReactions: 1,
			},
			expected: 1,
		},
		{
			name: "Reactions 5 — base slot only (no Mass Reactions)",
			skills: map[int64]int{
				SkillReactions: 5,
			},
			expected: 1,
		},
		{
			name: "Reactions 5, Mass Reactions 5",
			skills: map[int64]int{
				SkillReactions:     5,
				SkillMassReactions: 5,
			},
			expected: 6, // 1 + 5
		},
		{
			name: "Reactions 5, Mass Reactions 5, Advanced Mass Reactions 5",
			skills: map[int64]int{
				SkillReactions:      5,
				SkillMassReactions:  5,
				SkillAdvMassReactions: 5,
			},
			expected: 11, // 1 + 5 + 5
		},
		{
			name: "partial skills — Reactions 3, Mass Reactions 2, Advanced Mass Reactions 1",
			skills: map[int64]int{
				SkillReactions:      3,
				SkillMassReactions:  2,
				SkillAdvMassReactions: 1,
			},
			expected: 4, // 1 + 2 + 1
		},
		{
			name: "Mass Reactions trained but Reactions not trained — no slots",
			skills: map[int64]int{
				SkillMassReactions:  5,
				SkillAdvMassReactions: 5,
			},
			expected: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := CalculateReactionSlots(tc.skills)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// ---------------------------------------------------------------------------
// MfgSlotsAvailable / ReactSlotsAvailable
// ---------------------------------------------------------------------------

func TestMfgSlotsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		capacity *CharacterCapacity
		expected int
	}{
		{
			name:     "all slots free",
			capacity: &CharacterCapacity{MfgSlotsMax: 5, MfgSlotsUsed: 0},
			expected: 5,
		},
		{
			name:     "some slots used",
			capacity: &CharacterCapacity{MfgSlotsMax: 5, MfgSlotsUsed: 3},
			expected: 2,
		},
		{
			name:     "all slots used",
			capacity: &CharacterCapacity{MfgSlotsMax: 5, MfgSlotsUsed: 5},
			expected: 0,
		},
		{
			name:     "over-capacity clamps to zero",
			capacity: &CharacterCapacity{MfgSlotsMax: 5, MfgSlotsUsed: 7},
			expected: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, MfgSlotsAvailable(tc.capacity))
		})
	}
}

func TestReactSlotsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		capacity *CharacterCapacity
		expected int
	}{
		{
			name:     "all slots free",
			capacity: &CharacterCapacity{ReactSlotsMax: 11, ReactSlotsUsed: 0},
			expected: 11,
		},
		{
			name:     "some slots used",
			capacity: &CharacterCapacity{ReactSlotsMax: 11, ReactSlotsUsed: 6},
			expected: 5,
		},
		{
			name:     "all slots used",
			capacity: &CharacterCapacity{ReactSlotsMax: 11, ReactSlotsUsed: 11},
			expected: 0,
		},
		{
			name:     "over-capacity clamps to zero",
			capacity: &CharacterCapacity{ReactSlotsMax: 3, ReactSlotsUsed: 5},
			expected: 0,
		},
		{
			name:     "no reaction slots at all",
			capacity: &CharacterCapacity{ReactSlotsMax: 0, ReactSlotsUsed: 0},
			expected: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ReactSlotsAvailable(tc.capacity))
		})
	}
}

// ---------------------------------------------------------------------------
// BuildCharacterCapacities
// ---------------------------------------------------------------------------

func TestBuildCharacterCapacities_BasicMfg(t *testing.T) {
	names := map[int64]string{
		1001: "Alice",
	}
	skills := map[int64]map[int64]int{
		1001: {
			SkillIndustry:       5,
			SkillAdvIndustry:    5,
			SkillMassProduction: 5,
		},
	}
	usage := map[int64]map[string]int{
		1001: {"manufacturing": 3},
	}

	result := BuildCharacterCapacities(names, skills, usage)

	require.Len(t, result, 1)
	c := result[0]
	assert.Equal(t, int64(1001), c.CharacterID)
	assert.Equal(t, "Alice", c.CharacterName)
	assert.Equal(t, 6, c.MfgSlotsMax)   // 1 + 5
	assert.Equal(t, 3, c.MfgSlotsUsed)
	assert.Equal(t, 0, c.ReactSlotsMax) // no reactions skill
	assert.Equal(t, 0, c.ReactSlotsUsed)
	assert.Equal(t, 5, c.IndustrySkill)
	assert.Equal(t, 5, c.AdvIndustrySkill)
	assert.Equal(t, 0, c.ReactionsSkill)
}

func TestBuildCharacterCapacities_BasicReaction(t *testing.T) {
	names := map[int64]string{
		1002: "Bob",
	}
	skills := map[int64]map[int64]int{
		1002: {
			SkillReactions:     5,
			SkillMassReactions: 5,
		},
	}
	usage := map[int64]map[string]int{
		1002: {"reaction": 2},
	}

	result := BuildCharacterCapacities(names, skills, usage)

	require.Len(t, result, 1)
	c := result[0]
	assert.Equal(t, int64(1002), c.CharacterID)
	assert.Equal(t, "Bob", c.CharacterName)
	assert.Equal(t, 1, c.MfgSlotsMax)   // base only — no Industry skill
	assert.Equal(t, 0, c.MfgSlotsUsed)
	assert.Equal(t, 6, c.ReactSlotsMax) // 1 + 5
	assert.Equal(t, 2, c.ReactSlotsUsed)
	assert.Equal(t, 0, c.IndustrySkill)
	assert.Equal(t, 5, c.ReactionsSkill)
}

func TestBuildCharacterCapacities_ExcludesUntrained(t *testing.T) {
	names := map[int64]string{
		1001: "Alice",   // has industry skills
		9999: "Nobody",  // no relevant skills
	}
	skills := map[int64]map[int64]int{
		1001: {
			SkillIndustry: 3,
		},
		9999: {
			// only unrelated skills
			SkillMassProduction: 5, // has slots but no Industry skill — excluded
		},
	}

	result := BuildCharacterCapacities(names, skills, nil)

	require.Len(t, result, 1)
	assert.Equal(t, int64(1001), result[0].CharacterID)
}

func TestBuildCharacterCapacities_ExcludesNoSkills(t *testing.T) {
	names := map[int64]string{
		1001: "Alice",
		1002: "Bob",
	}
	// No skills at all for any character.
	skills := map[int64]map[int64]int{}

	result := BuildCharacterCapacities(names, skills, nil)

	assert.Len(t, result, 0)
}

func TestBuildCharacterCapacities_NilSkillMap(t *testing.T) {
	names := map[int64]string{
		1001: "Alice",
	}
	// skillsByCharacter is nil for Alice — should not panic.
	skills := map[int64]map[int64]int{}

	result := BuildCharacterCapacities(names, skills, nil)

	// Alice has no skills — excluded.
	assert.Len(t, result, 0)
}

func TestBuildCharacterCapacities_NilUsageMap(t *testing.T) {
	names := map[int64]string{
		1001: "Alice",
	}
	skills := map[int64]map[int64]int{
		1001: {SkillIndustry: 1},
	}

	// slotUsage is nil — should default to 0 usage.
	result := BuildCharacterCapacities(names, skills, nil)

	require.Len(t, result, 1)
	assert.Equal(t, 0, result[0].MfgSlotsUsed)
	assert.Equal(t, 0, result[0].ReactSlotsUsed)
}

func TestBuildCharacterCapacities_SortOrder(t *testing.T) {
	// Three manufacturing characters with different skill totals,
	// and one reaction-only character.
	names := map[int64]string{
		1001: "LowSkill",     // Industry 1
		1002: "HighSkill",    // Industry 5, AdvIndustry 5
		1003: "MidSkill",     // Industry 5, AdvIndustry 3
		1004: "ReactOnly",    // Reactions 5 only, no Industry
	}
	skills := map[int64]map[int64]int{
		1001: {SkillIndustry: 1, SkillAdvIndustry: 0},
		1002: {SkillIndustry: 5, SkillAdvIndustry: 5},
		1003: {SkillIndustry: 5, SkillAdvIndustry: 3},
		1004: {SkillReactions: 5},
	}

	result := BuildCharacterCapacities(names, skills, nil)

	require.Len(t, result, 4)

	// Manufacturing characters first, highest combined mfg skill at top.
	assert.Equal(t, int64(1002), result[0].CharacterID, "HighSkill (5+5=10) should be first")
	assert.Equal(t, int64(1003), result[1].CharacterID, "MidSkill (5+3=8) should be second")
	assert.Equal(t, int64(1001), result[2].CharacterID, "LowSkill (1+0=1) should be third")
	// Reaction-only character last.
	assert.Equal(t, int64(1004), result[3].CharacterID, "ReactOnly should be last")
}

func TestBuildCharacterCapacities_SortTieBreakByReactionsSkill(t *testing.T) {
	// Two manufacturing characters with equal mfg scores — tie-broken by ReactionsSkill.
	names := map[int64]string{
		2001: "NoReact", // Industry 5, AdvIndustry 0, Reactions 0
		2002: "WithReact", // Industry 5, AdvIndustry 0, Reactions 5
	}
	skills := map[int64]map[int64]int{
		2001: {SkillIndustry: 5, SkillAdvIndustry: 0},
		2002: {SkillIndustry: 5, SkillAdvIndustry: 0, SkillReactions: 5},
	}

	result := BuildCharacterCapacities(names, skills, nil)

	require.Len(t, result, 2)
	// WithReact (ReactionsSkill 5) should sort before NoReact (ReactionsSkill 0).
	assert.Equal(t, int64(2002), result[0].CharacterID)
	assert.Equal(t, int64(2001), result[1].CharacterID)
}

func TestBuildCharacterCapacities_MixedWithSlotUsage(t *testing.T) {
	names := map[int64]string{
		3001: "Char1",
		3002: "Char2",
	}
	skills := map[int64]map[int64]int{
		3001: {
			SkillIndustry:         5,
			SkillAdvIndustry:      5,
			SkillMassProduction:   5,
			SkillAdvMassProduction: 5,
			SkillReactions:        5,
			SkillMassReactions:    5,
			SkillAdvMassReactions: 5,
		},
		3002: {
			SkillIndustry:   3,
			SkillReactions:  2,
		},
	}
	usage := map[int64]map[string]int{
		3001: {"manufacturing": 6, "reaction": 4},
		3002: {"manufacturing": 1, "reaction": 1},
	}

	result := BuildCharacterCapacities(names, skills, usage)

	require.Len(t, result, 2)

	// Char1 has higher skill totals — comes first.
	c1 := result[0]
	assert.Equal(t, int64(3001), c1.CharacterID)
	assert.Equal(t, 11, c1.MfgSlotsMax)   // 1 + 5 + 5
	assert.Equal(t, 6, c1.MfgSlotsUsed)
	assert.Equal(t, 11, c1.ReactSlotsMax) // 1 + 5 + 5
	assert.Equal(t, 4, c1.ReactSlotsUsed)
	assert.Equal(t, 5, MfgSlotsAvailable(c1))
	assert.Equal(t, 7, ReactSlotsAvailable(c1))

	c2 := result[1]
	assert.Equal(t, int64(3002), c2.CharacterID)
	assert.Equal(t, 1, c2.MfgSlotsMax)  // 1 + 0 + 0
	assert.Equal(t, 1, c2.MfgSlotsUsed)
	assert.Equal(t, 1, c2.ReactSlotsMax) // 1 + 0 + 0
	assert.Equal(t, 1, c2.ReactSlotsUsed)
	assert.Equal(t, 0, MfgSlotsAvailable(c2))
	assert.Equal(t, 0, ReactSlotsAvailable(c2))
}

func TestBuildCharacterCapacities_EmptyInputs(t *testing.T) {
	result := BuildCharacterCapacities(
		map[int64]string{},
		map[int64]map[int64]int{},
		map[int64]map[string]int{},
	)

	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestBuildCharacterCapacities_ReturnsEmptySliceNotNil(t *testing.T) {
	// Ensure the function returns [] not nil for JSON marshaling safety.
	result := BuildCharacterCapacities(
		map[int64]string{},
		map[int64]map[int64]int{},
		nil,
	)

	assert.NotNil(t, result)
}
