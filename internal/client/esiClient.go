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
	baseURL                    string
}

func NewEsiClient(clientID, clientSecret string) *EsiClient {
	return NewEsiClientWithHTTPClient(clientID, clientSecret, &http.Client{}, "")
}

func NewEsiClientWithHTTPClient(clientID, clientSecret string, httpClient HTTPDoer, baseURL string) *EsiClient {
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
	if baseURL == "" {
		baseURL = "https://esi.evetech.net"
	}

	return &EsiClient{
		oauthConfig:                oauthConfig,
		assetLocationFlagAllowList: assetLocationFlagAllowList,
		httpClient:                 httpClient,
		baseURL:                    baseURL,
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
	assets := []*models.EveAsset{}

	page := 1
	for {

		url, err := url.Parse(fmt.Sprintf("%s/characters/%d/assets?page=%d", c.baseURL, characterID, page))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse url")
		}

		req := &http.Request{
			Method: "GET",
			URL:    url,
			Header: c.getAuthHeaders(token),
		}

		res, err := c.httpClient.Do(req)
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

	// todo: handle more than 1000 locations because black omega things
	names := map[int64]string{}

	jsonIds, err := json.Marshal(ids)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal ids into json")
	}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/characters/%d/assets/names", c.baseURL, characterID),
		bytes.NewReader(jsonIds))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	req.Header = c.getAuthHeaders(token)

	res, err := c.httpClient.Do(req)
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

	stations := []models.Station{}
	for _, id := range ids {
		url, err := url.Parse(fmt.Sprintf("%s/universe/structures/%d", c.baseURL, id))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse url")
		}

		req := &http.Request{
			Method: "GET",
			URL:    url,
			Header: c.getAuthHeaders(token),
		}

		res, err := c.httpClient.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get player owned structure")
		}
		defer res.Body.Close()

		if res.StatusCode == 403 {
			// Player no longer has access — save as "Unknown Structure" so the
			// station row exists (the Upsert preserves any real name already stored).
			res.Body.Close()
			stations = append(stations, models.Station{
				ID:   id,
				Name: "Unknown Structure",
			})
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

type allianceInformation struct {
	Name string `json:"name"`
}

func (c *EsiClient) GetCharacterCorporation(ctx context.Context, characterID int64, token, refresh string, expire time.Time) (*models.Corporation, error) {
	jsons := fmt.Sprintf("[%d]", characterID)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/characters/affiliation", c.baseURL), bytes.NewBuffer([]byte(jsons)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	req.Header = c.getAuthHeaders(token)

	res, err := c.httpClient.Do(req)
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

	url, err := url.Parse(fmt.Sprintf("%s/corporations/%d", c.baseURL, charAffiliation[0].CorporationID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse url")
	}

	req = &http.Request{
		Method: "GET",
		URL:    url,
		Header: c.getAuthHeaders(token),
	}

	res, err = c.httpClient.Do(req)
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

	corp := &models.Corporation{
		ID:   charAffiliation[0].CorporationID,
		Name: info.Name,
	}

	// Fetch alliance info if character is in an alliance
	if charAffiliation[0].AllianceID > 0 {
		corp.AllianceID = charAffiliation[0].AllianceID

		allianceURL, err := url.Parse(fmt.Sprintf("%s/alliances/%d", c.baseURL, charAffiliation[0].AllianceID))
		if err == nil {
			allianceReq := &http.Request{
				Method: "GET",
				URL:    allianceURL,
				Header: c.getAuthHeaders(token),
			}
			allianceRes, err := c.httpClient.Do(allianceReq)
			if err == nil {
				defer allianceRes.Body.Close()
				if allianceRes.StatusCode == 200 {
					var allianceInfo allianceInformation
					allianceBody, err := io.ReadAll(allianceRes.Body)
					if err == nil {
						if json.Unmarshal(allianceBody, &allianceInfo) == nil {
							corp.AllianceName = allianceInfo.Name
						}
					}
				}
			}
		}
	}

	return corp, nil
}

func (c *EsiClient) GetCorporationAssets(ctx context.Context, corpID int64, token, refresh string, expire time.Time) ([]*models.EveAsset, error) {
	assets := []*models.EveAsset{}

	page := 1
	for {

		url, err := url.Parse(fmt.Sprintf("%s/corporations/%d/assets?page=%d", c.baseURL, corpID, page))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse url")
		}

		req := &http.Request{
			Method: "GET",
			URL:    url,
			Header: c.getAuthHeaders(token),
		}

		res, err := c.httpClient.Do(req)
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

	// todo: handle more than 1000 locations because black omega things
	names := map[int64]string{}

	jsonIds, err := json.Marshal(ids)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal ids into json")
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/corporations/%d/assets/names", c.baseURL, corpID), bytes.NewReader(jsonIds))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	req.Header = c.getAuthHeaders(token)

	res, err := c.httpClient.Do(req)
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
	url, err := url.Parse(fmt.Sprintf("%s/corporations/%d/divisions", c.baseURL, corpID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse url")
	}

	req := &http.Request{
		Method: "GET",
		URL:    url,
		Header: c.getAuthHeaders(token),
	}

	res, err := c.httpClient.Do(req)
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
	orders := []*MarketOrder{}

	page := 1
	for {
		url, err := url.Parse(fmt.Sprintf("%s/latest/markets/%d/orders/?page=%d", c.baseURL, regionID, page))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse market orders url")
		}

		req := &http.Request{
			Method: "GET",
			URL:    url,
			Header: c.getCommonHeaders(),
		}

		res, err := c.httpClient.Do(req)
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

func (c *EsiClient) getAuthHeaders(token string) http.Header {
	headers := c.getCommonHeaders()
	headers.Set("Authorization", "Bearer "+token)
	return headers
}

// CcpMarketPrice represents a market price from the CCP ESI /markets/prices/ endpoint
type CcpMarketPrice struct {
	TypeID        int64    `json:"type_id"`
	AdjustedPrice *float64 `json:"adjusted_price"`
	AveragePrice  *float64 `json:"average_price"`
}

// GetCcpMarketPrices fetches adjusted/average market prices from ESI (public, no auth required)
func (c *EsiClient) GetCcpMarketPrices(ctx context.Context) ([]*CcpMarketPrice, error) {
	reqURL := fmt.Sprintf("%s/latest/markets/prices/", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create CCP market prices request")
	}
	req.Header = c.getCommonHeaders()

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch CCP market prices")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code fetching CCP market prices: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read CCP market prices body")
	}

	var prices []*CcpMarketPrice
	if err := json.Unmarshal(body, &prices); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal CCP market prices")
	}

	return prices, nil
}

// IndustryCostIndexActivity represents a single activity cost index
type IndustryCostIndexActivity struct {
	Activity  string  `json:"activity"`
	CostIndex float64 `json:"cost_index"`
}

// IndustryCostIndexSystem represents cost indices for a solar system
type IndustryCostIndexSystem struct {
	SolarSystemID int64                       `json:"solar_system_id"`
	CostIndices   []IndustryCostIndexActivity `json:"cost_indices"`
}

// GetIndustryCostIndices fetches industry cost indices from ESI (public, no auth required)
func (c *EsiClient) GetIndustryCostIndices(ctx context.Context) ([]*IndustryCostIndexSystem, error) {
	reqURL := fmt.Sprintf("%s/latest/industry/systems/", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create industry cost indices request")
	}
	req.Header = c.getCommonHeaders()

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch industry cost indices")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code fetching industry cost indices: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read industry cost indices body")
	}

	var systems []*IndustryCostIndexSystem
	if err := json.Unmarshal(body, &systems); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal industry cost indices")
	}

	return systems, nil
}

// universeNameEntry represents a single entry from the ESI /universe/names/ response
type universeNameEntry struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

// GetUniverseNames resolves IDs to names using the public ESI /universe/names/ endpoint.
// Accepts up to 1000 IDs per call; larger sets are batched automatically.
func (c *EsiClient) GetUniverseNames(ctx context.Context, ids []int64) (map[int64]string, error) {
	if len(ids) == 0 {
		return map[int64]string{}, nil
	}

	result := make(map[int64]string, len(ids))
	batchSize := 1000

	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		batch := ids[i:end]

		body, err := json.Marshal(batch)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal universe names request")
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/universe/names/", c.baseURL), bytes.NewReader(body))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create universe names request")
		}
		req.Header = c.getCommonHeaders()

		res, err := c.httpClient.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch universe names")
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(res.Body)
			return nil, fmt.Errorf("unexpected status code fetching universe names: %d, %s", res.StatusCode, respBody)
		}

		respBody, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read universe names response")
		}

		var entries []universeNameEntry
		if err := json.Unmarshal(respBody, &entries); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal universe names response")
		}

		for _, entry := range entries {
			result[entry.ID] = entry.Name
		}
	}

	return result, nil
}

// PI ESI response types

type EsiPiPlanet struct {
	LastUpdate    string `json:"last_update"`
	NumPins       int    `json:"num_pins"`
	OwnerID       int64  `json:"owner_id"`
	PlanetID      int64  `json:"planet_id"`
	PlanetType    string `json:"planet_type"`
	SolarSystemID int64  `json:"solar_system_id"`
	UpgradeLevel  int    `json:"upgrade_level"`
}

type EsiPiColony struct {
	Links  []EsiPiLink  `json:"links"`
	Pins   []EsiPiPin   `json:"pins"`
	Routes []EsiPiRoute `json:"routes"`
}

type EsiPiLink struct {
	SourcePinID      int64 `json:"source_pin_id"`
	DestinationPinID int64 `json:"destination_pin_id"`
	LinkLevel        int   `json:"link_level"`
}

type EsiPiPin struct {
	PinID            int64               `json:"pin_id"`
	TypeID           int64               `json:"type_id"`
	Latitude         float64             `json:"latitude"`
	Longitude        float64             `json:"longitude"`
	InstallTime      *string             `json:"install_time"`
	ExpiryTime       *string             `json:"expiry_time"`
	LastCycleStart   *string             `json:"last_cycle_start"`
	SchematicID      *int                `json:"schematic_id"`
	Contents         []EsiPiPinContent   `json:"contents"`
	ExtractorDetails *EsiExtractorDetail `json:"extractor_details"`
	FactoryDetails   *EsiFactoryDetail   `json:"factory_details"`
}

type EsiPiPinContent struct {
	Amount float64 `json:"amount"`
	TypeID int64   `json:"type_id"`
}

type EsiExtractorDetail struct {
	CycleTime     int     `json:"cycle_time"`
	HeadRadius    float64 `json:"head_radius"`
	ProductTypeID int64   `json:"product_type_id"`
	QtyPerCycle   int     `json:"qty_per_cycle"`
	Heads         []struct {
		HeadID    int     `json:"head_id"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"heads"`
}

type EsiFactoryDetail struct {
	SchematicID int `json:"schematic_id"`
}

type EsiPiRoute struct {
	RouteID          int64   `json:"route_id"`
	SourcePinID      int64   `json:"source_pin_id"`
	DestinationPinID int64   `json:"destination_pin_id"`
	ContentTypeID    int64   `json:"content_type_id"`
	Quantity         float64 `json:"quantity"`
	Waypoints        []int64 `json:"waypoints"`
}

// GetCharacterPlanets fetches the list of PI colonies for a character.
func (c *EsiClient) GetCharacterPlanets(ctx context.Context, characterID int64, token string) ([]*EsiPiPlanet, error) {
	reqURL := fmt.Sprintf("%s/v1/characters/%d/planets/", c.baseURL, characterID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create character planets request")
	}
	req.Header = c.getAuthHeaders(token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch character planets")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errText, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed to get character planets, expected 200 got %d, %s", res.StatusCode, errText)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read character planets body")
	}

	var planets []*EsiPiPlanet
	if err := json.Unmarshal(body, &planets); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal character planets")
	}

	return planets, nil
}

// GetCharacterPlanetDetails fetches the full colony layout for a single planet.
func (c *EsiClient) GetCharacterPlanetDetails(ctx context.Context, characterID, planetID int64, token string) (*EsiPiColony, error) {
	reqURL := fmt.Sprintf("%s/v3/characters/%d/planets/%d/", c.baseURL, characterID, planetID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create planet details request")
	}
	req.Header = c.getAuthHeaders(token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch planet details")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errText, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed to get planet details, expected 200 got %d, %s", res.StatusCode, errText)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read planet details body")
	}

	var colony EsiPiColony
	if err := json.Unmarshal(body, &colony); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal planet details")
	}

	return &colony, nil
}

// RefreshedToken holds the result of a token refresh.
type RefreshedToken struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

// EsiContract represents a contract from the ESI contracts endpoint.
type EsiContract struct {
	ContractID     int64   `json:"contract_id"`
	IssuerID       int64   `json:"issuer_id"`
	AcceptorID     int64   `json:"acceptor_id"`
	AssigneeID     int64   `json:"assignee_id"`
	Type           string  `json:"type"`
	Status         string  `json:"status"`
	Title          string  `json:"title"`
	DateCompleted  string  `json:"date_completed"`
	DateExpired    string  `json:"date_expired"`
	ForCorporation bool    `json:"for_corporation"`
	Price          float64 `json:"price"`
}

// GetCharacterContracts fetches all contracts for a character from ESI.
func (c *EsiClient) GetCharacterContracts(ctx context.Context, characterID int64, token, refresh string, expire time.Time) ([]*EsiContract, error) {
	contracts := []*EsiContract{}

	page := 1
	for {
		url, err := url.Parse(fmt.Sprintf("%s/v1/characters/%d/contracts/?page=%d", c.baseURL, characterID, page))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse url")
		}

		req := &http.Request{
			Method: "GET",
			URL:    url,
			Header: c.getAuthHeaders(token),
		}

		res, err := c.httpClient.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get character contracts")
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			errText, _ := io.ReadAll(res.Body)
			return nil, errors.New(fmt.Sprintf("failed to get character contracts, expected statusCode 200 got %d, %s", res.StatusCode, errText))
		}

		totalPagesStr := res.Header.Get("X-PAGES")
		totalPages, err := strconv.Atoi(totalPagesStr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse x-pages")
		}

		bytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read response body")
		}

		pageContracts := []*EsiContract{}
		err = json.Unmarshal(bytes, &pageContracts)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal contracts")
		}

		contracts = append(contracts, pageContracts...)

		if totalPages == page {
			return contracts, nil
		}

		page++
	}
}

// GetCorporationContracts fetches all contracts for a corporation from ESI.
func (c *EsiClient) GetCorporationContracts(ctx context.Context, corporationID int64, token, refresh string, expire time.Time) ([]*EsiContract, error) {
	contracts := []*EsiContract{}

	page := 1
	for {
		url, err := url.Parse(fmt.Sprintf("%s/v1/corporations/%d/contracts/?page=%d", c.baseURL, corporationID, page))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse url")
		}

		req := &http.Request{
			Method: "GET",
			URL:    url,
			Header: c.getAuthHeaders(token),
		}

		res, err := c.httpClient.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get corporation contracts")
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			errText, _ := io.ReadAll(res.Body)
			return nil, errors.New(fmt.Sprintf("failed to get corporation contracts, expected statusCode 200 got %d, %s", res.StatusCode, errText))
		}

		totalPagesStr := res.Header.Get("X-PAGES")
		totalPages, err := strconv.Atoi(totalPagesStr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse x-pages")
		}

		bytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read response body")
		}

		pageContracts := []*EsiContract{}
		err = json.Unmarshal(bytes, &pageContracts)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal contracts")
		}

		contracts = append(contracts, pageContracts...)

		if totalPages == page {
			return contracts, nil
		}

		page++
	}
}

// RefreshAccessToken uses the refresh token to obtain a new access token from EVE SSO.
// Returns the new access token, refresh token, and expiry. The caller is responsible
// for persisting these back to the database.
func (c *EsiClient) RefreshAccessToken(ctx context.Context, refreshToken string) (*RefreshedToken, error) {
	t := &oauth2.Token{
		RefreshToken: refreshToken,
	}
	src := c.oauthConfig.TokenSource(ctx, t)
	newToken, err := src.Token()
	if err != nil {
		return nil, errors.Wrap(err, "failed to refresh ESI access token")
	}
	return &RefreshedToken{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		Expiry:       newToken.Expiry,
	}, nil
}
