package models

import "time"

// BlueprintMaterial is an input material for an arbiter BOM calculation.
type BlueprintMaterial struct {
	TypeID   int64  `json:"type_id"`
	TypeName string `json:"type_name"`
	Quantity int    `json:"quantity"`
}

// BlueprintProduct is a product from a blueprint activity.
type BlueprintProduct struct {
	TypeID      int64   `json:"type_id"`
	TypeName    string  `json:"type_name"`
	Quantity    int     `json:"quantity"`
	Probability float64 `json:"probability"`
}

// ArbiterSettings holds per-user configuration for the Arbiter profit advisor.
type ArbiterSettings struct {
	UserID int64 `json:"user_id"`

	// Reaction structure (composite reactions, moon goo reactions)
	ReactionStructure string `json:"reaction_structure"`
	ReactionRig       string `json:"reaction_rig"`
	ReactionSystemID  *int64 `json:"reaction_system_id"`

	// Invention structure (T2 BPC invention)
	InventionStructure string `json:"invention_structure"`
	InventionRig       string `json:"invention_rig"`
	InventionSystemID  *int64 `json:"invention_system_id"`

	// Component build structure (T2 components)
	ComponentStructure string `json:"component_structure"`
	ComponentRig       string `json:"component_rig"`
	ComponentSystemID  *int64 `json:"component_system_id"`

	// Final build structure (T2 ships and modules)
	FinalStructure string `json:"final_structure"`
	FinalRig       string `json:"final_rig"`
	FinalSystemID  *int64 `json:"final_system_id"`

	// Feature toggles
	UseWhitelist    bool   `json:"use_whitelist"`
	UseBlacklist    bool   `json:"use_blacklist"`
	DecryptorTypeID *int64 `json:"decryptor_type_id"`
	DefaultScopeID  *int64 `json:"default_scope_id"`
}

// Decryptor represents a T2 invention decryptor from sde_decryptors.
type Decryptor struct {
	TypeID                int64   `json:"type_id"`
	Name                  string  `json:"name"`
	ProbabilityMultiplier float64 `json:"probability_multiplier"`
	MEModifier            int     `json:"me_modifier"`
	TEModifier            int     `json:"te_modifier"`
	RunModifier           int     `json:"run_modifier"`
}

// T2BlueprintScanItem represents a single T2 item eligible for Arbiter scanning.
type T2BlueprintScanItem struct {
	ProductTypeID       int64   `json:"product_type_id"`
	ProductName         string  `json:"product_name"`
	BlueprintTypeID     int64   `json:"blueprint_type_id"`    // T2 blueprint
	T1BlueprintTypeID   int64   `json:"t1_blueprint_type_id"` // T1 blueprint used for invention
	BaseInventionChance float64 `json:"base_invention_chance"` // from sde_blueprint_products.probability
	BaseResultME        int     `json:"base_result_me"`        // base BPC ME produced by invention
	BaseResultRuns      int     `json:"base_result_runs"`      // base run count produced by invention
	Category            string  `json:"category"`              // "ship" or "module"
}

// BlueprintScanItem represents a manufacturable item for Arbiter scanning (all tech levels).
type BlueprintScanItem struct {
	ProductTypeID       int64   `json:"product_type_id"`
	ProductName         string  `json:"product_name"`
	BlueprintTypeID     int64   `json:"blueprint_type_id"`
	T1BlueprintTypeID   *int64  `json:"t1_blueprint_type_id"`  // non-nil only for T2 items
	BaseInventionChance float64 `json:"base_invention_chance"` // 0 for non-T2
	BaseResultRuns      int     `json:"base_result_runs"`
	TechLevel           string  `json:"tech_level"`    // "T1", "T2", "T3", "Faction", etc.
	CategoryName        string  `json:"category_name"` // e.g. "Ship", "Module"
	GroupName           string  `json:"group_name"`
	Category            string  `json:"category"` // lowercased, e.g. "ship", "module"
}

// InventionCharacter holds a character's relevant skills for invention chance calculation.
type InventionCharacter struct {
	CharacterID          int64  `json:"character_id"`
	Name                 string `json:"name"`
	EncryptionSkillLevel int    `json:"encryption_skill_level"`
	Science1SkillLevel   int    `json:"science1_skill_level"`
	Science2SkillLevel   int    `json:"science2_skill_level"`
}

// DecryptorOption represents one invention scenario (a specific decryptor or no decryptor).
type DecryptorOption struct {
	TypeID                *int64  `json:"type_id"` // nil = no decryptor
	Name                  string  `json:"name"`
	ProbabilityMultiplier float64 `json:"probability_multiplier"`
	MEModifier            int     `json:"me_modifier"`
	TEModifier            int     `json:"te_modifier"`
	RunModifier           int     `json:"run_modifier"`
	ResultingME           int     `json:"resulting_me"`
	ResultingTE           int     `json:"resulting_te"`
	ResultingRuns         int     `json:"resulting_runs"`
	ME                    int     `json:"me"`
	TE                    int     `json:"te"`
	InventionCost         float64 `json:"invention_cost"`  // datacores + decryptor + copy cost, per successful BPC
	MaterialCost          float64 `json:"material_cost"`   // full BOM cost with this ME level
	JobCost               float64 `json:"job_cost"`
	TotalCost             float64 `json:"total_cost"`
	Profit                float64 `json:"profit"`
	ROI                   float64 `json:"roi"`
	ISKPerDay             float64 `json:"isk_per_day"`
	BuildTimeSec          int64   `json:"build_time_sec"`
}

// ArbiterScope is a named group of characters and/or corporations whose assets are pooled.
type ArbiterScope struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

// ArbiterScopeMember links a character or corporation to an ArbiterScope.
type ArbiterScopeMember struct {
	ID         int64  `json:"id"`
	ScopeID    int64  `json:"scope_id"`
	MemberType string `json:"member_type"` // "character" or "corporation"
	MemberID   int64  `json:"member_id"`
	Name       string `json:"name"` // resolved name for display
}

// ArbiterTaxProfile holds per-user pricing and fee configuration for Arbiter.
type ArbiterTaxProfile struct {
	UserID             int64   `json:"user_id"`
	TraderCharacterID  *int64  `json:"trader_character_id"`
	SalesTaxRate       float64 `json:"sales_tax_rate"`
	BrokerFeeRate      float64 `json:"broker_fee_rate"`
	StructureBrokerFee float64 `json:"structure_broker_fee"`
	InputPriceType     string  `json:"input_price_type"`  // "buy" or "sell"
	OutputPriceType    string  `json:"output_price_type"` // "buy" or "sell"
}

// ArbiterListItem represents an entry in the Arbiter blacklist or whitelist.
type ArbiterListItem struct {
	UserID  int64     `json:"user_id"`
	TypeID  int64     `json:"type_id"`
	Name    string    `json:"name"`
	AddedAt time.Time `json:"added_at"`
}

// DemandStats holds demand and order book volume for a type.
type DemandStats struct {
	TypeID          int64   `json:"type_id"`
	DemandPerDay    float64 `json:"demand_per_day"`    // 30d avg
	OrderBookVolume int64   `json:"order_book_volume"` // current sell order volume
	DaysOfSupply    float64 `json:"days_of_supply"`    // order_book_volume / demand_per_day
}

// SolarSystemSearchResult is a result from solar system name search.
type SolarSystemSearchResult struct {
	SolarSystemID int64   `json:"solar_system_id"`
	Name          string  `json:"name"`
	SecurityClass string  `json:"security_class"` // "high", "low", "null", "wormhole"
	Security      float64 `json:"security"`
}

// BOMNode represents one node in the full production tree.
type BOMNode struct {
	TypeID        int64      `json:"type_id"`
	Name          string     `json:"name"`
	Category      string     `json:"category"`
	Quantity      int64      `json:"quantity"`
	Available     int64      `json:"available"`    // from scope assets
	Needed        int64      `json:"needed"`
	Delta         int64      `json:"delta"`        // needed - available (clamped to 0)
	UnitBuyPrice  float64    `json:"unit_buy_price"`
	UnitBuildCost float64    `json:"unit_build_cost"`
	Decision      string     `json:"decision"` // "build", "buy", "buy_override", "build_override"
	Children      []*BOMNode `json:"children"` // sub-components
	IsBlacklisted bool       `json:"is_blacklisted"`
	IsWhitelisted bool       `json:"is_whitelisted"`
}

// ArbiterOpportunity represents a single manufacturable item with all its profit data.
type ArbiterOpportunity struct {
	ProductTypeID   int64              `json:"product_type_id"`
	ProductName     string             `json:"product_name"`
	Category        string             `json:"category"`
	Group           string             `json:"group"`
	TechLevel       string             `json:"tech_level"` // "T1", "T2", "T3", "Faction" etc
	JitaSellPrice   float64            `json:"jita_sell_price"`
	JitaBuyPrice    float64            `json:"jita_buy_price"`
	DemandPerDay    float64            `json:"demand_per_day"`  // 30d avg daily volume
	DaysOfSupply    float64            `json:"days_of_supply"`  // order_book_volume / demand
	Duration        int64              `json:"duration_sec"`
	Runs            int                `json:"runs"`
	ME              int                `json:"me"`
	TE              int                `json:"te"`
	MaterialCost    float64            `json:"material_cost"`
	JobCost         float64            `json:"job_cost"`
	InventionCost   float64            `json:"invention_cost"`
	TotalCost       float64            `json:"total_cost"`
	Revenue         float64            `json:"revenue"`
	SalesTax        float64            `json:"sales_tax"`
	BrokerFee       float64            `json:"broker_fee"`
	Profit          float64            `json:"profit"`
	ROI             float64            `json:"roi"`
	BestDecryptor   *DecryptorOption   `json:"best_decryptor"`
	AllDecryptors   []*DecryptorOption `json:"all_decryptors"`
	IsBlacklisted   bool               `json:"is_blacklisted"`
	IsWhitelisted   bool               `json:"is_whitelisted"`
}

// ArbiterScanResult is the full result of a scan run.
type ArbiterScanResult struct {
	Opportunities     []*ArbiterOpportunity `json:"opportunities"`
	GeneratedAt       time.Time             `json:"generated_at"`
	TotalScanned      int                   `json:"total_scanned"`
	BestCharacterID   int64                 `json:"best_character_id"`
	BestCharacterName string                `json:"best_character_name"`
}
