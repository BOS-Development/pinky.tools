package calculator

import (
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCalculateGateTransportCost(t *testing.T) {
	t.Run("basic freighter cost", func(t *testing.T) {
		result := CalculateGateTransportCost(&GateTransportCostParams{
			TotalVolumeM3:    100000,
			TotalCollateral:  1000000000,
			Jumps:            10,
			CargoM3:          435000,
			RatePerM3PerJump: 2.5,
			CollateralRate:   0.02,
		})

		assert.Equal(t, 1, result.Trips)
		// haulCost = 100000 * 2.5 * 10 = 2,500,000
		// collateralCost = 1,000,000,000 * 0.02 = 20,000,000
		// total = (2,500,000 + 20,000,000) * 1 = 22,500,000
		assert.InDelta(t, 22500000, result.Cost, 0.01)
	})

	t.Run("multi-trip due to volume", func(t *testing.T) {
		result := CalculateGateTransportCost(&GateTransportCostParams{
			TotalVolumeM3:    1000000,
			TotalCollateral:  500000000,
			Jumps:            5,
			CargoM3:          435000,
			RatePerM3PerJump: 2.0,
			CollateralRate:   0.01,
		})

		assert.Equal(t, 3, result.Trips) // ceil(1000000 / 435000) = 3
		// haulCost = 1000000 * 2.0 * 5 = 10,000,000
		// collateralCost = 500,000,000 * 0.01 = 5,000,000
		// total = (10,000,000 + 5,000,000) * 3 = 45,000,000
		assert.InDelta(t, 45000000, result.Cost, 0.01)
	})

	t.Run("zero jumps returns zero cost", func(t *testing.T) {
		result := CalculateGateTransportCost(&GateTransportCostParams{
			TotalVolumeM3: 100000,
			Jumps:         0,
			CargoM3:       435000,
		})

		assert.Equal(t, 0.0, result.Cost)
	})
}

func TestCalculateJFTransportCost(t *testing.T) {
	t.Run("basic JF cost with fuel", func(t *testing.T) {
		result := CalculateJFTransportCost(&JFTransportCostParams{
			TotalVolumeM3:         20000,
			TotalCollateral:       500000000,
			CargoM3:               34246,
			CollateralRate:        0.02,
			FuelPerLY:             500,
			FuelConservationLevel: 5,
			IsotopePrice:          800,
			Waypoints: []*models.JFRouteWaypoint{
				{DistanceLY: 0},   // Origin
				{DistanceLY: 5.0}, // Leg 1: 5 LY
				{DistanceLY: 3.0}, // Leg 2: 3 LY
			},
		})

		assert.Equal(t, 1, result.Trips)
		// fuel_conservation_factor = 1 - 5 * 0.10 = 0.50
		// leg1 fuel = ceil(500 * 5.0 * 0.50) = ceil(1250) = 1250
		// leg2 fuel = ceil(500 * 3.0 * 0.50) = ceil(750) = 750
		// total_fuel = 2000
		assert.Equal(t, 2000, result.TotalFuel)
		// fuel_cost = 2000 * 800 = 1,600,000
		// collateral_cost = 500,000,000 * 0.02 = 10,000,000
		// cost = (1,600,000 + 10,000,000) * 1 = 11,600,000
		assert.InDelta(t, 11600000, result.Cost, 0.01)
	})

	t.Run("multi-trip JF", func(t *testing.T) {
		result := CalculateJFTransportCost(&JFTransportCostParams{
			TotalVolumeM3:         80000,
			TotalCollateral:       200000000,
			CargoM3:               34246,
			CollateralRate:        0.01,
			FuelPerLY:             500,
			FuelConservationLevel: 0,
			IsotopePrice:          1000,
			Waypoints: []*models.JFRouteWaypoint{
				{DistanceLY: 0},
				{DistanceLY: 4.0},
			},
		})

		assert.Equal(t, 3, result.Trips) // ceil(80000 / 34246) = 3
		// fuel per trip: ceil(500 * 4.0 * 1.0) = 2000
		// total fuel across all trips = 2000 * 3 = 6000
		assert.Equal(t, 6000, result.TotalFuel)
	})

	t.Run("zero cargo capacity returns single trip", func(t *testing.T) {
		result := CalculateJFTransportCost(&JFTransportCostParams{
			CargoM3: 0,
		})

		assert.Equal(t, 1, result.Trips)
	})
}

func TestCalculateCourierCost(t *testing.T) {
	cost := CalculateCourierCost(&CourierCostParams{
		TotalVolumeM3:        50000,
		TotalCollateral:      1000000000,
		CourierRatePerM3:     500,
		CourierCollateralRate: 0.02,
	})

	// volume_cost = 50000 * 500 = 25,000,000
	// collateral_cost = 1,000,000,000 * 0.02 = 20,000,000
	// total = 45,000,000
	assert.InDelta(t, 45000000, cost, 0.01)
}

func TestCalculateCollateralValue(t *testing.T) {
	buyPrice := 100.0
	sellPrice := 120.0

	prices := map[int64]*models.MarketPrice{
		34: {TypeID: 34, BuyPrice: &buyPrice, SellPrice: &sellPrice},
	}

	items := []*models.TransportJobItem{
		{TypeID: 34, Quantity: 1000},
	}

	t.Run("buy basis", func(t *testing.T) {
		value := CalculateCollateralValue(items, prices, "buy")
		assert.InDelta(t, 100000, value, 0.01) // 1000 * 100
	})

	t.Run("sell basis", func(t *testing.T) {
		value := CalculateCollateralValue(items, prices, "sell")
		assert.InDelta(t, 120000, value, 0.01) // 1000 * 120
	})

	t.Run("split basis", func(t *testing.T) {
		value := CalculateCollateralValue(items, prices, "split")
		assert.InDelta(t, 110000, value, 0.01) // 1000 * (100+120)/2
	})

	t.Run("unknown item returns zero", func(t *testing.T) {
		unknownItems := []*models.TransportJobItem{
			{TypeID: 99999, Quantity: 1000},
		}
		value := CalculateCollateralValue(unknownItems, prices, "sell")
		assert.Equal(t, 0.0, value)
	})
}
