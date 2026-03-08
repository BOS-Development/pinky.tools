package repositories

import (
	"database/sql"
)

type Asset struct {
	Name            string   `json:"name"`
	TypeID          int64    `json:"typeId"`
	Quantity        int64    `json:"quantity"`
	Volume          float64  `json:"volume"`
	OwnerType       string   `json:"ownerType"`
	OwnerName       string   `json:"ownerName"`
	OwnerID         int64    `json:"ownerId"`
	DesiredQuantity *int64   `json:"desiredQuantity"`
	StockpileDelta  *int64   `json:"stockpileDelta"`
	UnitPrice       *float64 `json:"unitPrice"`
	TotalValue      *float64 `json:"totalValue"`
	DeficitValue    *float64 `json:"deficitValue"`
}

type AssetsResponse struct {
	Structures []*AssetStructure `json:"structures"`
}

type AssetStructure struct {
	ID                 int64                 `json:"id"`
	Name               string                `json:"name"`
	SolarSystem        string                `json:"solarSystem"`
	Region             string                `json:"region"`
	HangarAssets       []*Asset              `json:"hangarAssets"`
	HangarContainers   []*AssetContainer     `json:"hangarContainers"`
	Deliveries         []*Asset              `json:"deliveries"`
	AssetSafety        []*Asset              `json:"assetSafety"`
	CorporationHangers []*CorporationHanger  `json:"corporationHangers"`
}

type CorporationHanger struct {
	ID               int64             `json:"id"`
	Name             string            `json:"name"`
	CorporationID    int64             `json:"corporationId"`
	CorporationName  string            `json:"corporationName"`
	Assets           []*Asset          `json:"assets"`
	HangarContainers []*AssetContainer `json:"hangarContainers"`
}

type AssetContainer struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	OwnerType string   `json:"ownerType"`
	OwnerName string   `json:"ownerName"`
	OwnerID   int64    `json:"ownerId"`
	Assets    []*Asset `json:"assets"`
}

type AssetsSummary struct {
	TotalValue   float64 `json:"totalValue"`
	TotalDeficit float64 `json:"totalDeficit"`
	ActiveJobs   int     `json:"activeJobs"`
}

type Assets struct {
	db *sql.DB
}

func NewAssets(db *sql.DB) *Assets {
	return &Assets{
		db: db,
	}
}
