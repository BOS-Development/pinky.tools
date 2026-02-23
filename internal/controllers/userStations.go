package controllers

import (
	"context"
	"encoding/json"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/parser"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type UserStationsRepository interface {
	GetByUser(ctx context.Context, userID int64) ([]*models.UserStation, error)
	GetByID(ctx context.Context, id, userID int64) (*models.UserStation, error)
	Create(ctx context.Context, station *models.UserStation) (*models.UserStation, error)
	Update(ctx context.Context, station *models.UserStation) error
	Delete(ctx context.Context, id, userID int64) error
}

type UserStationsStationRepository interface {
	IsNpcStation(ctx context.Context, stationID int64) (bool, error)
}

type UserStations struct {
	stationsRepo UserStationsRepository
}

func NewUserStations(router Routerer, stationsRepo UserStationsRepository) *UserStations {
	c := &UserStations{
		stationsRepo: stationsRepo,
	}

	router.RegisterRestAPIRoute("/v1/user-stations/parse-scan", web.AuthAccessUser, c.ParseScan, "POST")
	router.RegisterRestAPIRoute("/v1/user-stations", web.AuthAccessUser, c.GetStations, "GET")
	router.RegisterRestAPIRoute("/v1/user-stations", web.AuthAccessUser, c.CreateStation, "POST")
	router.RegisterRestAPIRoute("/v1/user-stations/{id:[0-9]+}", web.AuthAccessUser, c.UpdateStation, "PUT")
	router.RegisterRestAPIRoute("/v1/user-stations/{id:[0-9]+}", web.AuthAccessUser, c.DeleteStation, "DELETE")

	return c
}

func (c *UserStations) GetStations(args *web.HandlerArgs) (any, *web.HttpError) {
	stations, err := c.stationsRepo.GetByUser(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get user stations")}
	}
	return stations, nil
}

type createStationRequest struct {
	StationID   int64                      `json:"stationId"`
	Structure   string                     `json:"structure"`
	FacilityTax float64                    `json:"facilityTax"`
	Rigs        []*models.UserStationRig    `json:"rigs"`
	Services    []*models.UserStationService `json:"services"`
}

func (c *UserStations) CreateStation(args *web.HandlerArgs) (any, *web.HttpError) {
	var req createStationRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	if req.StationID <= 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("station_id is required")}
	}
	if req.Structure == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("structure is required")}
	}

	station := &models.UserStation{
		UserID:      *args.User,
		StationID:   req.StationID,
		Structure:   req.Structure,
		FacilityTax: req.FacilityTax,
		Rigs:        req.Rigs,
		Services:    req.Services,
	}

	if station.Rigs == nil {
		station.Rigs = []*models.UserStationRig{}
	}
	if station.Services == nil {
		station.Services = []*models.UserStationService{}
	}

	created, err := c.stationsRepo.Create(args.Request.Context(), station)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to create user station")}
	}

	return created, nil
}

type updateStationRequest struct {
	Structure   string                     `json:"structure"`
	FacilityTax float64                    `json:"facilityTax"`
	Rigs        []*models.UserStationRig    `json:"rigs"`
	Services    []*models.UserStationService `json:"services"`
}

func (c *UserStations) UpdateStation(args *web.HandlerArgs) (any, *web.HttpError) {
	id, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid station ID")}
	}

	var req updateStationRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	station := &models.UserStation{
		ID:          id,
		UserID:      *args.User,
		Structure:   req.Structure,
		FacilityTax: req.FacilityTax,
		Rigs:        req.Rigs,
		Services:    req.Services,
	}

	if station.Rigs == nil {
		station.Rigs = []*models.UserStationRig{}
	}
	if station.Services == nil {
		station.Services = []*models.UserStationService{}
	}

	if err := c.stationsRepo.Update(args.Request.Context(), station); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to update user station")}
	}

	return map[string]string{"status": "updated"}, nil
}

func (c *UserStations) DeleteStation(args *web.HandlerArgs) (any, *web.HttpError) {
	id, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid station ID")}
	}

	if err := c.stationsRepo.Delete(args.Request.Context(), id, *args.User); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to delete user station")}
	}

	return map[string]string{"status": "deleted"}, nil
}

type parseScanRequest struct {
	ScanText string `json:"scanText"`
}

func (c *UserStations) ParseScan(args *web.HandlerArgs) (any, *web.HttpError) {
	var req parseScanRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	result := parser.ParseStructureScan(req.ScanText)
	return result, nil
}
