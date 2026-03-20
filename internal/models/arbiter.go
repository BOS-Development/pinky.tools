package models

import "time"

// BlueprintMaterial is an input material for an arbiter BOM calculation.
type BlueprintMaterial struct {
	TypeID   int64
	TypeName string
	Quantity int
}

// BlueprintProduct is a product from a blueprint activity.
type BlueprintProduct struct {
	TypeID      int64
	TypeName    string
	Quantity    int
	Probability float64
}

// ArbiterSettings holds per-user configuration for the Arbiter T2 profit advisor.
// Each of the four structure slots drives material efficiency, time efficiency,
// and job cost calculations for that step of the production chain.
type ArbiterSettings struct {
	UserID int64

	// Reaction structure (composite reactions, moon goo reactions)
	ReactionStructure string
	ReactionRig       string
	ReactionSecurity  string
	ReactionSystemID  *int64

	// Invention structure (T2 BPC invention)
	InventionStructure string
	InventionRig       string
	InventionSecurity  string
	InventionSystemID  *int64

	// Component build structure (T2 components)
	ComponentStructure string
	ComponentRig       string
	ComponentSecurity  string
	ComponentSystemID  *int64

	// Final build structure (T2 ships and modules)
	FinalStructure string
	FinalRig       string
	FinalSecurity  string
	FinalSystemID  *int64
}

// Decryptor represents a T2 invention decryptor from sde_decryptors.
type Decryptor struct {
	TypeID                int64
	Name                  string
	ProbabilityMultiplier float64
	MEModifier            int
	TEModifier            int
	RunModifier           int
}

// T2BlueprintScanItem represents a single T2 item eligible for Arbiter scanning.
type T2BlueprintScanItem struct {
	ProductTypeID       int64
	ProductName         string
	BlueprintTypeID     int64   // T2 blueprint
	T1BlueprintTypeID   int64   // T1 blueprint used for invention
	BaseInventionChance float64 // from sde_blueprint_products.probability
	BaseResultME        int     // base BPC ME produced by invention
	BaseResultRuns      int     // base run count produced by invention
	Category            string  // "ship" or "module"
}

// InventionCharacter holds a character's relevant skills for invention chance calculation.
type InventionCharacter struct {
	CharacterID          int64
	Name                 string
	EncryptionSkillLevel int
	Science1SkillLevel   int
	Science2SkillLevel   int
}

// DecryptorOption represents one invention scenario (a specific decryptor or no decryptor).
type DecryptorOption struct {
	TypeID                *int64  // nil = no decryptor
	Name                  string
	ProbabilityMultiplier float64
	MEModifier            int
	TEModifier            int
	RunModifier           int
	ResultingME           int
	ResultingRuns         int
	InventionCost         float64 // datacores + decryptor + copy cost, per successful BPC
	MaterialCost          float64 // full BOM cost with this ME level
	JobCost               float64
	TotalCost             float64
	Profit                float64
	ROI                   float64
	ISKPerDay             float64
	BuildTimeSec          int64
}

// ArbiterOpportunity represents a single T2 item with all its invention/build options.
type ArbiterOpportunity struct {
	ProductTypeID int64
	ProductName   string
	Category      string            // "ship" or "module"
	JitaSellPrice float64
	JitaBuyPrice  float64
	BestDecryptor *DecryptorOption
	AllDecryptors []*DecryptorOption // all 8 decryptors + no-decryptor option
}

// ArbiterScanResult is the full result of a scan run.
type ArbiterScanResult struct {
	Opportunities     []*ArbiterOpportunity
	GeneratedAt       time.Time
	TotalScanned      int
	BestCharacterID   int64
	BestCharacterName string
}
