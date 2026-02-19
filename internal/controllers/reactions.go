package controllers

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/annymsMthd/industry-tool/internal/calculator"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type ReactionsSDERepository interface {
	GetAllReactions(ctx context.Context) ([]*repositories.ReactionRow, error)
	GetAllReactionMaterials(ctx context.Context) ([]*repositories.ReactionMaterialRow, error)
	GetReactionSystems(ctx context.Context) ([]*models.ReactionSystem, error)
}

type ReactionsMarketRepository interface {
	GetAllJitaPrices(ctx context.Context) (map[int64]*models.MarketPrice, error)
	GetAllAdjustedPrices(ctx context.Context) (map[int64]float64, error)
}

type ReactionsCostIndicesRepository interface {
	GetCostIndex(ctx context.Context, systemID int64, activity string) (*models.IndustryCostIndex, error)
}

type Reactions struct {
	sdeRepo         ReactionsSDERepository
	marketRepo      ReactionsMarketRepository
	costIndicesRepo ReactionsCostIndicesRepository
}

func NewReactions(router Routerer, sdeRepo ReactionsSDERepository, marketRepo ReactionsMarketRepository, costIndicesRepo ReactionsCostIndicesRepository) *Reactions {
	c := &Reactions{
		sdeRepo:         sdeRepo,
		marketRepo:      marketRepo,
		costIndicesRepo: costIndicesRepo,
	}

	router.RegisterRestAPIRoute("/v1/reactions", web.AuthAccessBackend, c.GetReactions, "GET")
	router.RegisterRestAPIRoute("/v1/reaction-systems", web.AuthAccessBackend, c.GetReactionSystems, "GET")
	router.RegisterRestAPIRoute("/v1/reactions/plan", web.AuthAccessBackend, c.ComputePlan, "POST")

	return c
}

func (c *Reactions) GetReactions(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	params := parseCalcParams(args)

	// Fetch data
	reactions, err := c.sdeRepo.GetAllReactions(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get reactions")}
	}

	materials, err := c.sdeRepo.GetAllReactionMaterials(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get reaction materials")}
	}

	jitaPrices, err := c.marketRepo.GetAllJitaPrices(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get Jita prices")}
	}

	adjustedPrices, err := c.marketRepo.GetAllAdjustedPrices(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get adjusted prices")}
	}

	var costIndex float64
	if params.SystemID > 0 {
		idx, err := c.costIndicesRepo.GetCostIndex(ctx, params.SystemID, "reaction")
		if err != nil {
			return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get cost index")}
		}
		if idx != nil {
			costIndex = idx.CostIndex
		}
	}

	data := &calculator.CalcData{
		Reactions:      reactions,
		Materials:      materials,
		CostIndex:      costIndex,
		JitaPrices:     jitaPrices,
		AdjustedPrices: adjustedPrices,
	}

	result := calculator.Calculate(params, data)
	return result, nil
}

func (c *Reactions) GetReactionSystems(args *web.HandlerArgs) (any, *web.HttpError) {
	systems, err := c.sdeRepo.GetReactionSystems(args.Request.Context())
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get reaction systems")}
	}

	if systems == nil {
		systems = []*models.ReactionSystem{}
	}

	return systems, nil
}

type planRequest struct {
	Selections         []models.PlanSelection `json:"selections"`
	SystemID           int64                  `json:"system_id"`
	Structure          string                 `json:"structure"`
	Rig                string                 `json:"rig"`
	Security           string                 `json:"security"`
	ReactionsSkill     int                    `json:"reactions_skill"`
	FacilityTax        float64                `json:"facility_tax"`
	CycleDays          int                    `json:"cycle_days"`
	BrokerFee          float64                `json:"broker_fee"`
	SalesTax           float64                `json:"sales_tax"`
	ShippingM3         float64                `json:"shipping_m3"`
	ShippingCollateral float64                `json:"shipping_collateral"`
	InputPrice         string                 `json:"input_price"`
	OutputPrice        string                 `json:"output_price"`
	ShipInputs         bool                   `json:"ship_inputs"`
	ShipOutputs        bool                   `json:"ship_outputs"`
}

func (c *Reactions) ComputePlan(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()

	var req planRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	params := &calculator.CalcParams{
		SystemID:           req.SystemID,
		Structure:          withDefault(req.Structure, "tatara"),
		Rig:                withDefault(req.Rig, "t2"),
		Security:           withDefault(req.Security, "null"),
		ReactionsSkill:     req.ReactionsSkill,
		FacilityTax:        req.FacilityTax,
		CycleDays:          intWithDefault(req.CycleDays, 7),
		BrokerFee:          req.BrokerFee,
		SalesTax:           req.SalesTax,
		ShippingM3:         req.ShippingM3,
		ShippingCollateral: req.ShippingCollateral,
		InputPrice:         withDefault(req.InputPrice, "sell"),
		OutputPrice:        withDefault(req.OutputPrice, "sell"),
		ShipInputs:         req.ShipInputs,
		ShipOutputs:        req.ShipOutputs,
	}

	// Fetch data
	reactions, err := c.sdeRepo.GetAllReactions(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get reactions")}
	}

	materials, err := c.sdeRepo.GetAllReactionMaterials(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get reaction materials")}
	}

	jitaPrices, err := c.marketRepo.GetAllJitaPrices(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get Jita prices")}
	}

	adjustedPrices, err := c.marketRepo.GetAllAdjustedPrices(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get adjusted prices")}
	}

	var costIndex float64
	if params.SystemID > 0 {
		idx, err := c.costIndicesRepo.GetCostIndex(ctx, params.SystemID, "reaction")
		if err != nil {
			return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get cost index")}
		}
		if idx != nil {
			costIndex = idx.CostIndex
		}
	}

	data := &calculator.CalcData{
		Reactions:      reactions,
		Materials:      materials,
		CostIndex:      costIndex,
		JitaPrices:     jitaPrices,
		AdjustedPrices: adjustedPrices,
	}

	// First calculate all reactions
	reactionsResponse := calculator.Calculate(params, data)

	// Then compute the plan
	planResponse := calculator.ComputePlan(req.Selections, params, data, reactionsResponse)

	return planResponse, nil
}

func parseCalcParams(args *web.HandlerArgs) *calculator.CalcParams {
	q := args.Request.URL.Query()

	return &calculator.CalcParams{
		SystemID:           parseInt64(q.Get("system_id"), 0),
		Structure:          withDefault(q.Get("structure"), "tatara"),
		Rig:                withDefault(q.Get("rig"), "t2"),
		Security:           withDefault(q.Get("security"), "null"),
		ReactionsSkill:     int(parseInt64(q.Get("reactions_skill"), 5)),
		FacilityTax:        parseFloat(q.Get("facility_tax"), 0.25),
		CycleDays:          int(parseInt64(q.Get("cycle_days"), 7)),
		BrokerFee:          parseFloat(q.Get("broker_fee"), 3.5),
		SalesTax:           parseFloat(q.Get("sales_tax"), 2.25),
		ShippingM3:         parseFloat(q.Get("shipping_m3"), 0),
		ShippingCollateral: parseFloat(q.Get("shipping_collateral"), 0),
		InputPrice:         withDefault(q.Get("input_price"), "sell"),
		OutputPrice:        withDefault(q.Get("output_price"), "sell"),
		ShipInputs:         q.Get("ship_inputs") != "0",
		ShipOutputs:        q.Get("ship_outputs") != "0",
	}
}

func parseInt64(s string, defaultVal int64) int64 {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return defaultVal
	}
	return v
}

func parseFloat(s string, defaultVal float64) float64 {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return defaultVal
	}
	return v
}

func withDefault(s, defaultVal string) string {
	if s == "" {
		return defaultVal
	}
	return s
}

func intWithDefault(v, defaultVal int) int {
	if v <= 0 {
		return defaultVal
	}
	return v
}
