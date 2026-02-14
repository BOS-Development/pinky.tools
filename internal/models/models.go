package models

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
	TypeID   int64
	TypeName string
	Volume   float64
	IconID   *int64
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
