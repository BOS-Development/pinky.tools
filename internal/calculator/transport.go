package calculator

import (
	"math"

	"github.com/annymsMthd/industry-tool/internal/models"
)

// GateTransportCostParams holds inputs for gate-based transport cost calculation.
type GateTransportCostParams struct {
	TotalVolumeM3   float64
	TotalCollateral float64
	Jumps           int
	CargoM3         float64
	RatePerM3PerJump float64
	CollateralRate  float64
}

// GateTransportCostResult holds the result of a gate transport cost calculation.
type GateTransportCostResult struct {
	Trips int     `json:"trips"`
	Cost  float64 `json:"cost"`
}

// CalculateGateTransportCost calculates the cost for gate-based transport (freighter, DST, blockade runner).
func CalculateGateTransportCost(params *GateTransportCostParams) *GateTransportCostResult {
	if params.CargoM3 <= 0 || params.Jumps <= 0 {
		return &GateTransportCostResult{Trips: 1, Cost: 0}
	}

	trips := int(math.Ceil(params.TotalVolumeM3 / params.CargoM3))
	if trips < 1 {
		trips = 1
	}

	haulCost := params.TotalVolumeM3 * params.RatePerM3PerJump * float64(params.Jumps)
	collateralCost := params.TotalCollateral * params.CollateralRate
	cost := (haulCost + collateralCost) * float64(trips)

	return &GateTransportCostResult{Trips: trips, Cost: cost}
}

// JFTransportCostParams holds inputs for jump freighter transport cost calculation.
type JFTransportCostParams struct {
	TotalVolumeM3         float64
	TotalCollateral       float64
	CargoM3               float64
	CollateralRate        float64
	FuelPerLY             float64
	FuelConservationLevel int
	IsotopePrice          float64
	Waypoints             []*models.JFRouteWaypoint
}

// JFTransportCostResult holds the result of a JF transport cost calculation.
type JFTransportCostResult struct {
	Trips     int     `json:"trips"`
	TotalFuel int     `json:"totalFuel"`
	FuelCost  float64 `json:"fuelCost"`
	Cost      float64 `json:"cost"`
}

// CalculateJFTransportCost calculates the cost for jump freighter transport.
func CalculateJFTransportCost(params *JFTransportCostParams) *JFTransportCostResult {
	if params.CargoM3 <= 0 {
		return &JFTransportCostResult{Trips: 1}
	}

	trips := int(math.Ceil(params.TotalVolumeM3 / params.CargoM3))
	if trips < 1 {
		trips = 1
	}

	// Calculate fuel for one trip across all waypoint legs
	fuelConservationFactor := 1.0 - float64(params.FuelConservationLevel)*0.10
	totalFuel := 0
	for _, wp := range params.Waypoints {
		if wp.DistanceLY > 0 {
			fuelUnits := int(math.Ceil(params.FuelPerLY * wp.DistanceLY * fuelConservationFactor))
			totalFuel += fuelUnits
		}
	}

	fuelCost := float64(totalFuel) * params.IsotopePrice
	collateralCost := params.TotalCollateral * params.CollateralRate
	cost := (fuelCost + collateralCost) * float64(trips)

	return &JFTransportCostResult{
		Trips:     trips,
		TotalFuel: totalFuel * trips,
		FuelCost:  fuelCost * float64(trips),
		Cost:      cost,
	}
}

// CourierCostParams holds inputs for courier contract cost calculation.
type CourierCostParams struct {
	TotalVolumeM3       float64
	TotalCollateral     float64
	CourierRatePerM3    float64
	CourierCollateralRate float64
}

// CalculateCourierCost calculates the cost for a courier contract or contact haul (flat rate).
func CalculateCourierCost(params *CourierCostParams) float64 {
	return (params.TotalVolumeM3 * params.CourierRatePerM3) + (params.TotalCollateral * params.CourierCollateralRate)
}

// CalculateCollateralValue calculates the total collateral value for a set of items
// using the specified price basis (buy, sell, or split).
func CalculateCollateralValue(items []*models.TransportJobItem, jitaPrices map[int64]*models.MarketPrice, priceBasis string) float64 {
	total := 0.0
	for _, item := range items {
		price := jitaPrices[item.TypeID]
		if price == nil {
			continue
		}

		var unitPrice float64
		switch priceBasis {
		case "buy":
			if price.BuyPrice != nil {
				unitPrice = *price.BuyPrice
			}
		case "sell":
			if price.SellPrice != nil {
				unitPrice = *price.SellPrice
			}
		case "split":
			buyPrice := 0.0
			sellPrice := 0.0
			if price.BuyPrice != nil {
				buyPrice = *price.BuyPrice
			}
			if price.SellPrice != nil {
				sellPrice = *price.SellPrice
			}
			unitPrice = (buyPrice + sellPrice) / 2.0
		default:
			if price.SellPrice != nil {
				unitPrice = *price.SellPrice
			}
		}

		total += unitPrice * float64(item.Quantity)
	}
	return total
}
