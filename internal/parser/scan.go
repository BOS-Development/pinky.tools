package parser

import (
	"strings"

	"github.com/annymsMthd/industry-tool/internal/models"
)

// ParseStructureScan parses an EVE Online structure fitting scan text and extracts
// rigs (with category and tier) and services (with activity type).
func ParseStructureScan(scanText string) *models.ScanResult {
	result := &models.ScanResult{
		Rigs:     []models.ScanRig{},
		Services: []models.ScanService{},
	}

	lines := strings.Split(scanText, "\n")
	section := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Detect section headers
		lower := strings.ToLower(line)
		if lower == "high power slots" || lower == "medium power slots" || lower == "low power slots" {
			section = "other"
			continue
		}
		if lower == "rig slots" {
			section = "rigs"
			continue
		}
		if lower == "service slots" {
			section = "services"
			continue
		}

		switch section {
		case "rigs":
			rig := parseRig(line)
			if rig != nil {
				// Detect structure type from first valid rig
				if result.Structure == "" {
					result.Structure = detectStructureFromRig(line, rig.Category)
				}
				result.Rigs = append(result.Rigs, *rig)
			}
		case "services":
			svc := parseService(line)
			if svc != nil {
				result.Services = append(result.Services, *svc)
			}
		}
	}

	return result
}

func parseRig(name string) *models.ScanRig {
	tier := detectTier(name)
	if tier == "" {
		return nil
	}

	category := detectRigCategory(name)
	if category == "" {
		return nil
	}

	return &models.ScanRig{
		Name:     name,
		Category: category,
		Tier:     tier,
	}
}

func detectTier(name string) string {
	if strings.HasSuffix(name, " II") {
		return "t2"
	}
	if strings.HasSuffix(name, " I") {
		return "t1"
	}
	return ""
}

func detectRigCategory(name string) string {
	upper := strings.ToUpper(name)

	if strings.Contains(upper, "SHIP MANUFACTURING") {
		return "ship"
	}
	if strings.Contains(upper, "STRUCTURE AND COMPONENT") {
		return "component"
	}
	if strings.Contains(upper, "EQUIPMENT MANUFACTURING") {
		return "equipment"
	}
	if strings.Contains(upper, "AMMUNITION MANUFACTURING") {
		return "ammo"
	}
	if strings.Contains(upper, "DRONE AND FIGHTER") {
		return "drone"
	}
	if strings.Contains(upper, "BIOCHEMICAL REACTOR") {
		return "reaction"
	}
	if strings.Contains(upper, "COMPOSITE REACTOR") {
		return "reaction"
	}
	if strings.Contains(upper, "HYBRID REACTOR") {
		return "reaction"
	}
	if strings.Contains(upper, "POLYMER REACTOR") {
		return "reaction"
	}
	if strings.Contains(upper, "THUKKER COMPONENT") {
		return "thukker"
	}
	if strings.Contains(upper, "REACTOR") {
		return "reaction"
	}
	if strings.Contains(upper, "REPROCESSING") {
		return "reprocessing"
	}

	// Lab rigs, moon drilling, etc. — not manufacturing/reaction
	return ""
}

func detectStructureFromRig(name string, category string) string {
	upper := strings.ToUpper(name)

	isRefinery := category == "reaction" || category == "reprocessing"

	if strings.Contains(upper, "XL-SET") {
		return "sotiyo"
	}
	if strings.Contains(upper, "L-SET") {
		if isRefinery {
			return "tatara"
		}
		return "azbel"
	}
	if strings.Contains(upper, "M-SET") {
		if isRefinery {
			return "athanor"
		}
		return "raitaru"
	}

	return ""
}

func parseService(name string) *models.ScanService {
	upper := strings.ToUpper(name)

	if strings.Contains(upper, "MANUFACTURING PLANT") ||
		strings.Contains(upper, "CAPITAL SHIPYARD") ||
		strings.Contains(upper, "SUPERCAPITAL SHIPYARD") {
		return &models.ScanService{
			Name:     name,
			Activity: "manufacturing",
		}
	}

	if strings.Contains(upper, "COMPOSITE REACTOR") ||
		strings.Contains(upper, "BIOCHEMICAL REACTOR") ||
		strings.Contains(upper, "HYBRID REACTOR") ||
		strings.Contains(upper, "POLYMER REACTOR") {
		return &models.ScanService{
			Name:     name,
			Activity: "reaction",
		}
	}

	return nil
}
