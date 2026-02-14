package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

//go:generate mockgen -source=./esiClient.go -destination=./esiClient_mock_test.go -package=client_test

// HTTPDoer interface for making HTTP requests (allows mocking)
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type EsiClient struct {
	oauthConfig                *oauth2.Config
	assetLocationFlagAllowList []string
	httpClient                 HTTPDoer
}

func NewEsiClient(clientID, clientSecret string) *EsiClient {
	return NewEsiClientWithHTTPClient(clientID, clientSecret, nil)
}

func NewEsiClientWithHTTPClient(clientID, clientSecret string, httpClient HTTPDoer) *EsiClient {
	endpoint := oauth2.Endpoint{
		AuthURL:  "https://login.eveonline.com/v2/oauth/authorize",
		TokenURL: "https://login.eveonline.com/v2/oauth/token",
	}
	oauthConfig := &oauth2.Config{
		Endpoint:     endpoint,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
	assetLocationFlagAllowList := []string{
		"Hangar",
		"Unlocked",
		"Cargo",
		"CapsuleerDeliveries",
		"CorpDeliveries",
		"Deliveries",
		"ExpeditionHold",
		"HangarAll",
		"InfrastructureHangar",
		"Locked",
		"MoonMaterialBay",
		"SpecializedAsteroidHold",
		"AssetSafety",
		"CorporationGoalDeliveries",
		"CorpSAG1",
		"CorpSAG2",
		"CorpSAG3",
		"CorpSAG4",
		"CorpSAG5",
		"CorpSAG6",
		"CorpSAG7",
		"OfficeFolder",
	}
	return &EsiClient{
		oauthConfig:                oauthConfig,
		assetLocationFlagAllowList: assetLocationFlagAllowList,
		httpClient:                 httpClient,
	}
}

// MarketOrder represents a market order from ESI
type MarketOrder struct {
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

func (c *EsiClient) GetCharacterAssets(ctx context.Context, characterID int64, token, refresh string, expire time.Time) ([]*models.EveAsset, error) {
	var client HTTPDoer
	if c.httpClient != nil {
		client = c.httpClient
	} else {
		t := new(oauth2.Token)
		t.AccessToken = token
		t.RefreshToken = refresh
		t.TokenType = ""
		t.Expiry = expire
		client = c.oauthConfig.Client(ctx, t)
	}

	assets := []*models.EveAsset{}

	page := 1
	for {

		url, err := url.Parse(fmt.Sprintf("https://esi.evetech.net/characters/%d/assets?page=%d", characterID, page))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse url")
		}

		req := &http.Request{
			Method: "GET",
			URL:    url,
			Header: c.getCommonHeaders(),
		}

		res, err := client.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get character assets")
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			errText, _ := io.ReadAll(res.Body)
			return nil, errors.New(fmt.Sprintf("failed get character assets, expected statusCode 200 got %d, %s", res.StatusCode, errText))
		}

		totalPagess := res.Header.Get("X-PAGES")
		totalPages, err := strconv.Atoi(totalPagess)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse x-pages")
		}

		bytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read response body")
		}

		moreAssets := []*models.EveAsset{}
		err = json.Unmarshal(bytes, &moreAssets)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal data")
		}

		for _, asset := range moreAssets {
			allowed := slices.Contains(c.assetLocationFlagAllowList, asset.LocationFlag)
			if !allowed {
				continue
			}
			assets = append(assets, asset)
		}

		if totalPages == page {
			return assets, nil
		}

		page++
	}
}

type nameResponse struct {
	ItemID int64  `json:"item_id"`
	Name   string `json:"name"`
}

func (c *EsiClient) GetCharacterLocationNames(ctx context.Context, characterID int64, token, refresh string, expire time.Time, ids []int64) (map[int64]string, error) {
	if len(ids) == 0 {
		return map[int64]string{}, nil
	}

	var client HTTPDoer
	if c.httpClient != nil {
		client = c.httpClient
	} else {
		t := new(oauth2.Token)
		t.AccessToken = token
		t.RefreshToken = refresh
		t.TokenType = ""
		t.Expiry = expire
		client = c.oauthConfig.Client(ctx, t)
	}

	// todo: handle more than 1000 locations because black omega things
	names := map[int64]string{}

	jsonIds, err := json.Marshal(ids)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal ids into json")
	}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("https://esi.evetech.net/characters/%d/assets/names", characterID),
		bytes.NewReader(jsonIds))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	req.Header = c.getCommonHeaders()

	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get character location names")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		errText, _ := io.ReadAll(res.Body)
		return nil, errors.New(fmt.Sprintf("failed get character location names, expected statusCode 200 got %d, %s", res.StatusCode, errText))
	}

	nameJSON := []nameResponse{}
	j, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read names body")
	}
	err = json.Unmarshal(j, &nameJSON)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal name json")
	}

	for _, name := range nameJSON {
		names[name.ItemID] = name.Name
	}

	return names, nil
}

type playerOwnedStructure struct {
	Name          string `json:"name"`
	OwnerID       int64  `json:"owner_id"`
	SolarSystemID int64  `json:"solar_system_id"`
}

func (c *EsiClient) GetPlayerOwnedStationInformation(ctx context.Context, token, refresh string, expire time.Time, ids []int64) ([]models.Station, error) {
	if len(ids) == 0 {
		return []models.Station{}, nil
	}

	var client HTTPDoer
	if c.httpClient != nil {
		client = c.httpClient
	} else {
		t := new(oauth2.Token)
		t.AccessToken = token
		t.RefreshToken = refresh
		t.TokenType = ""
		t.Expiry = expire
		client = c.oauthConfig.Client(ctx, t)
	}

	stations := []models.Station{}
	for _, id := range ids {
		url, err := url.Parse(fmt.Sprintf("https://esi.evetech.net/universe/structures/%d", id))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse url")
		}

		req := &http.Request{
			Method: "GET",
			URL:    url,
			Header: c.getCommonHeaders(),
		}

		res, err := client.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get player owned structure")
		}
		defer res.Body.Close()

		if res.StatusCode == 403 {
			// player does not have access to query this structure so just continue
			continue
		}

		if res.StatusCode != 200 {
			errText, _ := io.ReadAll(res.Body)
			return nil, errors.New(fmt.Sprintf("failed get player owned structure, expected statusCode 200 got %d, %s", res.StatusCode, errText))
		}

		structure := playerOwnedStructure{}
		j, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read names body")
		}
		err = json.Unmarshal(j, &structure)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal name json")
		}

		stations = append(stations, models.Station{
			ID:            id,
			Name:          structure.Name,
			SolarSystemID: structure.SolarSystemID,
			CorporationID: structure.OwnerID,
			IsNPC:         false,
		})

	}

	return stations, nil
}

type characterAffiliation struct {
	CorporationID int64 `json:"corporation_id"`
	FactionID     int64 `json:"faction_id"`
	CharacterID   int64 `json:"character_id"`
	AllianceID    int64 `json:"alliance_id"`
}

type corporationInformation struct {
	Name string `json:"name"`
}

func (c *EsiClient) GetCharacterCorporation(ctx context.Context, characterID int64, token, refresh string, expire time.Time) (*models.Corporation, error) {
	var client HTTPDoer
	if c.httpClient != nil {
		client = c.httpClient
	} else {
		t := new(oauth2.Token)
		t.AccessToken = token
		t.RefreshToken = refresh
		t.TokenType = ""
		t.Expiry = expire
		client = c.oauthConfig.Client(ctx, t)
	}

	jsons := fmt.Sprintf("[%d]", characterID)

	req, err := http.NewRequest("POST", "https://esi.evetech.net/characters/affiliation", bytes.NewBuffer([]byte(jsons)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	req.Header = c.getCommonHeaders()

	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do character affiliation")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		errText, _ := io.ReadAll(res.Body)
		return nil, errors.New(fmt.Sprintf("failed to get character affiliation, expected statusCode 200 got %d, %s", res.StatusCode, errText))
	}

	charAffiliation := []characterAffiliation{}
	j, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read character affiliation body")
	}
	err = json.Unmarshal(j, &charAffiliation)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal character affiliation json")
	}

	url, err := url.Parse(fmt.Sprintf("https://esi.evetech.net/corporations/%d", charAffiliation[0].CorporationID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse url")
	}

	req = &http.Request{
		Method: "GET",
		URL:    url,
		Header: c.getCommonHeaders(),
	}

	res, err = client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do corporation information")
	}
	defer res.Body.Close()

	var info corporationInformation
	j, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read corporation information")
	}
	err = json.Unmarshal(j, &info)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal corporation information")
	}

	return &models.Corporation{
		ID:   charAffiliation[0].CorporationID,
		Name: info.Name,
	}, nil
}

func (c *EsiClient) GetCorporationAssets(ctx context.Context, corpID int64, token, refresh string, expire time.Time) ([]*models.EveAsset, error) {
	var client HTTPDoer
	if c.httpClient != nil {
		client = c.httpClient
	} else {
		t := new(oauth2.Token)
		t.AccessToken = token
		t.RefreshToken = refresh
		t.TokenType = ""
		t.Expiry = expire
		client = c.oauthConfig.Client(ctx, t)
	}

	assets := []*models.EveAsset{}

	page := 1
	for {

		url, err := url.Parse(fmt.Sprintf("https://esi.evetech.net/corporations/%d/assets?page=%d", corpID, page))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse url")
		}

		req := &http.Request{
			Method: "GET",
			URL:    url,
			Header: c.getCommonHeaders(),
		}

		res, err := client.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get corporation assets")
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			errText, _ := io.ReadAll(res.Body)
			return nil, errors.New(fmt.Sprintf("failed get corporation assets, expected statusCode 200 got %d, %s", res.StatusCode, errText))
		}

		totalPagess := res.Header.Get("X-PAGES")
		totalPages, err := strconv.Atoi(totalPagess)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse x-pages")
		}

		bytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read response body")
		}

		moreAssets := []*models.EveAsset{}
		err = json.Unmarshal(bytes, &moreAssets)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal data")
		}

		for _, asset := range moreAssets {
			allowed := slices.Contains(c.assetLocationFlagAllowList, asset.LocationFlag)
			if !allowed {
				continue
			}
			assets = append(assets, asset)
		}

		if totalPages == page {
			return assets, nil
		}

		page++
	}
}

func (c *EsiClient) GetCorporationLocationNames(ctx context.Context, corpID int64, token, refresh string, expire time.Time, ids []int64) (map[int64]string, error) {
	if len(ids) == 0 {
		return map[int64]string{}, nil
	}

	var client HTTPDoer
	if c.httpClient != nil {
		client = c.httpClient
	} else {
		t := new(oauth2.Token)
		t.AccessToken = token
		t.RefreshToken = refresh
		t.TokenType = ""
		t.Expiry = expire
		client = c.oauthConfig.Client(ctx, t)
	}

	// todo: handle more than 1000 locations because black omega things
	names := map[int64]string{}

	jsonIds, err := json.Marshal(ids)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal ids into json")
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("https://esi.evetech.net/corporations/%d/assets/names", corpID), bytes.NewReader(jsonIds))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	req.Header = c.getCommonHeaders()

	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get corporation location names")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		errText, _ := io.ReadAll(res.Body)
		return nil, errors.New(fmt.Sprintf("failed get corporation location names, expected statusCode 200 got %d, %s", res.StatusCode, errText))
	}

	nameJSON := []nameResponse{}
	j, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read names body")
	}
	err = json.Unmarshal(j, &nameJSON)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal name json")
	}

	for _, name := range nameJSON {
		names[name.ItemID] = name.Name
	}

	return names, nil
}

type divisionResponse struct {
	Hangar []struct {
		Division int    `json:"division"`
		Name     string `json:"name"`
	} `json:"hangar"`
	Wallet []struct {
		Division int    `json:"division"`
		Name     string `json:"name"`
	} `json:"wallet"`
}

func (c *EsiClient) GetCorporationDivisions(ctx context.Context, corpID int64, token, refresh string, expire time.Time) (*models.CorporationDivisions, error) {
	var client HTTPDoer
	if c.httpClient != nil {
		client = c.httpClient
	} else {
		t := new(oauth2.Token)
		t.AccessToken = token
		t.RefreshToken = refresh
		t.TokenType = ""
		t.Expiry = expire
		client = c.oauthConfig.Client(ctx, t)
	}

	url, err := url.Parse(fmt.Sprintf("https://esi.evetech.net/corporations/%d/divisions", corpID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse url")
	}

	req := &http.Request{
		Method: "GET",
		URL:    url,
		Header: c.getCommonHeaders(),
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get corporation divisions")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		errText, _ := io.ReadAll(res.Body)
		return nil, errors.New(fmt.Sprintf("failed get corporation divisions, expected statusCode 200 got %d, %s", res.StatusCode, errText))
	}

	var divisions divisionResponse
	j, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read corporation divisions body")
	}
	err = json.Unmarshal(j, &divisions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal corporation divisions json")
	}

	hangar := map[int]string{}
	for _, h := range divisions.Hangar {
		hangar[h.Division] = h.Name
	}

	wallet := map[int]string{}
	for _, w := range divisions.Wallet {
		wallet[w.Division] = w.Name
	}

	return &models.CorporationDivisions{
		Hanger: hangar,
		Wallet: wallet,
	}, nil
}

func (c *EsiClient) GetMarketOrders(ctx context.Context, regionID int64) ([]*MarketOrder, error) {
	var client HTTPDoer
	if c.httpClient != nil {
		client = c.httpClient
	} else {
		client = &http.Client{}
	}

	orders := []*MarketOrder{}

	page := 1
	for {
		url, err := url.Parse(fmt.Sprintf("https://esi.evetech.net/latest/markets/%d/orders/?page=%d", regionID, page))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse market orders url")
		}

		req := &http.Request{
			Method: "GET",
			URL:    url,
			Header: c.getCommonHeaders(),
		}

		res, err := client.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get market orders")
		}

		if res.StatusCode != 200 {
			errText, _ := io.ReadAll(res.Body)
			res.Body.Close()
			return nil, errors.New(fmt.Sprintf("failed to get market orders, expected statusCode 200 got %d, %s", res.StatusCode, errText))
		}

		var pageOrders []*MarketOrder
		j, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return nil, errors.Wrap(err, "failed to read market orders body")
		}

		err = json.Unmarshal(j, &pageOrders)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal market orders json")
		}

		orders = append(orders, pageOrders...)

		pages := res.Header.Get("X-Pages")
		if pages == "" {
			break
		}

		totalPages, err := strconv.Atoi(pages)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse X-Pages header")
		}

		if page >= totalPages {
			break
		}

		page++
	}

	return orders, nil
}

func (c *EsiClient) getCommonHeaders() http.Header {
	headers := http.Header{}
	headers.Add("X-Compatibility-Date", "2025-12-16")
	headers.Add("Accept", "application/json")
	headers.Add("Content-Type", "application/json")
	return headers
}
