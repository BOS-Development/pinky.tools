package client

import (
	"compress/bzip2"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

//go:generate mockgen -source=./fuzzWorks.go -destination=./fuzzWorks_mock_test.go -package=client_test

type HttpGetter interface {
	Get(url string) (*http.Response, error)
}

type FuzzWorks struct {
	baseUrl string
	client  HttpGetter
}

func NewFuzzWorks(client HttpGetter) *FuzzWorks {
	return &FuzzWorks{
		client:  client,
		baseUrl: "https://www.fuzzwork.co.uk/dump/latest/",
	}
}

func (f *FuzzWorks) GetInventoryTypes(ctx context.Context) ([]models.EveInventoryType, error) {
	res, err := f.client.Get(fmt.Sprintf("%s/invTypes.csv.bz2", f.baseUrl))
	if err != nil {
		return nil, errors.Wrap(err, "failed to pull data from fuzzworks")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("pulling from fuzzworks failed, expected status code 200 got %d", res.StatusCode))
	}

	r := bzip2.NewReader(res.Body)
	csvR := csv.NewReader(r)

	headers, err := csvR.Read()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read headers")
	}

	headerToPostion := map[string]int{}
	for i, header := range headers {
		headerToPostion[header] = i
	}
	typePos := headerToPostion["typeID"]
	typeNamePos := headerToPostion["typeName"]
	volPosition := headerToPostion["volume"]
	iconPosition := headerToPostion["iconID"]

	invTypes := []models.EveInventoryType{}
	for {
		cols, err := csvR.Read()
		if err == io.EOF {
			break
		}

		ids := cols[typePos]
		id, err := strconv.ParseInt(ids, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse type id")
		}

		vols := cols[volPosition]
		vol, err := strconv.ParseFloat(vols, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse volume")
		}

		icons := cols[iconPosition]
		var icon *int64
		if icons != "None" {

			ic, err := strconv.ParseInt(icons, 10, 64)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse icon id")
			}
			icon = &ic
		}

		invTypes = append(invTypes, models.EveInventoryType{
			TypeID:   id,
			TypeName: cols[typeNamePos],
			Volume:   vol,
			IconID:   icon,
		})
	}

	return invTypes, nil
}

func (f *FuzzWorks) GetRegions(ctx context.Context) ([]models.Region, error) {
	res, err := f.client.Get(fmt.Sprintf("%s/mapRegions.csv.bz2", f.baseUrl))
	if err != nil {
		return nil, errors.Wrap(err, "failed to pull data from fuzzworks")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("pulling from fuzzworks failed, expected status code 200 got %d", res.StatusCode))
	}

	r := bzip2.NewReader(res.Body)
	csvR := csv.NewReader(r)

	headers, err := csvR.Read()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read headers")
	}

	headerToPostion := map[string]int{}
	for i, header := range headers {
		headerToPostion[header] = i
	}
	idPos := headerToPostion["regionID"]
	namePos := headerToPostion["regionName"]

	invTypes := []models.Region{}
	for {
		cols, err := csvR.Read()
		if err == io.EOF {
			break
		}

		ids := cols[idPos]
		id, err := strconv.ParseInt(ids, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse type id")
		}

		invTypes = append(invTypes, models.Region{
			ID:   id,
			Name: cols[namePos],
		})
	}

	return invTypes, nil
}

func (f *FuzzWorks) GetConstellations(ctx context.Context) ([]models.Constellation, error) {
	res, err := f.client.Get(fmt.Sprintf("%s/mapConstellations.csv.bz2", f.baseUrl))
	if err != nil {
		return nil, errors.Wrap(err, "failed to pull data from fuzzworks")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("pulling from fuzzworks failed, expected status code 200 got %d", res.StatusCode))
	}

	r := bzip2.NewReader(res.Body)
	csvR := csv.NewReader(r)

	headers, err := csvR.Read()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read headers")
	}

	headerToPostion := map[string]int{}
	for i, header := range headers {
		headerToPostion[header] = i
	}
	idPos := headerToPostion["constellationID"]
	namePos := headerToPostion["constellationName"]
	regionIDPos := headerToPostion["regionID"]

	invTypes := []models.Constellation{}
	for {
		cols, err := csvR.Read()
		if err == io.EOF {
			break
		}

		ids := cols[idPos]
		id, err := strconv.ParseInt(ids, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse type id")
		}

		rids := cols[regionIDPos]
		rid, err := strconv.ParseInt(rids, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse type id")
		}

		invTypes = append(invTypes, models.Constellation{
			ID:       id,
			Name:     cols[namePos],
			RegionID: rid,
		})
	}

	return invTypes, nil
}

func (f *FuzzWorks) GetSolarSystems(ctx context.Context) ([]models.SolarSystem, error) {
	res, err := f.client.Get(fmt.Sprintf("%s/mapSolarSystems.csv.bz2", f.baseUrl))
	if err != nil {
		return nil, errors.Wrap(err, "failed to pull data from fuzzworks")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("pulling from fuzzworks failed, expected status code 200 got %d", res.StatusCode))
	}

	r := bzip2.NewReader(res.Body)
	csvR := csv.NewReader(r)

	headers, err := csvR.Read()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read headers")
	}

	headerToPostion := map[string]int{}
	for i, header := range headers {
		headerToPostion[header] = i
	}
	idPos := headerToPostion["solarSystemID"]
	namePos := headerToPostion["solarSystemName"]
	otherIDPos := headerToPostion["constellationID"]

	invTypes := []models.SolarSystem{}
	for {
		cols, err := csvR.Read()
		if err == io.EOF {
			break
		}

		ids := cols[idPos]
		id, err := strconv.ParseInt(ids, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse type id")
		}

		rids := cols[otherIDPos]
		rid, err := strconv.ParseInt(rids, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse type id")
		}

		invTypes = append(invTypes, models.SolarSystem{
			ID:              id,
			Name:            cols[namePos],
			ConstellationID: rid,
		})
	}

	return invTypes, nil
}

func (f *FuzzWorks) GetNPCStations(ctx context.Context) ([]models.Station, error) {
	res, err := f.client.Get(fmt.Sprintf("%s/staStations.csv.bz2", f.baseUrl))
	if err != nil {
		return nil, errors.Wrap(err, "failed to pull data from fuzzworks")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("pulling from fuzzworks failed, expected status code 200 got %d", res.StatusCode))
	}

	r := bzip2.NewReader(res.Body)
	csvR := csv.NewReader(r)

	headers, err := csvR.Read()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read headers")
	}

	headerToPostion := map[string]int{}
	for i, header := range headers {
		headerToPostion[header] = i
	}
	idPos := headerToPostion["stationID"]
	namePos := headerToPostion["stationName"]
	otherIDPos := headerToPostion["solarSystemID"]
	otherOtherIDPos := headerToPostion["corporationID"]

	invTypes := []models.Station{}
	for {
		cols, err := csvR.Read()
		if err == io.EOF {
			break
		}

		ids := cols[idPos]
		id, err := strconv.ParseInt(ids, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse type id")
		}

		rids := cols[otherIDPos]
		rid, err := strconv.ParseInt(rids, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse type id")
		}

		crids := cols[otherOtherIDPos]
		crid, err := strconv.ParseInt(crids, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse type id")
		}

		invTypes = append(invTypes, models.Station{
			ID:            id,
			Name:          cols[namePos],
			SolarSystemID: rid,
			IsNPC:         true,
			CorporationID: crid,
		})
	}

	return invTypes, nil
}
