package updaters

import (
	"context"

	"github.com/pkg/errors"

	"github.com/annymsMthd/industry-tool/internal/models"
)

type FuzzWorksClient interface {
	GetInventoryTypes(ctx context.Context) ([]models.EveInventoryType, error)
	GetRegions(ctx context.Context) ([]models.Region, error)
	GetConstellations(ctx context.Context) ([]models.Constellation, error)
	GetSolarSystems(ctx context.Context) ([]models.SolarSystem, error)
	GetNPCStations(ctx context.Context) ([]models.Station, error)
}

type ItemTypeRepository interface {
	UpsertItemTypes(ctx context.Context, itemTypes []models.EveInventoryType) error
}

type RegionRepository interface {
	Upsert(ctx context.Context, regions []models.Region) error
}

type ConstellationRepository interface {
	Upsert(ctx context.Context, constellations []models.Constellation) error
}

type SolarSystemRepository interface {
	Upsert(ctx context.Context, systems []models.SolarSystem) error
}

type StationRepository interface {
	Upsert(ctx context.Context, stations []models.Station) error
}

type Static struct {
	client                  FuzzWorksClient
	itemTypeRepository      ItemTypeRepository
	regionRepository        RegionRepository
	constellationRepository ConstellationRepository
	solarSystemRepository   SolarSystemRepository
	stationRepository       StationRepository
}

func NewStatic(
	client FuzzWorksClient,
	itemTypeRepository ItemTypeRepository,
	regionRepository RegionRepository,
	constellationRepository ConstellationRepository,
	solarSystemRepository SolarSystemRepository,
	stationRepository StationRepository) *Static {
	return &Static{
		client:                  client,
		itemTypeRepository:      itemTypeRepository,
		regionRepository:        regionRepository,
		constellationRepository: constellationRepository,
		solarSystemRepository:   solarSystemRepository,
		stationRepository:       stationRepository,
	}
}

func (u *Static) Update(ctx context.Context) error {
	regions, err := u.client.GetRegions(ctx)
	if err != nil {
		return errors.Wrap(err, "failed getting regions from fuzzworks")
	}

	err = u.regionRepository.Upsert(ctx, regions)
	if err != nil {
		return errors.Wrap(err, "failed to upsert regions to repository")
	}

	constellations, err := u.client.GetConstellations(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get constellations from fuzzworks")
	}

	err = u.constellationRepository.Upsert(ctx, constellations)
	if err != nil {
		return errors.Wrap(err, "failed to upsert constellations to repository")
	}

	systems, err := u.client.GetSolarSystems(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get solar systems from fuzzworks")
	}

	err = u.solarSystemRepository.Upsert(ctx, systems)
	if err != nil {
		return errors.Wrap(err, "failed to upsert systems to repository")
	}

	stations, err := u.client.GetNPCStations(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get npc stations from fuzzworks")
	}

	err = u.stationRepository.Upsert(ctx, stations)
	if err != nil {
		return errors.Wrap(err, "failed to upsert stations to repository")
	}

	items, err := u.client.GetInventoryTypes(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get inventory types from client")
	}

	err = u.itemTypeRepository.UpsertItemTypes(ctx, items)
	if err != nil {
		return errors.Wrap(err, "failed to upsert items to itemTypeRepository")
	}

	return nil
}
