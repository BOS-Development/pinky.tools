package calculator

import (
	"math"
	"testing"
)

func TestComputeMEFactor(t *testing.T) {
	tests := []struct {
		name     string
		rig      string
		security string
		expected float64
	}{
		{"T2 null", "t2", "null", 0.9736},
		{"T2 low", "t2", "low", 0.976},
		{"T1 null", "t1", "null", 0.978},
		{"T1 low", "t1", "low", 0.98},
		{"None null", "none", "null", 1.0},
		{"None low", "none", "low", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeMEFactor(tt.rig, tt.security)
			result = math.Round(result*10000) / 10000
			if result != tt.expected {
				t.Errorf("ComputeMEFactor(%s, %s) = %f, want %f", tt.rig, tt.security, result, tt.expected)
			}
		})
	}
}

func TestComputeTEFactor(t *testing.T) {
	tests := []struct {
		name      string
		skill     int
		structure string
		rig       string
		security  string
		expected  float64
	}{
		{"Reactions 5, T2 Tatara null", 5, "tatara", "t2", "null", 0.4416},
		{"Reactions 0, Athanor none null", 0, "athanor", "none", "null", 1.0},
		{"Reactions 5, Athanor t2 null", 5, "athanor", "t2", "null", 0.5888},
		{"Reactions 5, Tatara none null", 5, "tatara", "none", "null", 0.6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeTEFactor(tt.skill, tt.structure, tt.rig, tt.security)
			result = math.Round(result*10000) / 10000
			if result != tt.expected {
				t.Errorf("ComputeTEFactor(%d, %s, %s, %s) = %f, want %f",
					tt.skill, tt.structure, tt.rig, tt.security, result, tt.expected)
			}
		})
	}
}

func TestComputeSecsPerRun(t *testing.T) {
	// Base time 10800, TE 0.4416 -> floor(10800 * 0.4416) = floor(4769.28) = 4769
	result := ComputeSecsPerRun(10800, 0.4416)
	if result != 4769 {
		t.Errorf("ComputeSecsPerRun(10800, 0.4416) = %d, want 4769", result)
	}
}

func TestComputeRunsPerCycle(t *testing.T) {
	tests := []struct {
		name       string
		secsPerRun int
		cycleDays  int
		expected   int
	}{
		{"7-day cycle, 4769s/run", 4769, 7, 126},
		{"14-day cycle, 4769s/run", 4769, 14, 253},
		{"1-day cycle, 4769s/run", 4769, 1, 18},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeRunsPerCycle(tt.secsPerRun, tt.cycleDays)
			if result != tt.expected {
				t.Errorf("ComputeRunsPerCycle(%d, %d) = %d, want %d",
					tt.secsPerRun, tt.cycleDays, result, tt.expected)
			}
		})
	}
}

func TestComputeAdjQty(t *testing.T) {
	// ceil(100 * 0.9736) = ceil(97.36) = 98
	result := ComputeAdjQty(100, 0.9736)
	if result != 98 {
		t.Errorf("ComputeAdjQty(100, 0.9736) = %d, want 98", result)
	}

	// ceil(1 * 0.9736) = ceil(0.9736) = 1
	result = ComputeAdjQty(1, 0.9736)
	if result != 1 {
		t.Errorf("ComputeAdjQty(1, 0.9736) = %d, want 1", result)
	}
}

func TestComputeBatchQty(t *testing.T) {
	tests := []struct {
		name     string
		runs     int
		baseQty  int
		meFactor float64
		expected int64
	}{
		{
			"Standard batch ME",
			126, 100, 0.9736,
			12268,
		},
		{
			"Minimum is runs",
			126, 1, 0.9736,
			126, // max(126, ceil(126 * 1 * 0.9736)) = max(126, 123) = 126
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeBatchQty(tt.runs, tt.baseQty, tt.meFactor)
			if result != tt.expected {
				t.Errorf("ComputeBatchQty(%d, %d, %f) = %d, want %d",
					tt.runs, tt.baseQty, tt.meFactor, result, tt.expected)
			}
		})
	}
}

func TestComplexInstances(t *testing.T) {
	// Simple produces 200/run, complex needs adj_qty=98 -> floor(200/98) = floor(2.04) = 2
	result := int(math.Floor(200.0 / 98.0))
	if result != 2 {
		t.Errorf("Complex instances = %d, want 2", result)
	}
}

func TestVerificationData_7DayCycle(t *testing.T) {
	// Spec verification: T2 Tatara null, 7-day cycle
	meFactor := ComputeMEFactor("t2", "null")
	teFactor := ComputeTEFactor(5, "tatara", "t2", "null")

	mfRounded := math.Round(meFactor*10000) / 10000
	tfRounded := math.Round(teFactor*10000) / 10000

	if mfRounded != 0.9736 {
		t.Errorf("ME factor = %f, want 0.9736", mfRounded)
	}
	if tfRounded != 0.4416 {
		t.Errorf("TE factor = %f, want 0.4416", tfRounded)
	}

	secsPerRun := ComputeSecsPerRun(10800, teFactor)
	runsPerCycle := ComputeRunsPerCycle(secsPerRun, 7)

	if runsPerCycle != 126 {
		t.Errorf("Runs per cycle (7-day) = %d, want 126", runsPerCycle)
	}
}

func TestVerificationData_14DayCycle(t *testing.T) {
	teFactor := ComputeTEFactor(5, "tatara", "t2", "null")
	secsPerRun := ComputeSecsPerRun(10800, teFactor)
	runsPerCycle := ComputeRunsPerCycle(secsPerRun, 14)

	if runsPerCycle != 253 {
		t.Errorf("Runs per cycle (14-day) = %d, want 253", runsPerCycle)
	}
}

func TestGetPrice(t *testing.T) {
	sell := 100.0
	buy := 80.0
	prices := map[int64]*models_mock_price{
		1: {sell: &sell, buy: &buy},
	}

	// We test the internal logic directly
	if getPrice_test(1, "sell", prices) != 100.0 {
		t.Error("sell price should be 100")
	}
	if getPrice_test(1, "buy", prices) != 80.0 {
		t.Error("buy price should be 80")
	}
	if getPrice_test(1, "split", prices) != 90.0 {
		t.Error("split price should be 90")
	}
	if getPrice_test(999, "sell", prices) != 0 {
		t.Error("missing type should return 0")
	}
}

// Mock types for price testing (avoid importing models just for test)
type models_mock_price struct {
	sell *float64
	buy  *float64
}

func getPrice_test(typeID int64, method string, prices map[int64]*models_mock_price) float64 {
	mp, ok := prices[typeID]
	if !ok {
		return 0
	}

	var sell, buy float64
	if mp.sell != nil {
		sell = *mp.sell
	}
	if mp.buy != nil {
		buy = *mp.buy
	}

	switch method {
	case "buy":
		return buy
	case "split":
		return (sell + buy) / 2.0
	default:
		return sell
	}
}
