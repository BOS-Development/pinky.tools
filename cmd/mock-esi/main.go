package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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

// Character assets keyed by character ID
var characterAssets = map[int64][]asset{
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
}

// Character asset names (container names)
var characterNames = map[int64][]nameEntry{
	2001001: {
		{ItemID: 100010, Name: "Minerals Box"},
	},
}

// Corporation assets keyed by corp ID
var corpAssets = map[int64][]asset{
	3001001: { // Stargazer Industries
		{ItemID: 500000, LocationFlag: "OfficeFolder", LocationID: 60003760, LocationType: "item", Quantity: 1, TypeID: 27, IsSingleton: true}, // Office
		{ItemID: 500001, LocationFlag: "CorpSAG1", LocationID: 500000, LocationType: "item", Quantity: 100000, TypeID: 34},                    // Tritanium
		{ItemID: 500002, LocationFlag: "CorpSAG2", LocationID: 500000, LocationType: "item", Quantity: 5, TypeID: 587, IsSingleton: true},      // Rifter
	},
}

var corpDivisions = map[int64]divisionsResponse{
	3001001: {
		Hangar: []divisionEntry{
			{Division: 1, Name: "Main Hangar"},
			{Division: 2, Name: "Production Materials"},
		},
		Wallet: []divisionEntry{
			{Division: 1, Name: "Master Wallet"},
		},
	},
}

// Character to corporation mapping
var charToCorp = map[int64]int64{
	2001001: 3001001,
	2001002: 3001001,
	2002001: 3002001,
	2003001: 3003001,
	2004001: 3004001,
}

var corpNames = map[int64]string{
	3001001: "Stargazer Industries",
	3002001: "Bob's Mining Co",
	3003001: "Charlie Trade Corp",
	3004001: "Scout Fleet",
}

// Character skills keyed by character ID
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

var characterSkills = map[int64]skillsResponse{
	2001001: {
		Skills: []skillEntry{
			{SkillID: 3380, TrainedSkillLevel: 5, ActiveSkillLevel: 5, SkillpointsInSkill: 256000},  // Industry
			{SkillID: 3388, TrainedSkillLevel: 5, ActiveSkillLevel: 5, SkillpointsInSkill: 256000},  // Advanced Industry
			{SkillID: 45746, TrainedSkillLevel: 4, ActiveSkillLevel: 4, SkillpointsInSkill: 45255},  // Reactions
		},
		TotalSP: 5000000,
	},
	2002001: {
		Skills: []skillEntry{
			{SkillID: 3380, TrainedSkillLevel: 4, ActiveSkillLevel: 4, SkillpointsInSkill: 45255},
		},
		TotalSP: 2000000,
	},
}

// Character industry jobs
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

// Character blueprints keyed by character ID
var characterBlueprints = map[int64][]blueprintEntry{
	// Alice Alpha — a BPO ME10 and a BPC ME8
	2001001: {
		{ItemID: 700001, TypeID: 787, LocationID: 60003760, LocationFlag: "Hangar", Quantity: -1, MaterialEfficiency: 10, TimeEfficiency: 20, Runs: -1},
		{ItemID: 700002, TypeID: 46166, LocationID: 60003760, LocationFlag: "Hangar", Quantity: -2, MaterialEfficiency: 8, TimeEfficiency: 16, Runs: 50},
	},
	// Bob Bravo — a BPO ME8
	2002001: {
		{ItemID: 700003, TypeID: 787, LocationID: 60003760, LocationFlag: "Hangar", Quantity: -1, MaterialEfficiency: 8, TimeEfficiency: 16, Runs: -1},
	},
}

// Corporation blueprints keyed by corp ID
var corpBlueprints = map[int64][]blueprintEntry{
	3001001: {
		{ItemID: 710001, TypeID: 787, LocationID: 60003760, LocationFlag: "CorpSAG1", Quantity: -1, MaterialEfficiency: 9, TimeEfficiency: 18, Runs: -1},
	},
}

var characterIndustryJobs = map[int64][]industryJob{
	2001001: {
		{
			JobID: 500001, InstallerID: 2001001, FacilityID: 60003760, StationID: 60003760,
			ActivityID: 1, BlueprintID: 9876, BlueprintTypeID: 787,
			BlueprintLocationID: 60003760, OutputLocationID: 60003760,
			Runs: 10, Cost: 1500000, ProductTypeID: 587, Status: "active",
			Duration: 3600, StartDate: "2026-02-22T00:00:00Z", EndDate: "2026-02-22T01:00:00Z",
		},
	},
}

var marketOrders = []marketOrder{
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
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
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

	mux := http.NewServeMux()

	// GET /characters/{id}/assets
	mux.HandleFunc("/characters/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// POST /characters/affiliation
		if strings.HasSuffix(path, "/affiliation") && r.Method == "POST" {
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
			names, ok := characterNames[charID]
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
			assets, ok := characterAssets[charID]
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
			skills, ok := characterSkills[charID]
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
			jobs, ok := characterIndustryJobs[charID]
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
			bps, ok := characterBlueprints[charID]
			if !ok {
				bps = []blueprintEntry{}
			}
			w.Header().Set("X-Pages", "1")
			writeJSON(w, bps)
			return
		}

		http.Error(w, "not found", 404)
	})

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
			assets, ok := corpAssets[corpID]
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
			bps, ok := corpBlueprints[corpID]
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
			divs, ok := corpDivisions[corpID]
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
		name, ok := corpNames[corpID]
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
	var knownNames = map[int64]struct {
		Name     string
		Category string
	}{
		60003760: {Name: "Jita IV - Moon 4 - Caldari Navy Assembly Plant", Category: "station"},
		60008494: {Name: "Amarr VIII (Oris) - Emperor Family Academy", Category: "station"},
	}
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
		results := []nameResult{}
		for _, id := range ids {
			if entry, ok := knownNames[id]; ok {
				results = append(results, nameResult{ID: id, Name: entry.Name, Category: entry.Category})
			}
		}
		writeJSON(w, results)
	})

	// GET /latest/markets/{regionID}/orders/
	mux.HandleFunc("/latest/markets/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Pages", "1")
		writeJSON(w, marketOrders)
	})

	log.Printf("Mock ESI server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
