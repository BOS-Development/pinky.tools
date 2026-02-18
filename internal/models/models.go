package models

import "time"

type EveAsset struct {
	ItemID          int64  `json:"item_id"`
	IsBlueprintCopy bool   `json:"is_blueprint_copy"`
	IsSingleton     bool   `json:"is_singleton"`
	LocationFlag    string `json:"location_flag"`
	LocationID      int64  `json:"location_id"`
	LocationType    string `json:"location_type"`
	Quantity        int64  `json:"quantity"`
	TypeID          int64  `json:"type_id"`
}

type EveInventoryType struct {
	TypeID         int64
	TypeName       string
	Volume         float64
	IconID         *int64
	GroupID        *int64
	PackagedVolume *float64
	Mass           *float64
	Capacity       *float64
	PortionSize    *int
	Published      *bool
	MarketGroupID  *int64
	GraphicID      *int64
	RaceID         *int64
	Description    *string
}

type Region struct {
	ID   int64
	Name string
}

type Constellation struct {
	ID       int64
	Name     string
	RegionID int64
}

type SolarSystem struct {
	ID              int64
	Name            string
	ConstellationID int64
	Security        float64
}

type Station struct {
	ID            int64
	Name          string
	SolarSystemID int64
	CorporationID int64
	IsNPC         bool
}

type Corporation struct {
	ID   int64
	Name string
}

type CorporationDivisions struct {
	Hanger map[int]string
	Wallet map[int]string
}

type StockpileMarker struct {
	UserID          int64   `json:"userId"`
	TypeID          int64   `json:"typeId"`
	OwnerType       string  `json:"ownerType"`
	OwnerID         int64   `json:"ownerId"`
	LocationID      int64   `json:"locationId"`
	ContainerID     *int64  `json:"containerId"`
	DivisionNumber  *int    `json:"divisionNumber"`
	DesiredQuantity int64   `json:"desiredQuantity"`
	Notes           *string `json:"notes"`
}

type MarketPrice struct {
	TypeID      int64
	RegionID    int64
	BuyPrice    *float64
	SellPrice   *float64
	DailyVolume *int64
	UpdatedAt   string
}

type Contact struct {
	ID              int64      `json:"id"`
	RequesterUserID int64      `json:"requesterUserId"`
	RecipientUserID int64      `json:"recipientUserId"`
	RequesterName   string     `json:"requesterName"`
	RecipientName   string     `json:"recipientName"`
	Status          string     `json:"status"`
	RequestedAt     time.Time  `json:"requestedAt"`
	RespondedAt     *time.Time `json:"respondedAt"`
}

type ContactPermission struct {
	ID              int64  `json:"id"`
	ContactID       int64  `json:"contactId"`
	GrantingUserID  int64  `json:"grantingUserId"`
	ReceivingUserID int64  `json:"receivingUserId"`
	ServiceType     string `json:"serviceType"`
	CanAccess       bool   `json:"canAccess"`
}

type ForSaleItem struct {
	ID                  int64     `json:"id"`
	UserID              int64     `json:"userId"`
	TypeID              int64     `json:"typeId"`
	TypeName            string    `json:"typeName"`
	OwnerType           string    `json:"ownerType"`
	OwnerID             int64     `json:"ownerId"`
	OwnerName           string    `json:"ownerName"`
	LocationID          int64     `json:"locationId"`
	LocationName        string    `json:"locationName"`
	ContainerID         *int64    `json:"containerId"`
	DivisionNumber      *int      `json:"divisionNumber"`
	QuantityAvailable   int64     `json:"quantityAvailable"`
	PricePerUnit        float64   `json:"pricePerUnit"`
	Notes               *string   `json:"notes"`
	AutoSellContainerID *int64    `json:"autoSellContainerId"`
	IsActive            bool      `json:"isActive"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type AutoSellContainer struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"userId"`
	OwnerType       string    `json:"ownerType"`
	OwnerID         int64     `json:"ownerId"`
	LocationID      int64     `json:"locationId"`
	ContainerID     int64     `json:"containerId"`
	DivisionNumber  *int      `json:"divisionNumber"`
	PricePercentage float64   `json:"pricePercentage"`
	IsActive        bool      `json:"isActive"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type ContainerItem struct {
	TypeID   int64
	Quantity int64
}

type PurchaseTransaction struct {
	ID                int64     `json:"id"`
	ForSaleItemID     int64     `json:"forSaleItemId"`
	BuyerUserID       int64     `json:"buyerUserId"`
	BuyerName         string    `json:"buyerName"`
	SellerUserID      int64     `json:"sellerUserId"`
	TypeID            int64     `json:"typeId"`
	TypeName          string    `json:"typeName"`
	LocationID        int64     `json:"locationId"`
	LocationName      string    `json:"locationName"`
	QuantityPurchased int64     `json:"quantityPurchased"`
	PricePerUnit      float64   `json:"pricePerUnit"`
	TotalPrice        float64   `json:"totalPrice"`
	Status            string    `json:"status"`
	ContractKey       *string   `json:"contractKey,omitempty"`
	TransactionNotes  *string   `json:"transactionNotes"`
	PurchasedAt       time.Time `json:"purchasedAt"`
}

type BuyOrder struct {
	ID               int64     `json:"id"`
	BuyerUserID      int64     `json:"buyerUserId"`
	TypeID           int64     `json:"typeId"`
	TypeName         string    `json:"typeName"`
	QuantityDesired  int64     `json:"quantityDesired"`
	MaxPricePerUnit  float64   `json:"maxPricePerUnit"`
	Notes            *string   `json:"notes"`
	IsActive         bool      `json:"isActive"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// Sales Analytics Models

type SalesMetrics struct {
	TotalRevenue      float64        `json:"totalRevenue"`
	TotalTransactions int64          `json:"totalTransactions"`
	TotalQuantitySold int64          `json:"totalQuantitySold"`
	UniqueItemTypes   int64          `json:"uniqueItemTypes"`
	UniqueBuyers      int64          `json:"uniqueBuyers"`
	TimeSeriesData    []TimeSeriesData `json:"timeSeriesData"`
	TopItems          []ItemSalesData  `json:"topItems"`
}

type TimeSeriesData struct {
	Date              string `json:"date"`
	Revenue           float64 `json:"revenue"`
	Transactions      int64   `json:"transactions"`
	QuantitySold      int64  `json:"quantitySold"`
}

type ItemSalesData struct {
	TypeID            int64   `json:"typeId"`
	TypeName          string  `json:"typeName"`
	QuantitySold      int64   `json:"quantitySold"`
	Revenue             float64 `json:"revenue"`
	TransactionCount    int64   `json:"transactionCount"`
	AveragePricePerUnit float64 `json:"averagePricePerUnit"`
}

type BuyerAnalytics struct {
	BuyerUserID       int64     `json:"buyerUserId"`
	BuyerName         string    `json:"buyerName"`
	TotalSpent        float64   `json:"totalSpent"`
	TotalPurchases    int64     `json:"totalPurchases"`
	TotalQuantity     int64     `json:"totalQuantity"`
	FirstPurchaseDate time.Time `json:"firstPurchaseDate"`
	LastPurchaseDate  time.Time `json:"lastPurchaseDate"`
	RepeatCustomer    bool      `json:"repeatCustomer"`
}

// SDE Models

type SdeMetadata struct {
	Key       string
	Value     string
	UpdatedAt time.Time
}

type SdeCategory struct {
	CategoryID int64
	Name       string
	Published  bool
	IconID     *int64
}

type SdeGroup struct {
	GroupID    int64
	Name       string
	CategoryID int64
	Published  bool
	IconID     *int64
}

type SdeMetaGroup struct {
	MetaGroupID int64
	Name        string
}

type SdeMarketGroup struct {
	MarketGroupID int64
	ParentGroupID *int64
	Name          string
	Description   *string
	IconID        *int64
	HasTypes      bool
}

type SdeIcon struct {
	IconID      int64
	Description *string
}

type SdeGraphic struct {
	GraphicID   int64
	Description *string
}

type SdeBlueprint struct {
	BlueprintTypeID    int64
	MaxProductionLimit *int
}

type SdeBlueprintActivity struct {
	BlueprintTypeID int64
	Activity        string
	Time            int
}

type SdeBlueprintMaterial struct {
	BlueprintTypeID int64
	Activity        string
	TypeID          int64
	Quantity        int
}

type SdeBlueprintProduct struct {
	BlueprintTypeID int64
	Activity        string
	TypeID          int64
	Quantity        int
	Probability     *float64
}

type SdeBlueprintSkill struct {
	BlueprintTypeID int64
	Activity        string
	TypeID          int64
	Level           int
}

type SdeDogmaAttributeCategory struct {
	CategoryID  int64
	Name        *string
	Description *string
}

type SdeDogmaAttribute struct {
	AttributeID  int64
	Name         *string
	Description  *string
	DefaultValue *float64
	DisplayName  *string
	CategoryID   *int64
	HighIsGood   *bool
	Stackable    *bool
	Published    *bool
	UnitID       *int64
}

type SdeDogmaEffect struct {
	EffectID    int64
	Name        *string
	Description *string
	DisplayName *string
	CategoryID  *int64
}

type SdeTypeDogmaAttribute struct {
	TypeID      int64
	AttributeID int64
	Value       float64
}

type SdeTypeDogmaEffect struct {
	TypeID    int64
	EffectID  int64
	IsDefault bool
}

type SdeFaction struct {
	FactionID     int64
	Name          string
	Description   *string
	CorporationID *int64
	IconID        *int64
}

type SdeNpcCorporation struct {
	CorporationID int64
	Name          string
	FactionID     *int64
	IconID        *int64
}

type SdeNpcCorporationDivision struct {
	CorporationID int64
	DivisionID    int64
	Name          string
}

type SdeAgent struct {
	AgentID       int64
	Name          *string
	CorporationID *int64
	DivisionID    *int64
	Level         *int
}

type SdeAgentInSpace struct {
	AgentID       int64
	SolarSystemID *int64
}

type SdeRace struct {
	RaceID      int64
	Name        string
	Description *string
	IconID      *int64
}

type SdeBloodline struct {
	BloodlineID int64
	Name        string
	RaceID      *int64
	Description *string
	IconID      *int64
}

type SdeAncestry struct {
	AncestryID  int64
	Name        string
	BloodlineID *int64
	Description *string
	IconID      *int64
}

type SdePlanetSchematic struct {
	SchematicID int64
	Name        string
	CycleTime   int
}

type SdePlanetSchematicType struct {
	SchematicID int64
	TypeID      int64
	Quantity    int
	IsInput     bool
}

type SdeControlTowerResource struct {
	ControlTowerTypeID int64
	ResourceTypeID     int64
	Purpose            *int
	Quantity           int
	MinSecurity        *float64
	FactionID          *int64
}

type IndustryCostIndex struct {
	SystemID  int64
	Activity  string
	CostIndex float64
	UpdatedAt time.Time
}

type SdeSkin struct {
	SkinID     int64
	TypeID     *int64
	MaterialID *int64
}

type SdeSkinLicense struct {
	LicenseTypeID int64
	SkinID        *int64
	Duration      *int
}

type SdeSkinMaterial struct {
	SkinMaterialID int64
	Name           *string
}

type SdeCertificate struct {
	CertificateID int64
	Name          *string
	Description   *string
	GroupID       *int64
}

type SdeLandmark struct {
	LandmarkID  int64
	Name        *string
	Description *string
}

type SdeStationOperation struct {
	OperationID int64
	Name        *string
	Description *string
}

type SdeStationService struct {
	ServiceID   int64
	Name        *string
	Description *string
}

type SdeContrabandType struct {
	FactionID    int64
	TypeID       int64
	StandingLoss *float64
	FineByValue  *float64
}

type SdeResearchAgent struct {
	AgentID int64
	TypeID  int64
}

type SdeCharacterAttribute struct {
	AttributeID int64
	Name        *string
	Description *string
	IconID      *int64
}

type SdeCorporationActivity struct {
	ActivityID int64
	Name       *string
}

type SdeTournamentRuleSet struct {
	RuleSetID int64
	Data      *string
}
