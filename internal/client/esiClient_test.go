package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_ClientShouldGetCharacterAssets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := NewMockHTTPDoer(ctrl)

	// Create mock asset data
	mockAssets := []*models.EveAsset{
		{
			ItemID:       123456,
			TypeID:       34,
			Quantity:     1000,
			LocationID:   60003760,
			LocationFlag: "Hangar",
		},
		{
			ItemID:       123457,
			TypeID:       35,
			Quantity:     500,
			LocationID:   60003760,
			LocationFlag: "Hangar",
		},
	}

	// Marshal to JSON
	assetsJSON, _ := json.Marshal(mockAssets)

	// Create mock HTTP response
	mockResponse := &http.Response{
		StatusCode: 200,
		Header: http.Header{
			"X-Pages": []string{"1"},
		},
		Body: io.NopCloser(bytes.NewReader(assetsJSON)),
	}

	// Expect the HTTP call
	mockHTTPClient.EXPECT().
		Do(gomock.Any()).
		Return(mockResponse, nil).
		Times(1)

	// Create ESI client with mock HTTP client
	esiClient := client.NewEsiClientWithHTTPClient("test-client-id", "test-client-secret", mockHTTPClient, "")

	// Test GetCharacterAssets
	assets, err := esiClient.GetCharacterAssets(context.Background(), 12345, "test-token", "test-refresh", time.Now())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(assets))
	assert.Equal(t, int64(34), assets[0].TypeID)
	assert.Equal(t, int64(1000), assets[0].Quantity)
}

func Test_ClientShouldGetMarketOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := NewMockHTTPDoer(ctrl)

	// Create mock market orders
	mockOrders := []*client.MarketOrder{
		{
			OrderID:      9876543,
			TypeID:       34,
			LocationID:   60003760,
			VolumeTotal:  10000,
			VolumeRemain: 5000,
			Price:        5.50,
			IsBuyOrder:   false,
		},
		{
			OrderID:      9876544,
			TypeID:       34,
			LocationID:   60003760,
			VolumeTotal:  8000,
			VolumeRemain: 8000,
			Price:        5.45,
			IsBuyOrder:   true,
		},
	}

	// Marshal to JSON
	ordersJSON, _ := json.Marshal(mockOrders)

	// Create mock HTTP response
	mockResponse := &http.Response{
		StatusCode: 200,
		Header: http.Header{
			"X-Pages": []string{"1"},
		},
		Body: io.NopCloser(bytes.NewReader(ordersJSON)),
	}

	// Expect the HTTP call
	mockHTTPClient.EXPECT().
		Do(gomock.Any()).
		Return(mockResponse, nil).
		Times(1)

	// Create ESI client with mock HTTP client
	esiClient := client.NewEsiClientWithHTTPClient("test-client-id", "test-client-secret", mockHTTPClient, "")

	// Test GetMarketOrders
	orders, err := esiClient.GetMarketOrders(context.Background(), 10000002)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(orders))
	assert.Equal(t, int64(34), orders[0].TypeID)
	assert.Equal(t, 5.50, orders[0].Price)
	assert.False(t, orders[0].IsBuyOrder)
	assert.Equal(t, 5.45, orders[1].Price)
	assert.True(t, orders[1].IsBuyOrder)
}

func Test_ClientShouldGetUniverseNames(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := NewMockHTTPDoer(ctrl)

	mockNames := []struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		Category string `json:"category"`
	}{
		{ID: 60003760, Name: "Jita IV - Moon 4 - Caldari Navy Assembly Plant", Category: "station"},
		{ID: 60008494, Name: "Amarr VIII (Oris) - Emperor Family Academy", Category: "station"},
	}

	namesJSON, _ := json.Marshal(mockNames)

	mockHTTPClient.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "POST", req.Method)
			assert.Contains(t, req.URL.String(), "/universe/names/")

			var ids []int64
			body, _ := io.ReadAll(req.Body)
			json.Unmarshal(body, &ids)
			assert.ElementsMatch(t, []int64{60003760, 60008494}, ids)

			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(namesJSON)),
			}, nil
		}).
		Times(1)

	esiClient := client.NewEsiClientWithHTTPClient("test-client-id", "test-client-secret", mockHTTPClient, "https://esi.test.com")

	names, err := esiClient.GetUniverseNames(context.Background(), []int64{60003760, 60008494})
	assert.NoError(t, err)
	assert.Len(t, names, 2)
	assert.Equal(t, "Jita IV - Moon 4 - Caldari Navy Assembly Plant", names[60003760])
	assert.Equal(t, "Amarr VIII (Oris) - Emperor Family Academy", names[60008494])
}

func Test_ClientShouldReturnEmptyMapForNoIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := NewMockHTTPDoer(ctrl)

	esiClient := client.NewEsiClientWithHTTPClient("test-client-id", "test-client-secret", mockHTTPClient, "")

	names, err := esiClient.GetUniverseNames(context.Background(), []int64{})
	assert.NoError(t, err)
	assert.Empty(t, names)
}

func Test_ClientShouldHandleUniverseNamesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := NewMockHTTPDoer(ctrl)

	mockHTTPClient.EXPECT().
		Do(gomock.Any()).
		Return(&http.Response{
			StatusCode: 500,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"internal error"}`))),
		}, nil).
		Times(1)

	esiClient := client.NewEsiClientWithHTTPClient("test-client-id", "test-client-secret", mockHTTPClient, "https://esi.test.com")

	names, err := esiClient.GetUniverseNames(context.Background(), []int64{60003760})
	assert.Error(t, err)
	assert.Nil(t, names)
	assert.Contains(t, err.Error(), "unexpected status code")
}

func Test_ClientShouldGetCharacterContracts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := NewMockHTTPDoer(ctrl)

	mockContracts := []*client.EsiContract{
		{
			ContractID:     12345,
			IssuerID:       100,
			AcceptorID:     200,
			AssigneeID:     200,
			Type:           "item_exchange",
			Status:         "finished",
			Title:          "Minerals for PT-42",
			DateCompleted:  "2025-01-15T12:00:00Z",
			DateExpired:    "2025-02-15T12:00:00Z",
			ForCorporation: false,
			Price:          1000000.0,
		},
		{
			ContractID:     12346,
			IssuerID:       100,
			AcceptorID:     300,
			AssigneeID:     300,
			Type:           "courier",
			Status:         "outstanding",
			Title:          "Delivery",
			ForCorporation: false,
			Price:          50000.0,
		},
	}

	contractsJSON, _ := json.Marshal(mockContracts)

	mockResponse := &http.Response{
		StatusCode: 200,
		Header: http.Header{
			"X-Pages": []string{"1"},
		},
		Body: io.NopCloser(bytes.NewReader(contractsJSON)),
	}

	mockHTTPClient.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "GET", req.Method)
			assert.Contains(t, req.URL.String(), "/v1/characters/12345/contracts/")
			return mockResponse, nil
		}).
		Times(1)

	esiClient := client.NewEsiClientWithHTTPClient("test-client-id", "test-client-secret", mockHTTPClient, "https://esi.test.com")

	contracts, err := esiClient.GetCharacterContracts(context.Background(), 12345, "test-token", "test-refresh", time.Now())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(contracts))
	assert.Equal(t, int64(12345), contracts[0].ContractID)
	assert.Equal(t, "item_exchange", contracts[0].Type)
	assert.Equal(t, "finished", contracts[0].Status)
	assert.Equal(t, "Minerals for PT-42", contracts[0].Title)
	assert.Equal(t, 1000000.0, contracts[0].Price)
	assert.Equal(t, int64(12346), contracts[1].ContractID)
	assert.Equal(t, "courier", contracts[1].Type)
	assert.Equal(t, "outstanding", contracts[1].Status)
}

func Test_ClientShouldGetCharacterContractsMultiplePages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := NewMockHTTPDoer(ctrl)

	page1Contracts := []*client.EsiContract{
		{ContractID: 1001, Type: "item_exchange", Status: "finished", Title: "Page 1 contract"},
	}
	page2Contracts := []*client.EsiContract{
		{ContractID: 2001, Type: "item_exchange", Status: "outstanding", Title: "Page 2 contract"},
	}

	page1JSON, _ := json.Marshal(page1Contracts)
	page2JSON, _ := json.Marshal(page2Contracts)

	callCount := 0
	mockHTTPClient.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				return &http.Response{
					StatusCode: 200,
					Header:     http.Header{"X-Pages": []string{"2"}},
					Body:       io.NopCloser(bytes.NewReader(page1JSON)),
				}, nil
			}
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"X-Pages": []string{"2"}},
				Body:       io.NopCloser(bytes.NewReader(page2JSON)),
			}, nil
		}).
		Times(2)

	esiClient := client.NewEsiClientWithHTTPClient("test-client-id", "test-client-secret", mockHTTPClient, "https://esi.test.com")

	contracts, err := esiClient.GetCharacterContracts(context.Background(), 99999, "tok", "ref", time.Now())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(contracts))
	assert.Equal(t, int64(1001), contracts[0].ContractID)
	assert.Equal(t, int64(2001), contracts[1].ContractID)
}

func Test_ClientShouldHandleCharacterContractsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := NewMockHTTPDoer(ctrl)

	mockHTTPClient.EXPECT().
		Do(gomock.Any()).
		Return(&http.Response{
			StatusCode: 403,
			Header:     http.Header{"X-Pages": []string{"1"}},
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"forbidden"}`))),
		}, nil).
		Times(1)

	esiClient := client.NewEsiClientWithHTTPClient("test-client-id", "test-client-secret", mockHTTPClient, "https://esi.test.com")

	contracts, err := esiClient.GetCharacterContracts(context.Background(), 12345, "bad-token", "ref", time.Now())
	assert.Error(t, err)
	assert.Nil(t, contracts)
	assert.Contains(t, err.Error(), "failed to get character contracts")
}
