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

func Test_ClientShouldGetCorporationContracts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := NewMockHTTPDoer(ctrl)

	mockContracts := []*client.EsiContract{
		{
			ContractID:     54321,
			IssuerID:       100,
			AcceptorID:     200,
			AssigneeID:     5001,
			Type:           "item_exchange",
			Status:         "finished",
			Title:          "Corp delivery PT-99",
			DateCompleted:  "2025-01-15T12:00:00Z",
			ForCorporation: true,
			Price:          2000000.0,
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
			assert.Contains(t, req.URL.String(), "/v1/corporations/5001/contracts/")
			return mockResponse, nil
		}).
		Times(1)

	esiClient := client.NewEsiClientWithHTTPClient("test-client-id", "test-client-secret", mockHTTPClient, "https://esi.test.com")

	contracts, err := esiClient.GetCorporationContracts(context.Background(), 5001, "test-token", "test-refresh", time.Now())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(contracts))
	assert.Equal(t, int64(54321), contracts[0].ContractID)
	assert.Equal(t, "item_exchange", contracts[0].Type)
	assert.Equal(t, "finished", contracts[0].Status)
	assert.Equal(t, "Corp delivery PT-99", contracts[0].Title)
	assert.True(t, contracts[0].ForCorporation)
}

func Test_ClientShouldHandleCorporationContractsError(t *testing.T) {
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

	contracts, err := esiClient.GetCorporationContracts(context.Background(), 5001, "bad-token", "ref", time.Now())
	assert.Error(t, err)
	assert.Nil(t, contracts)
	assert.Contains(t, err.Error(), "failed to get corporation contracts")
}

func Test_ClientShouldGetCharacterSkills(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := NewMockHTTPDoer(ctrl)

	mockSkills := client.EsiSkillsResponse{
		Skills: []client.EsiSkillEntry{
			{SkillID: 3380, TrainedSkillLevel: 5, ActiveSkillLevel: 5, SkillpointsInSkill: 256000},
			{SkillID: 3388, TrainedSkillLevel: 4, ActiveSkillLevel: 4, SkillpointsInSkill: 45255},
			{SkillID: 22242, TrainedSkillLevel: 3, ActiveSkillLevel: 3, SkillpointsInSkill: 16000},
		},
		TotalSP:   5000000,
		UnallocSP: 100000,
	}

	skillsJSON, _ := json.Marshal(mockSkills)

	mockHTTPClient.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "GET", req.Method)
			assert.Contains(t, req.URL.String(), "/v4/characters/12345/skills/")
			assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(skillsJSON)),
			}, nil
		}).
		Times(1)

	esiClient := client.NewEsiClientWithHTTPClient("test-client-id", "test-client-secret", mockHTTPClient, "https://esi.test.com")

	resp, err := esiClient.GetCharacterSkills(context.Background(), 12345, "test-token")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(resp.Skills))
	assert.Equal(t, int64(3380), resp.Skills[0].SkillID)
	assert.Equal(t, 5, resp.Skills[0].TrainedSkillLevel)
	assert.Equal(t, 5, resp.Skills[0].ActiveSkillLevel)
	assert.Equal(t, int64(256000), resp.Skills[0].SkillpointsInSkill)
	assert.Equal(t, int64(5000000), resp.TotalSP)
	assert.Equal(t, int64(100000), resp.UnallocSP)
}

func Test_ClientShouldHandleCharacterSkillsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := NewMockHTTPDoer(ctrl)

	mockHTTPClient.EXPECT().
		Do(gomock.Any()).
		Return(&http.Response{
			StatusCode: 403,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"token is expired"}`))),
		}, nil).
		Times(1)

	esiClient := client.NewEsiClientWithHTTPClient("test-client-id", "test-client-secret", mockHTTPClient, "https://esi.test.com")

	resp, err := esiClient.GetCharacterSkills(context.Background(), 12345, "expired-token")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to get character skills")
}

func Test_ClientShouldGetCharacterIndustryJobs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := NewMockHTTPDoer(ctrl)

	cost := 1500000.0
	productTypeID := int64(34)
	mockJobs := []*client.EsiIndustryJob{
		{
			JobID:               12345,
			InstallerID:         100,
			FacilityID:          60003760,
			StationID:           60003760,
			ActivityID:          1,
			BlueprintID:         9876,
			BlueprintTypeID:     787,
			BlueprintLocationID: 60003760,
			OutputLocationID:    60003760,
			Runs:                10,
			Cost:                &cost,
			ProductTypeID:       &productTypeID,
			Status:              "active",
			Duration:            3600,
			StartDate:           "2026-02-22T10:00:00Z",
			EndDate:             "2026-02-22T11:00:00Z",
		},
	}

	jobsJSON, _ := json.Marshal(mockJobs)

	mockHTTPClient.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "GET", req.Method)
			assert.Contains(t, req.URL.String(), "/v1/characters/12345/industry/jobs/")
			assert.NotContains(t, req.URL.String(), "include_completed=true")
			assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"X-Pages": []string{"1"}},
				Body:       io.NopCloser(bytes.NewReader(jobsJSON)),
			}, nil
		}).
		Times(1)

	esiClient := client.NewEsiClientWithHTTPClient("test-client-id", "test-client-secret", mockHTTPClient, "https://esi.test.com")

	jobs, err := esiClient.GetCharacterIndustryJobs(context.Background(), 12345, "test-token", false)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(jobs))
	assert.Equal(t, int64(12345), jobs[0].JobID)
	assert.Equal(t, 1, jobs[0].ActivityID)
	assert.Equal(t, "active", jobs[0].Status)
	assert.Equal(t, 10, jobs[0].Runs)
	assert.Equal(t, 1500000.0, *jobs[0].Cost)
	assert.Equal(t, int64(34), *jobs[0].ProductTypeID)
}

func Test_ClientShouldGetCharacterIndustryJobsIncludeCompleted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := NewMockHTTPDoer(ctrl)

	completedDate := "2026-02-22T11:00:00Z"
	successfulRuns := 10
	mockJobs := []*client.EsiIndustryJob{
		{
			JobID:           12345,
			InstallerID:     100,
			FacilityID:      60003760,
			StationID:       60003760,
			ActivityID:      1,
			BlueprintID:     9876,
			BlueprintTypeID: 787,
			BlueprintLocationID: 60003760,
			OutputLocationID:    60003760,
			Runs:            10,
			Status:          "delivered",
			Duration:        3600,
			StartDate:       "2026-02-22T10:00:00Z",
			EndDate:         "2026-02-22T11:00:00Z",
			CompletedDate:   &completedDate,
			SuccessfulRuns:  &successfulRuns,
		},
	}

	jobsJSON, _ := json.Marshal(mockJobs)

	mockHTTPClient.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			assert.Contains(t, req.URL.String(), "include_completed=true")
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"X-Pages": []string{"1"}},
				Body:       io.NopCloser(bytes.NewReader(jobsJSON)),
			}, nil
		}).
		Times(1)

	esiClient := client.NewEsiClientWithHTTPClient("test-client-id", "test-client-secret", mockHTTPClient, "https://esi.test.com")

	jobs, err := esiClient.GetCharacterIndustryJobs(context.Background(), 12345, "test-token", true)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(jobs))
	assert.Equal(t, "delivered", jobs[0].Status)
	assert.Equal(t, 10, *jobs[0].SuccessfulRuns)
}

func Test_ClientShouldGetCharacterIndustryJobsMultiplePages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := NewMockHTTPDoer(ctrl)

	page1Jobs := []*client.EsiIndustryJob{
		{JobID: 1001, InstallerID: 100, FacilityID: 60003760, StationID: 60003760, ActivityID: 1, BlueprintID: 1, BlueprintTypeID: 787, BlueprintLocationID: 60003760, OutputLocationID: 60003760, Runs: 1, Status: "active", Duration: 3600, StartDate: "2026-02-22T10:00:00Z", EndDate: "2026-02-22T11:00:00Z"},
	}
	page2Jobs := []*client.EsiIndustryJob{
		{JobID: 2001, InstallerID: 100, FacilityID: 60003760, StationID: 60003760, ActivityID: 9, BlueprintID: 2, BlueprintTypeID: 788, BlueprintLocationID: 60003760, OutputLocationID: 60003760, Runs: 5, Status: "active", Duration: 7200, StartDate: "2026-02-22T12:00:00Z", EndDate: "2026-02-22T14:00:00Z"},
	}

	page1JSON, _ := json.Marshal(page1Jobs)
	page2JSON, _ := json.Marshal(page2Jobs)

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

	jobs, err := esiClient.GetCharacterIndustryJobs(context.Background(), 12345, "test-token", false)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(jobs))
	assert.Equal(t, int64(1001), jobs[0].JobID)
	assert.Equal(t, int64(2001), jobs[1].JobID)
}

func Test_ClientShouldHandleCharacterIndustryJobsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := NewMockHTTPDoer(ctrl)

	mockHTTPClient.EXPECT().
		Do(gomock.Any()).
		Return(&http.Response{
			StatusCode: 403,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"forbidden"}`))),
		}, nil).
		Times(1)

	esiClient := client.NewEsiClientWithHTTPClient("test-client-id", "test-client-secret", mockHTTPClient, "https://esi.test.com")

	jobs, err := esiClient.GetCharacterIndustryJobs(context.Background(), 12345, "bad-token", false)
	assert.Error(t, err)
	assert.Nil(t, jobs)
	assert.Contains(t, err.Error(), "failed to get character industry jobs")
}

func Test_ClientShouldGetCharacterIndustryJobsNoXPagesHeader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := NewMockHTTPDoer(ctrl)

	mockJobs := []*client.EsiIndustryJob{
		{JobID: 5001, InstallerID: 100, FacilityID: 60003760, StationID: 60003760, ActivityID: 1, BlueprintID: 1, BlueprintTypeID: 787, BlueprintLocationID: 60003760, OutputLocationID: 60003760, Runs: 1, Status: "active", Duration: 3600, StartDate: "2026-02-22T10:00:00Z", EndDate: "2026-02-22T11:00:00Z"},
	}
	jobsJSON, _ := json.Marshal(mockJobs)

	mockHTTPClient.EXPECT().
		Do(gomock.Any()).
		Return(&http.Response{
			StatusCode: 200,
			Header:     http.Header{},
			Body:       io.NopCloser(bytes.NewReader(jobsJSON)),
		}, nil).
		Times(1)

	esiClient := client.NewEsiClientWithHTTPClient("test-client-id", "test-client-secret", mockHTTPClient, "https://esi.test.com")

	jobs, err := esiClient.GetCharacterIndustryJobs(context.Background(), 12345, "test-token", false)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(jobs))
	assert.Equal(t, int64(5001), jobs[0].JobID)
}

func Test_ClientShouldGetCorporationIndustryJobs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := NewMockHTTPDoer(ctrl)

	cost := 5000000.0
	mockJobs := []*client.EsiIndustryJob{
		{
			JobID:               99001,
			InstallerID:         200,
			FacilityID:          1000000025,
			StationID:           1000000025,
			ActivityID:          9,
			BlueprintID:         5555,
			BlueprintTypeID:     46166,
			BlueprintLocationID: 1000000025,
			OutputLocationID:    1000000025,
			Runs:                100,
			Cost:                &cost,
			Status:              "active",
			Duration:            14400,
			StartDate:           "2026-02-22T08:00:00Z",
			EndDate:             "2026-02-22T12:00:00Z",
		},
	}

	jobsJSON, _ := json.Marshal(mockJobs)

	mockHTTPClient.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "GET", req.Method)
			assert.Contains(t, req.URL.String(), "/v1/corporations/5001/industry/jobs/")
			assert.Equal(t, "Bearer corp-token", req.Header.Get("Authorization"))
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"X-Pages": []string{"1"}},
				Body:       io.NopCloser(bytes.NewReader(jobsJSON)),
			}, nil
		}).
		Times(1)

	esiClient := client.NewEsiClientWithHTTPClient("test-client-id", "test-client-secret", mockHTTPClient, "https://esi.test.com")

	jobs, err := esiClient.GetCorporationIndustryJobs(context.Background(), 5001, "corp-token", false)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(jobs))
	assert.Equal(t, int64(99001), jobs[0].JobID)
	assert.Equal(t, 9, jobs[0].ActivityID)
	assert.Equal(t, "active", jobs[0].Status)
	assert.Equal(t, 100, jobs[0].Runs)
	assert.Equal(t, 5000000.0, *jobs[0].Cost)
}

func Test_ClientShouldHandleCorporationIndustryJobsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := NewMockHTTPDoer(ctrl)

	mockHTTPClient.EXPECT().
		Do(gomock.Any()).
		Return(&http.Response{
			StatusCode: 403,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"forbidden"}`))),
		}, nil).
		Times(1)

	esiClient := client.NewEsiClientWithHTTPClient("test-client-id", "test-client-secret", mockHTTPClient, "https://esi.test.com")

	jobs, err := esiClient.GetCorporationIndustryJobs(context.Background(), 5001, "bad-token", false)
	assert.Error(t, err)
	assert.Nil(t, jobs)
	assert.Contains(t, err.Error(), "failed to get corporation industry jobs")
}
