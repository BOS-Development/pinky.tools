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
	esiClient := client.NewEsiClientWithHTTPClient("test-client-id", "test-client-secret", mockHTTPClient)

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
	esiClient := client.NewEsiClientWithHTTPClient("test-client-id", "test-client-secret", mockHTTPClient)

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
