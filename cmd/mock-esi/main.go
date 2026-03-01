package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

type asset struct {
	ItemID       int64  `json:"item_id"`
	IsBPC        bool   `json:"is_blueprint_copy"`
	IsSingleton  bool   `json:"is_singleton"`
	LocationFlag string `json:"location_flag"`
	LocationID   int64  `json:"location_id"`
	LocationType string `json:"location_type"`
	Quantity     int64  `json:"quantity"`
	TypeID       int64  `json:"type_id"`
}

type nameEntry struct {
	ItemID int64  `json:"item_id"`
	Name   string `json:"name"`
}

type affiliation struct {
	CorporationID int64 `json:"corporation_id"`
	CharacterID   int64 `json:"character_id"`
}

type corpInfo struct {
	Name string `json:"name"`
}

type divisionEntry struct {
	Division int    `json:"division"`
	Name     string `json:"name"`
}

type divisionsResponse struct {
	Hangar []divisionEntry `json:"hangar"`
	Wallet []divisionEntry `json:"wallet"`
}

type structureInfo struct {
	Name          string `json:"name"`
	OwnerID       int64  `json:"owner_id"`
	SolarSystemID int64  `json:"solar_system_id"`
}

type marketOrder struct {
	OrderID      int64   `json:"order_id"`
	TypeID       int64   `json:"type_id"`
	LocationID   int64   `json:"location_id"`
	VolumeTotal  int64   `json:"volume_total"`
	VolumeRemain int64   `json:"volume_remain"`
	MinVolume    int64   `json:"min_volume"`
	Price        float64 `json:"price"`
	IsBuyOrder   bool    `json:"is_buy_order"`
	Duration     int     `json:"duration"`
	Issued       string  `json:"issued"`
	Range        string  `json:"range"`
}

type skillEntry struct {
	SkillID            int64 `json:"skill_id"`
	TrainedSkillLevel  int   `json:"trained_skill_level"`
	ActiveSkillLevel   int   `json:"active_skill_level"`
	SkillpointsInSkill int64 `json:"skillpoints_in_skill"`
}

type skillsResponse struct {
	Skills  []skillEntry `json:"skills"`
	TotalSP int64        `json:"total_sp"`
}

type industryJob struct {
	JobID               int64   `json:"job_id"`
	InstallerID         int64   `json:"installer_id"`
	FacilityID          int64   `json:"facility_id"`
	StationID           int64   `json:"station_id"`
	ActivityID          int     `json:"activity_id"`
	BlueprintID         int64   `json:"blueprint_id"`
	BlueprintTypeID     int64   `json:"blueprint_type_id"`
	BlueprintLocationID int64   `json:"blueprint_location_id"`
	OutputLocationID    int64   `json:"output_location_id"`
	Runs                int     `json:"runs"`
	Cost                float64 `json:"cost"`
	ProductTypeID       int64   `json:"product_type_id"`
	Status              string  `json:"status"`
	Duration            int     `json:"duration"`
	StartDate           string  `json:"start_date"`
	EndDate             string  `json:"end_date"`
}

type blueprintEntry struct {
	ItemID             int64  `json:"item_id"`
	TypeID             int64  `json:"type_id"`
	LocationID         int64  `json:"location_id"`
	LocationFlag       string `json:"location_flag"`
	Quantity           int    `json:"quantity"`
	MaterialEfficiency int    `json:"material_efficiency"`
	TimeEfficiency     int    `json:"time_efficiency"`
	Runs               int    `json:"runs"`
}

type knownNameEntry struct {
	Name     string
	Category string
}

type marketHistoryEntry struct {
	Date       string  `json:"date"`
	Average    float64 `json:"average"`
	Highest    float64 `json:"highest"`
	Lowest     float64 `json:"lowest"`
	Volume     int64   `json:"volume"`
	OrderCount int64   `json:"order_count"`
}

// PI types

type piPlanet struct {
	LastUpdate    string `json:"last_update"`
	NumPins       int    `json:"num_pins"`
	OwnerID       int64  `json:"owner_id"`
	PlanetID      int64  `json:"planet_id"`
	PlanetType    string `json:"planet_type"`
	SolarSystemID int64  `json:"solar_system_id"`
	UpgradeLevel  int    `json:"upgrade_level"`
}

type piPinContent struct {
	Amount float64 `json:"amount"`
	TypeID int64   `json:"type_id"`
}

type piExtractorDetail struct {
	CycleTime     int   `json:"cycle_time"`
	HeadRadius    float64 `json:"head_radius"`
	Heads         []struct{} `json:"heads"`
	ProductTypeID int64 `json:"product_type_id"`
	QtyPerCycle   int   `json:"qty_per_cycle"`
}

type piFactoryDetail struct {
	SchematicID int `json:"schematic_id"`
}

type piPin struct {
	PinID            int64              `json:"pin_id"`
	TypeID           int64              `json:"type_id"`
	Latitude         float64            `json:"latitude"`
	Longitude        float64            `json:"longitude"`
	InstallTime      *string            `json:"install_time,omitempty"`
	ExpiryTime       *string            `json:"expiry_time,omitempty"`
	LastCycleStart   *string            `json:"last_cycle_start,omitempty"`
	SchematicID      *int               `json:"schematic_id,omitempty"`
	Contents         []piPinContent     `json:"contents"`
	ExtractorDetails *piExtractorDetail `json:"extractor_details,omitempty"`
	FactoryDetails   *piFactoryDetail   `json:"factory_details,omitempty"`
}

type piLink struct {
	SourcePinID      int64 `json:"source_pin_id"`
	DestinationPinID int64 `json:"destination_pin_id"`
	LinkLevel        int   `json:"link_level"`
}

type piRoute struct {
	RouteID          int64   `json:"route_id"`
	SourcePinID      int64   `json:"source_pin_id"`
	DestinationPinID int64   `json:"destination_pin_id"`
	ContentTypeID    int64   `json:"content_type_id"`
	Quantity         float64 `json:"quantity"`
	Waypoints        []int64 `json:"waypoints"`
}

type piColony struct {
	Links  []piLink  `json:"links"`
	Pins   []piPin   `json:"pins"`
	Routes []piRoute `json:"routes"`
}

// State holds all mock ESI data behind a RWMutex for concurrent access safety.
// The backend fetches multiple characters in parallel, so reads must be protected.
type State struct {
	mu                    sync.RWMutex
	characterAssets       map[int64][]asset
	characterNames        map[int64][]nameEntry
	corpAssets            map[int64][]asset
	corpDivisions         map[int64]divisionsResponse
	charToCorp            map[int64]int64
	corpNames             map[int64]string
	characterSkills       map[int64]skillsResponse
	characterIndustryJobs map[int64][]industryJob
	characterBlueprints   map[int64][]blueprintEntry
	corpBlueprints        map[int64][]blueprintEntry
	marketOrders          []marketOrder
	marketHistory         []marketHistoryEntry
	knownNames            map[int64]knownNameEntry
	characterPlanets      map[int64][]piPlanet
	planetDetails         map[string]piColony
}

// newDefaultState returns a fresh State populated with the standard E2E test fixtures.
func newDefaultState() *State {
	return &State{
		characterAssets: map[int64][]asset{
			// Alice Alpha — assets in Jita
			2001001: {
				{ItemID: 100001, LocationFlag: "Hangar", LocationID: 60003760, LocationType: "station", Quantity: 50000, TypeID: 34},  // Tritanium
				{ItemID: 100002, LocationFlag: "Hangar", LocationID: 60003760, LocationType: "station", Quantity: 25000, TypeID: 35},  // Pyerite
				{ItemID: 100003, LocationFlag: "Hangar", LocationID: 60003760, LocationType: "station", Quantity: 10000, TypeID: 36},  // Mexallon
				{ItemID: 100004, LocationFlag: "Hangar", LocationID: 60003760, LocationType: "station", Quantity: 1, TypeID: 11399, IsSingleton: true},    // Raven Navy Issue
				{ItemID: 100010, LocationFlag: "Hangar", LocationID: 60003760, LocationType: "station", Quantity: 1, TypeID: 9999001, IsSingleton: true}, // Container
				{ItemID: 100011, LocationFlag: "Hangar", LocationID: 100010, LocationType: "item", Quantity: 5000, TypeID: 37},        // Isogen in container
			},
			// Alice Beta — assets in Amarr
			2001002: {
				{ItemID: 110001, LocationFlag: "Hangar", LocationID: 60008494, LocationType: "station", Quantity: 3, TypeID: 587, IsSingleton: true},  // Rifter
				{ItemID: 110002, LocationFlag: "Hangar", LocationID: 60008494, LocationType: "station", Quantity: 5000, TypeID: 38},   // Nocxium
			},
			// Bob Bravo — assets in Jita
			2002001: {
				{ItemID: 200001, LocationFlag: "Hangar", LocationID: 60003760, LocationType: "station", Quantity: 30000, TypeID: 34},  // Tritanium
				{ItemID: 200002, LocationFlag: "Hangar", LocationID: 60003760, LocationType: "station", Quantity: 10, TypeID: 587, IsSingleton: true},     // Rifter
			},
			// Charlie Charlie — assets in Jita
			2003001: {
				{ItemID: 300001, LocationFlag: "Hangar", LocationID: 60003760, LocationType: "station", Quantity: 1000, TypeID: 35},   // Pyerite
			},
			// Diana Delta — assets in Amarr
			2004001: {
				{ItemID: 400001, LocationFlag: "Hangar", LocationID: 60008494, LocationType: "station", Quantity: 15000, TypeID: 34},  // Tritanium
			},
		},
		characterNames: map[int64][]nameEntry{
			2001001: {
				{ItemID: 100010, Name: "Minerals Box"},
			},
		},
		corpAssets: map[int64][]asset{
			3001001: { // Stargazer Industries
				{ItemID: 500000, LocationFlag: "OfficeFolder", LocationID: 60003760, LocationType: "item", Quantity: 1, TypeID: 27, IsSingleton: true}, // Office
				{ItemID: 500001, LocationFlag: "CorpSAG1", LocationID: 500000, LocationType: "item", Quantity: 100000, TypeID: 34},                    // Tritanium
				{ItemID: 500002, LocationFlag: "CorpSAG2", LocationID: 500000, LocationType: "item", Quantity: 5, TypeID: 587, IsSingleton: true},      // Rifter
			},
		},
		corpDivisions: map[int64]divisionsResponse{
			3001001: {
				Hangar: []divisionEntry{
					{Division: 1, Name: "Main Hangar"},
					{Division: 2, Name: "Production Materials"},
				},
				Wallet: []divisionEntry{
					{Division: 1, Name: "Master Wallet"},
				},
			},
		},
		charToCorp: map[int64]int64{
			2001001: 3001001,
			2001002: 3001001,
			2002001: 3002001,
			2003001: 3003001,
			2004001: 3004001,
		},
		corpNames: map[int64]string{
			3001001: "Stargazer Industries",
			3002001: "Bob's Mining Co",
			3003001: "Charlie Trade Corp",
			3004001: "Scout Fleet",
		},
		characterSkills: map[int64]skillsResponse{
			2001001: {
				Skills: []skillEntry{
					{SkillID: 3380, TrainedSkillLevel: 5, ActiveSkillLevel: 5, SkillpointsInSkill: 256000},   // Industry
					{SkillID: 3388, TrainedSkillLevel: 5, ActiveSkillLevel: 5, SkillpointsInSkill: 256000},   // Advanced Industry
					{SkillID: 3387, TrainedSkillLevel: 5, ActiveSkillLevel: 5, SkillpointsInSkill: 256000},   // Mass Production
					{SkillID: 24625, TrainedSkillLevel: 4, ActiveSkillLevel: 4, SkillpointsInSkill: 135765},  // Advanced Mass Production
					{SkillID: 45746, TrainedSkillLevel: 4, ActiveSkillLevel: 4, SkillpointsInSkill: 135765},  // Reactions
					{SkillID: 45748, TrainedSkillLevel: 3, ActiveSkillLevel: 3, SkillpointsInSkill: 40000},   // Mass Reactions
					{SkillID: 45749, TrainedSkillLevel: 2, ActiveSkillLevel: 2, SkillpointsInSkill: 11314},   // Advanced Mass Reactions
					{SkillID: 3402, TrainedSkillLevel: 4, ActiveSkillLevel: 4, SkillpointsInSkill: 135765},   // Science
					{SkillID: 3406, TrainedSkillLevel: 3, ActiveSkillLevel: 3, SkillpointsInSkill: 40000},    // Laboratory Operation
					{SkillID: 24624, TrainedSkillLevel: 2, ActiveSkillLevel: 2, SkillpointsInSkill: 11314},   // Advanced Laboratory Operation
				},
				TotalSP: 5000000,
			},
			2002001: {
				Skills: []skillEntry{
					{SkillID: 3380, TrainedSkillLevel: 4, ActiveSkillLevel: 4, SkillpointsInSkill: 135765},  // Industry
					{SkillID: 3387, TrainedSkillLevel: 3, ActiveSkillLevel: 3, SkillpointsInSkill: 40000},   // Mass Production
				},
				TotalSP: 2000000,
			},
		},
		characterIndustryJobs: map[int64][]industryJob{
			2001001: {
				{
					JobID: 500001, InstallerID: 2001001, FacilityID: 60003760, StationID: 60003760,
					ActivityID: 1, BlueprintID: 9876, BlueprintTypeID: 787,
					BlueprintLocationID: 60003760, OutputLocationID: 60003760,
					Runs: 10, Cost: 1500000, ProductTypeID: 587, Status: "active",
					Duration: 3600, StartDate: "2026-02-22T00:00:00Z", EndDate: "2026-02-22T01:00:00Z",
				},
			},
		},
		characterBlueprints: map[int64][]blueprintEntry{
			// Alice Alpha — a BPO ME10 and a BPC ME8
			2001001: {
				{ItemID: 700001, TypeID: 787, LocationID: 60003760, LocationFlag: "Hangar", Quantity: -1, MaterialEfficiency: 10, TimeEfficiency: 20, Runs: -1},
				{ItemID: 700002, TypeID: 46166, LocationID: 60003760, LocationFlag: "Hangar", Quantity: -2, MaterialEfficiency: 8, TimeEfficiency: 16, Runs: 50},
			},
			// Bob Bravo — a BPO ME8
			2002001: {
				{ItemID: 700003, TypeID: 787, LocationID: 60003760, LocationFlag: "Hangar", Quantity: -1, MaterialEfficiency: 8, TimeEfficiency: 16, Runs: -1},
			},
		},
		corpBlueprints: map[int64][]blueprintEntry{
			3001001: {
				{ItemID: 710001, TypeID: 787, LocationID: 60003760, LocationFlag: "CorpSAG1", Quantity: -1, MaterialEfficiency: 9, TimeEfficiency: 18, Runs: -1},
			},
		},
		marketOrders: []marketOrder{
			// Tritanium sell
			{OrderID: 1, TypeID: 34, LocationID: 60003760, VolumeTotal: 10000000, VolumeRemain: 5000000, MinVolume: 1, Price: 6.00, IsBuyOrder: false, Duration: 90, Issued: "2025-01-01T00:00:00Z", Range: "station"},
			// Tritanium buy
			{OrderID: 2, TypeID: 34, LocationID: 60003760, VolumeTotal: 10000000, VolumeRemain: 5000000, MinVolume: 1, Price: 5.50, IsBuyOrder: true, Duration: 90, Issued: "2025-01-01T00:00:00Z", Range: "station"},
			// Pyerite sell
			{OrderID: 3, TypeID: 35, LocationID: 60003760, VolumeTotal: 5000000, VolumeRemain: 2000000, MinVolume: 1, Price: 11.50, IsBuyOrder: false, Duration: 90, Issued: "2025-01-01T00:00:00Z", Range: "station"},
			// Pyerite buy
			{OrderID: 4, TypeID: 35, LocationID: 60003760, VolumeTotal: 5000000, VolumeRemain: 2000000, MinVolume: 1, Price: 10.00, IsBuyOrder: true, Duration: 90, Issued: "2025-01-01T00:00:00Z", Range: "station"},
			// Mexallon sell
			{OrderID: 5, TypeID: 36, LocationID: 60003760, VolumeTotal: 1000000, VolumeRemain: 500000, MinVolume: 1, Price: 75.00, IsBuyOrder: false, Duration: 90, Issued: "2025-01-01T00:00:00Z", Range: "station"},
			// Mexallon buy
			{OrderID: 6, TypeID: 36, LocationID: 60003760, VolumeTotal: 1000000, VolumeRemain: 500000, MinVolume: 1, Price: 70.00, IsBuyOrder: true, Duration: 90, Issued: "2025-01-01T00:00:00Z", Range: "station"},
			// Isogen sell
			{OrderID: 11, TypeID: 37, LocationID: 60003760, VolumeTotal: 500000, VolumeRemain: 200000, MinVolume: 1, Price: 55.00, IsBuyOrder: false, Duration: 90, Issued: "2025-01-01T00:00:00Z", Range: "station"},
			// Isogen buy
			{OrderID: 12, TypeID: 37, LocationID: 60003760, VolumeTotal: 500000, VolumeRemain: 200000, MinVolume: 1, Price: 50.00, IsBuyOrder: true, Duration: 90, Issued: "2025-01-01T00:00:00Z", Range: "station"},
			// Rifter sell
			{OrderID: 7, TypeID: 587, LocationID: 60003760, VolumeTotal: 100, VolumeRemain: 50, MinVolume: 1, Price: 600000, IsBuyOrder: false, Duration: 90, Issued: "2025-01-01T00:00:00Z", Range: "region"},
			// Rifter buy
			{OrderID: 8, TypeID: 587, LocationID: 60003760, VolumeTotal: 100, VolumeRemain: 50, MinVolume: 1, Price: 500000, IsBuyOrder: true, Duration: 90, Issued: "2025-01-01T00:00:00Z", Range: "region"},
			// Raven Navy Issue sell
			{OrderID: 9, TypeID: 11399, LocationID: 60003760, VolumeTotal: 10, VolumeRemain: 5, MinVolume: 1, Price: 520000000, IsBuyOrder: false, Duration: 90, Issued: "2025-01-01T00:00:00Z", Range: "region"},
			// Raven Navy Issue buy
			{OrderID: 10, TypeID: 11399, LocationID: 60003760, VolumeTotal: 10, VolumeRemain: 5, MinVolume: 1, Price: 500000000, IsBuyOrder: true, Duration: 90, Issued: "2025-01-01T00:00:00Z", Range: "region"},
		},
		knownNames: map[int64]knownNameEntry{
			60003760: {Name: "Jita IV - Moon 4 - Caldari Navy Assembly Plant", Category: "station"},
			60008494: {Name: "Amarr VIII (Oris) - Emperor Family Academy", Category: "station"},
		},
		marketHistory: []marketHistoryEntry{
			{Date: "2026-01-30", Average: 5.45, Highest: 5.60, Lowest: 5.30, Volume: 120000, OrderCount: 60},
			{Date: "2026-01-31", Average: 5.48, Highest: 5.62, Lowest: 5.32, Volume: 115000, OrderCount: 58},
			{Date: "2026-02-01", Average: 5.50, Highest: 5.65, Lowest: 5.35, Volume: 130000, OrderCount: 65},
			{Date: "2026-02-02", Average: 5.52, Highest: 5.68, Lowest: 5.38, Volume: 125000, OrderCount: 62},
			{Date: "2026-02-03", Average: 5.49, Highest: 5.64, Lowest: 5.34, Volume: 118000, OrderCount: 59},
			{Date: "2026-02-04", Average: 5.51, Highest: 5.66, Lowest: 5.36, Volume: 122000, OrderCount: 61},
			{Date: "2026-02-05", Average: 5.53, Highest: 5.70, Lowest: 5.40, Volume: 128000, OrderCount: 64},
			{Date: "2026-02-06", Average: 5.47, Highest: 5.63, Lowest: 5.33, Volume: 117000, OrderCount: 57},
			{Date: "2026-02-07", Average: 5.50, Highest: 5.65, Lowest: 5.35, Volume: 126000, OrderCount: 63},
			{Date: "2026-02-08", Average: 5.54, Highest: 5.72, Lowest: 5.42, Volume: 132000, OrderCount: 66},
			{Date: "2026-02-09", Average: 5.46, Highest: 5.61, Lowest: 5.31, Volume: 114000, OrderCount: 57},
			{Date: "2026-02-10", Average: 5.48, Highest: 5.63, Lowest: 5.33, Volume: 119000, OrderCount: 59},
			{Date: "2026-02-11", Average: 5.50, Highest: 5.66, Lowest: 5.36, Volume: 124000, OrderCount: 62},
			{Date: "2026-02-12", Average: 5.52, Highest: 5.68, Lowest: 5.38, Volume: 127000, OrderCount: 63},
			{Date: "2026-02-13", Average: 5.55, Highest: 5.71, Lowest: 5.41, Volume: 133000, OrderCount: 67},
			{Date: "2026-02-14", Average: 5.49, Highest: 5.64, Lowest: 5.34, Volume: 121000, OrderCount: 60},
			{Date: "2026-02-15", Average: 5.51, Highest: 5.67, Lowest: 5.37, Volume: 129000, OrderCount: 64},
			{Date: "2026-02-16", Average: 5.53, Highest: 5.69, Lowest: 5.39, Volume: 131000, OrderCount: 65},
			{Date: "2026-02-17", Average: 5.47, Highest: 5.62, Lowest: 5.32, Volume: 116000, OrderCount: 58},
			{Date: "2026-02-18", Average: 5.50, Highest: 5.65, Lowest: 5.35, Volume: 123000, OrderCount: 61},
			{Date: "2026-02-19", Average: 5.52, Highest: 5.67, Lowest: 5.37, Volume: 128000, OrderCount: 64},
			{Date: "2026-02-20", Average: 5.54, Highest: 5.70, Lowest: 5.40, Volume: 130000, OrderCount: 65},
			{Date: "2026-02-21", Average: 5.48, Highest: 5.63, Lowest: 5.33, Volume: 118000, OrderCount: 59},
			{Date: "2026-02-22", Average: 5.51, Highest: 5.66, Lowest: 5.36, Volume: 125000, OrderCount: 62},
			{Date: "2026-02-23", Average: 5.53, Highest: 5.68, Lowest: 5.38, Volume: 126000, OrderCount: 63},
			{Date: "2026-02-24", Average: 5.56, Highest: 5.73, Lowest: 5.43, Volume: 134000, OrderCount: 67},
			{Date: "2026-02-25", Average: 5.50, Highest: 5.65, Lowest: 5.35, Volume: 122000, OrderCount: 61},
			{Date: "2026-02-26", Average: 5.52, Highest: 5.67, Lowest: 5.37, Volume: 127000, OrderCount: 63},
			{Date: "2026-02-27", Average: 5.55, Highest: 5.71, Lowest: 5.41, Volume: 132000, OrderCount: 66},
			{Date: "2026-02-28", Average: 5.57, Highest: 5.74, Lowest: 5.44, Volume: 135000, OrderCount: 68},
		},
		// PI data — empty by default; tests inject via admin API
		characterPlanets: map[int64][]piPlanet{},
		planetDetails:    map[string]piColony{},
	}
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func writeAdminOK(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func writeAdminError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}

func extractID(path, prefix, suffix string) (int64, bool) {
	path = strings.TrimPrefix(path, prefix)
	if suffix != "" {
		path = strings.TrimSuffix(path, suffix)
	}
	// Remove query params
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}
	// Remove trailing slashes
	path = strings.TrimRight(path, "/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	state := newDefaultState()

	mux := http.NewServeMux()

	// GET /characters/{id}/assets (and other character endpoints)
	mux.HandleFunc("/characters/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// POST /characters/affiliation
		if strings.HasSuffix(path, "/affiliation") && r.Method == "POST" {
			state.mu.RLock()
			charToCorp := state.charToCorp
			state.mu.RUnlock()

			var charIDs []int64
			json.NewDecoder(r.Body).Decode(&charIDs)
			result := []affiliation{}
			for _, id := range charIDs {
				corpID, ok := charToCorp[id]
				if !ok {
					corpID = 98000001
				}
				result = append(result, affiliation{CorporationID: corpID, CharacterID: id})
			}
			writeJSON(w, result)
			return
		}

		// POST /characters/{id}/assets/names
		if strings.HasSuffix(path, "/assets/names") && r.Method == "POST" {
			charID, ok := extractID(path, "/characters/", "/assets/names")
			if !ok {
				http.Error(w, "invalid character id", 400)
				return
			}
			state.mu.RLock()
			names, ok := state.characterNames[charID]
			state.mu.RUnlock()
			if !ok {
				names = []nameEntry{}
			}
			writeJSON(w, names)
			return
		}

		// GET /characters/{id}/assets
		if strings.Contains(path, "/assets") && r.Method == "GET" {
			charID, ok := extractID(path, "/characters/", "/assets")
			if !ok {
				http.Error(w, "invalid character id", 400)
				return
			}
			state.mu.RLock()
			assets, ok := state.characterAssets[charID]
			state.mu.RUnlock()
			if !ok {
				assets = []asset{}
			}
			w.Header().Set("X-Pages", "1")
			writeJSON(w, assets)
			return
		}

		// GET /characters/{id}/skills
		if strings.HasSuffix(path, "/skills") || strings.HasSuffix(path, "/skills/") {
			charID, ok := extractID(path, "/characters/", "/skills")
			if !ok {
				http.Error(w, "invalid character id", 400)
				return
			}
			state.mu.RLock()
			skills, ok := state.characterSkills[charID]
			state.mu.RUnlock()
			if !ok {
				skills = skillsResponse{Skills: []skillEntry{}, TotalSP: 0}
			}
			writeJSON(w, skills)
			return
		}

		// GET /characters/{id}/industry/jobs
		if strings.Contains(path, "/industry/jobs") && r.Method == "GET" {
			charID, ok := extractID(path, "/characters/", "/industry/jobs")
			if !ok {
				http.Error(w, "invalid character id", 400)
				return
			}
			state.mu.RLock()
			jobs, ok := state.characterIndustryJobs[charID]
			state.mu.RUnlock()
			if !ok {
				jobs = []industryJob{}
			}
			w.Header().Set("X-Pages", "1")
			writeJSON(w, jobs)
			return
		}

		// GET /characters/{id}/blueprints/
		if strings.Contains(path, "/blueprints") && r.Method == "GET" {
			charID, ok := extractID(path, "/characters/", "/blueprints")
			if !ok {
				http.Error(w, "invalid character id", 400)
				return
			}
			state.mu.RLock()
			bps, ok := state.characterBlueprints[charID]
			state.mu.RUnlock()
			if !ok {
				bps = []blueprintEntry{}
			}
			w.Header().Set("X-Pages", "1")
			writeJSON(w, bps)
			return
		}

		http.Error(w, "not found", 404)
	})

	// Versioned character endpoints (v1, v3, v4) — ESI uses version-prefixed URLs for
	// some endpoints (e.g. /v1/characters/{id}/planets/, /v3/characters/{id}/planets/{id}/,
	// /v4/characters/{id}/skills/, /v3/characters/{id}/blueprints/).
	// Go's ServeMux prefix matching requires separate handlers for these paths.

	// handleVersionedCharacter parses versioned character paths and dispatches to state.
	handleVersionedCharacter := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Normalise: strip the version prefix (/v1/, /v3/, /v4/, etc.)
		// e.g. /v1/characters/2001001/planets/ → /characters/2001001/planets/
		//       /v3/characters/2001001/planets/40000001/ → /characters/2001001/planets/40000001/
		// path[1:] strips leading slash; find the first '/' in the remainder to skip "v1", "v3", etc.
		normPath := path
		if rest1 := path[1:]; len(rest1) > 0 {
			if idx := strings.Index(rest1, "/"); idx != -1 {
				// idx is the position of the slash after the version segment
				// normPath should start from that slash
				normPath = rest1[idx:]
			}
		}
		// normPath is now /characters/2001001/planets/ etc.
		// Strip any trailing slash and query string for extractID compatibility
		normPathClean := strings.TrimRight(strings.SplitN(normPath, "?", 2)[0], "/")

		// GET /characters/{id}/skills  (v4)
		if strings.Contains(normPath, "/skills") && r.Method == "GET" {
			charID, ok := extractID(normPathClean, "/characters/", "/skills")
			if !ok {
				http.Error(w, "invalid character id", 400)
				return
			}
			state.mu.RLock()
			skills, ok := state.characterSkills[charID]
			state.mu.RUnlock()
			if !ok {
				skills = skillsResponse{Skills: []skillEntry{}, TotalSP: 0}
			}
			writeJSON(w, skills)
			return
		}

		// GET /characters/{id}/industry/jobs  (v1)
		if strings.Contains(normPath, "/industry/jobs") && r.Method == "GET" {
			charID, ok := extractID(normPathClean, "/characters/", "/industry/jobs")
			if !ok {
				http.Error(w, "invalid character id", 400)
				return
			}
			state.mu.RLock()
			jobs, ok := state.characterIndustryJobs[charID]
			state.mu.RUnlock()
			if !ok {
				jobs = []industryJob{}
			}
			w.Header().Set("X-Pages", "1")
			writeJSON(w, jobs)
			return
		}

		// GET /characters/{id}/blueprints  (v3)
		if strings.Contains(normPath, "/blueprints") && r.Method == "GET" {
			charID, ok := extractID(normPathClean, "/characters/", "/blueprints")
			if !ok {
				http.Error(w, "invalid character id", 400)
				return
			}
			state.mu.RLock()
			bps, ok := state.characterBlueprints[charID]
			state.mu.RUnlock()
			if !ok {
				bps = []blueprintEntry{}
			}
			w.Header().Set("X-Pages", "1")
			writeJSON(w, bps)
			return
		}

		// GET /characters/{id}/contracts  (v1)
		if strings.Contains(normPath, "/contracts") && r.Method == "GET" {
			// No contracts in test data — return empty paginated list
			w.Header().Set("X-Pages", "1")
			writeJSON(w, []struct{}{})
			return
		}

		// GET /characters/{id}/planets/  (v1) — list of PI colonies
		// GET /characters/{id}/planets/{planetID}/  (v3) — colony details
		if strings.Contains(normPath, "/planets/") && r.Method == "GET" {
			// Determine whether this is the list or detail endpoint by checking
			// what comes after "/planets/" in the normalised path.
			afterPlanets := ""
			if idx := strings.Index(normPath, "/planets/"); idx != -1 {
				tail := normPath[idx+len("/planets/"):]
				// Strip trailing slashes and query params
				tail = strings.TrimRight(tail, "/")
				if qi := strings.Index(tail, "?"); qi != -1 {
					tail = tail[:qi]
				}
				afterPlanets = tail
			}

			if afterPlanets == "" {
				// List: GET /characters/{charID}/planets/
				// extractID strips prefix "/characters/" and suffix "/planets/"
				charID, ok := extractID(normPath, "/characters/", "/planets/")
				if !ok {
					http.Error(w, "invalid character id", 400)
					return
				}
				state.mu.RLock()
				planets, pok := state.characterPlanets[charID]
				state.mu.RUnlock()
				if !pok {
					planets = []piPlanet{}
				}
				writeJSON(w, planets)
				return
			}

			// Detail: GET /characters/{charID}/planets/{planetID}/
			// Extract charID from the segment between "/characters/" and "/planets/"
			rest := strings.TrimPrefix(normPath, "/characters/")
			slashIdx := strings.Index(rest, "/")
			if slashIdx == -1 {
				http.Error(w, "invalid path", 400)
				return
			}
			charID, err := strconv.ParseInt(rest[:slashIdx], 10, 64)
			if err != nil {
				http.Error(w, "invalid character id", 400)
				return
			}
			planetID, err := strconv.ParseInt(afterPlanets, 10, 64)
			if err != nil {
				http.Error(w, "invalid planet id", 400)
				return
			}

			key := fmt.Sprintf("%d:%d", charID, planetID)
			state.mu.RLock()
			colony, ok := state.planetDetails[key]
			state.mu.RUnlock()
			if !ok {
				colony = piColony{Links: []piLink{}, Pins: []piPin{}, Routes: []piRoute{}}
			}
			writeJSON(w, colony)
			return
		}

		http.Error(w, "not found", 404)
	}

	mux.HandleFunc("/v1/characters/", handleVersionedCharacter)
	mux.HandleFunc("/v3/characters/", handleVersionedCharacter)
	mux.HandleFunc("/v4/characters/", handleVersionedCharacter)

	// Corporation endpoints
	mux.HandleFunc("/corporations/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// POST /corporations/{id}/assets/names
		if strings.HasSuffix(path, "/assets/names") && r.Method == "POST" {
			// No named corp containers in our test data
			writeJSON(w, []nameEntry{})
			return
		}

		// GET /corporations/{id}/assets
		if strings.Contains(path, "/assets") && r.Method == "GET" {
			corpID, ok := extractID(path, "/corporations/", "/assets")
			if !ok {
				http.Error(w, "invalid corp id", 400)
				return
			}
			state.mu.RLock()
			assets, ok := state.corpAssets[corpID]
			state.mu.RUnlock()
			if !ok {
				assets = []asset{}
			}
			w.Header().Set("X-Pages", "1")
			writeJSON(w, assets)
			return
		}

		// GET /corporations/{id}/blueprints/
		if strings.Contains(path, "/blueprints") && r.Method == "GET" {
			corpID, ok := extractID(path, "/corporations/", "/blueprints")
			if !ok {
				http.Error(w, "invalid corp id", 400)
				return
			}
			state.mu.RLock()
			bps, ok := state.corpBlueprints[corpID]
			state.mu.RUnlock()
			if !ok {
				bps = []blueprintEntry{}
			}
			w.Header().Set("X-Pages", "1")
			writeJSON(w, bps)
			return
		}

		// GET /corporations/{id}/divisions
		if strings.HasSuffix(path, "/divisions") {
			corpID, ok := extractID(path, "/corporations/", "/divisions")
			if !ok {
				http.Error(w, "invalid corp id", 400)
				return
			}
			state.mu.RLock()
			divs, ok := state.corpDivisions[corpID]
			state.mu.RUnlock()
			if !ok {
				divs = divisionsResponse{Hangar: []divisionEntry{}, Wallet: []divisionEntry{}}
			}
			writeJSON(w, divs)
			return
		}

		// GET /corporations/{id} — corporation info
		corpID, ok := extractID(path, "/corporations/", "")
		if !ok {
			http.Error(w, "invalid corp id", 400)
			return
		}
		state.mu.RLock()
		name, ok := state.corpNames[corpID]
		state.mu.RUnlock()
		if !ok {
			name = fmt.Sprintf("Unknown Corp %d", corpID)
		}
		writeJSON(w, corpInfo{Name: name})
	})

	// GET /universe/structures/{id}
	mux.HandleFunc("/universe/structures/", func(w http.ResponseWriter, r *http.Request) {
		// No player-owned structures in test data, return 403 (access denied)
		http.Error(w, `{"error":"Forbidden"}`, 403)
	})

	// POST /universe/names/
	mux.HandleFunc("/universe/names/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", 405)
			return
		}
		var ids []int64
		if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
			http.Error(w, `{"error":"bad request"}`, 400)
			return
		}
		type nameResult struct {
			ID       int64  `json:"id"`
			Name     string `json:"name"`
			Category string `json:"category"`
		}
		state.mu.RLock()
		knownNames := state.knownNames
		state.mu.RUnlock()

		results := []nameResult{}
		for _, id := range ids {
			if entry, ok := knownNames[id]; ok {
				results = append(results, nameResult{ID: id, Name: entry.Name, Category: entry.Category})
			}
		}
		writeJSON(w, results)
	})

	// GET /latest/markets/{regionID}/orders/
	// GET /latest/markets/{regionID}/history/?type_id={typeID}
	mux.HandleFunc("/latest/markets/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Dispatch history requests: /latest/markets/{regionID}/history/
		if strings.Contains(path, "/history") {
			state.mu.RLock()
			history := state.marketHistory
			state.mu.RUnlock()
			writeJSON(w, history)
			return
		}

		// Default: return market orders for /latest/markets/{regionID}/orders/
		state.mu.RLock()
		orders := state.marketOrders
		state.mu.RUnlock()

		w.Header().Set("X-Pages", "1")
		writeJSON(w, orders)
	})

	// Admin API — only registered when E2E_TESTING=true
	if os.Getenv("E2E_TESTING") == "true" {
		log.Println("E2E_TESTING enabled: registering admin API endpoints")

		// POST /_admin/reset
		mux.HandleFunc("/_admin/reset", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				writeAdminError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			fresh := newDefaultState()
			state.mu.Lock()
			state.characterAssets = fresh.characterAssets
			state.characterNames = fresh.characterNames
			state.corpAssets = fresh.corpAssets
			state.corpDivisions = fresh.corpDivisions
			state.charToCorp = fresh.charToCorp
			state.corpNames = fresh.corpNames
			state.characterSkills = fresh.characterSkills
			state.characterIndustryJobs = fresh.characterIndustryJobs
			state.characterBlueprints = fresh.characterBlueprints
			state.corpBlueprints = fresh.corpBlueprints
			state.marketOrders = fresh.marketOrders
			state.marketHistory = fresh.marketHistory
			state.knownNames = fresh.knownNames
			state.characterPlanets = fresh.characterPlanets
			state.planetDetails = fresh.planetDetails
			state.mu.Unlock()
			writeAdminOK(w)
		})

		// PUT /_admin/character-assets/{charID}
		mux.HandleFunc("/_admin/character-assets/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				writeAdminError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			charID, ok := extractID(r.URL.Path, "/_admin/character-assets/", "")
			if !ok {
				writeAdminError(w, http.StatusBadRequest, "invalid character id")
				return
			}
			var assets []asset
			if err := json.NewDecoder(r.Body).Decode(&assets); err != nil {
				writeAdminError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
				return
			}
			state.mu.Lock()
			state.characterAssets[charID] = assets
			state.mu.Unlock()
			writeAdminOK(w)
		})

		// PUT /_admin/character-skills/{charID}
		mux.HandleFunc("/_admin/character-skills/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				writeAdminError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			charID, ok := extractID(r.URL.Path, "/_admin/character-skills/", "")
			if !ok {
				writeAdminError(w, http.StatusBadRequest, "invalid character id")
				return
			}
			var skills skillsResponse
			if err := json.NewDecoder(r.Body).Decode(&skills); err != nil {
				writeAdminError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
				return
			}
			state.mu.Lock()
			state.characterSkills[charID] = skills
			state.mu.Unlock()
			writeAdminOK(w)
		})

		// PUT /_admin/character-industry-jobs/{charID}
		mux.HandleFunc("/_admin/character-industry-jobs/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				writeAdminError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			charID, ok := extractID(r.URL.Path, "/_admin/character-industry-jobs/", "")
			if !ok {
				writeAdminError(w, http.StatusBadRequest, "invalid character id")
				return
			}
			var jobs []industryJob
			if err := json.NewDecoder(r.Body).Decode(&jobs); err != nil {
				writeAdminError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
				return
			}
			state.mu.Lock()
			state.characterIndustryJobs[charID] = jobs
			state.mu.Unlock()
			writeAdminOK(w)
		})

		// PUT /_admin/character-blueprints/{charID}
		mux.HandleFunc("/_admin/character-blueprints/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				writeAdminError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			charID, ok := extractID(r.URL.Path, "/_admin/character-blueprints/", "")
			if !ok {
				writeAdminError(w, http.StatusBadRequest, "invalid character id")
				return
			}
			var bps []blueprintEntry
			if err := json.NewDecoder(r.Body).Decode(&bps); err != nil {
				writeAdminError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
				return
			}
			state.mu.Lock()
			state.characterBlueprints[charID] = bps
			state.mu.Unlock()
			writeAdminOK(w)
		})

		// PUT /_admin/corp-assets/{corpID}
		mux.HandleFunc("/_admin/corp-assets/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				writeAdminError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			corpID, ok := extractID(r.URL.Path, "/_admin/corp-assets/", "")
			if !ok {
				writeAdminError(w, http.StatusBadRequest, "invalid corp id")
				return
			}
			var assets []asset
			if err := json.NewDecoder(r.Body).Decode(&assets); err != nil {
				writeAdminError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
				return
			}
			state.mu.Lock()
			state.corpAssets[corpID] = assets
			state.mu.Unlock()
			writeAdminOK(w)
		})

		// PUT /_admin/market-orders
		mux.HandleFunc("/_admin/market-orders", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				writeAdminError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			var orders []marketOrder
			if err := json.NewDecoder(r.Body).Decode(&orders); err != nil {
				writeAdminError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
				return
			}
			state.mu.Lock()
			state.marketOrders = orders
			state.mu.Unlock()
			writeAdminOK(w)
		})

		// PUT /_admin/character-planets/{charID}
		mux.HandleFunc("/_admin/character-planets/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				writeAdminError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			charID, ok := extractID(r.URL.Path, "/_admin/character-planets/", "")
			if !ok {
				writeAdminError(w, http.StatusBadRequest, "invalid character id")
				return
			}
			var planets []piPlanet
			if err := json.NewDecoder(r.Body).Decode(&planets); err != nil {
				writeAdminError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
				return
			}
			state.mu.Lock()
			state.characterPlanets[charID] = planets
			state.mu.Unlock()
			writeAdminOK(w)
		})

		// PUT /_admin/planet-details/{charID}/{planetID}
		mux.HandleFunc("/_admin/planet-details/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				writeAdminError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			// Path: /_admin/planet-details/{charID}/{planetID}
			rest := strings.TrimPrefix(r.URL.Path, "/_admin/planet-details/")
			rest = strings.TrimRight(rest, "/")
			parts := strings.SplitN(rest, "/", 2)
			if len(parts) != 2 {
				writeAdminError(w, http.StatusBadRequest, "expected path /_admin/planet-details/{charID}/{planetID}")
				return
			}
			charID, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				writeAdminError(w, http.StatusBadRequest, "invalid character id")
				return
			}
			planetID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				writeAdminError(w, http.StatusBadRequest, "invalid planet id")
				return
			}
			var colony piColony
			if err := json.NewDecoder(r.Body).Decode(&colony); err != nil {
				writeAdminError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
				return
			}
			key := fmt.Sprintf("%d:%d", charID, planetID)
			state.mu.Lock()
			state.planetDetails[key] = colony
			state.mu.Unlock()
			writeAdminOK(w)
		})
	}

	log.Printf("Mock ESI server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
