package repositories

import (
	"context"
)

func (r *Assets) GetUserAssets(ctx context.Context, user int64) (*AssetsResponse, error) {
	response := &AssetsResponse{}

	stationMap, structures, err := r.loadCharStations(ctx, user)
	if err != nil {
		return nil, err
	}
	response.Structures = structures

	if err = r.loadCharItems(ctx, user, stationMap); err != nil {
		return nil, err
	}

	containerMap, err := r.loadCharContainers(ctx, user, stationMap)
	if err != nil {
		return nil, err
	}

	if err = r.loadCharContainerItems(ctx, user, containerMap); err != nil {
		return nil, err
	}

	stationCorpMap, err := r.loadCorpStations(ctx, user, stationMap, response)
	if err != nil {
		return nil, err
	}

	divisionTemplates, err := r.loadCorpDivisionTemplates(ctx, user)
	if err != nil {
		return nil, err
	}

	hangerMap := map[int64]map[int64]map[int64]*CorporationHanger{}

	if err = r.loadCorpItems(ctx, user, stationCorpMap, hangerMap, divisionTemplates); err != nil {
		return nil, err
	}

	corpContainerMap, containersByDivision, err := r.loadCorpContainers(ctx, user, stationCorpMap, hangerMap, divisionTemplates)
	if err != nil {
		return nil, err
	}

	if err = r.loadCorpContainerItems(ctx, user, corpContainerMap); err != nil {
		return nil, err
	}

	attachCorpDivisions(stationMap, stationCorpMap, divisionTemplates, hangerMap, containersByDivision, response)

	return response, nil
}
