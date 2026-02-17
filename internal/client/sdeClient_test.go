package client_test

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_SdeClient_GetChecksum(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTP := NewMockHTTPDoer(ctrl)

	mockHTTP.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "https://test.example.com/tranquility/latest.jsonl", req.URL.String())
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"_key":"sde","buildNumber":3201939}`))),
			}, nil
		})

	c := client.NewSdeClientWithBaseURL(mockHTTP, "https://test.example.com/")
	checksum, err := c.GetChecksum(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "3201939", checksum)
}

func Test_SdeClient_GetChecksumHttpError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTP := NewMockHTTPDoer(ctrl)

	mockHTTP.EXPECT().
		Do(gomock.Any()).
		Return(nil, fmt.Errorf("connection refused"))

	c := client.NewSdeClientWithBaseURL(mockHTTP, "https://test.example.com/")
	_, err := c.GetChecksum(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch checksum")
}

func Test_SdeClient_GetChecksumBadStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTP := NewMockHTTPDoer(ctrl)

	mockHTTP.EXPECT().
		Do(gomock.Any()).
		Return(&http.Response{
			StatusCode: 404,
			Body:       io.NopCloser(bytes.NewReader([]byte("not found"))),
		}, nil)

	c := client.NewSdeClientWithBaseURL(mockHTTP, "https://test.example.com/")
	_, err := c.GetChecksum(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code fetching checksum: 404")
}

func Test_SdeClient_DownloadSDE(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTP := NewMockHTTPDoer(ctrl)

	zipContent := []byte("fake-zip-content")
	mockHTTP.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "https://test.example.com/eve-online-static-data-latest-yaml.zip", req.URL.String())
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(zipContent)),
			}, nil
		})

	c := client.NewSdeClientWithBaseURL(mockHTTP, "https://test.example.com/")
	path, err := c.DownloadSDE(context.Background())

	assert.NoError(t, err)
	assert.NotEmpty(t, path)
	defer os.Remove(path)

	// Verify file was written
	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.Equal(t, zipContent, content)
}

func Test_SdeClient_DownloadSDEHttpError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTP := NewMockHTTPDoer(ctrl)

	mockHTTP.EXPECT().
		Do(gomock.Any()).
		Return(nil, fmt.Errorf("timeout"))

	c := client.NewSdeClientWithBaseURL(mockHTTP, "https://test.example.com/")
	_, err := c.DownloadSDE(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to download SDE")
}

func Test_SdeClient_DownloadSDEBadStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTP := NewMockHTTPDoer(ctrl)

	mockHTTP.EXPECT().
		Do(gomock.Any()).
		Return(&http.Response{
			StatusCode: 500,
			Body:       io.NopCloser(bytes.NewReader([]byte("error"))),
		}, nil)

	c := client.NewSdeClientWithBaseURL(mockHTTP, "https://test.example.com/")
	_, err := c.DownloadSDE(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code downloading SDE: 500")
}

func Test_SdeClient_ParseSDEWithCategoryIDs(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"categories.yaml": `
6:
  name:
    en: Ship
  published: true
  iconID: 22
9:
  name:
    en: Blueprint
  published: true
`,
	})
	defer os.Remove(zipPath)

	c := client.NewSdeClientWithBaseURL(nil, "https://test.example.com/")
	data, err := c.ParseSDE(zipPath)

	assert.NoError(t, err)
	assert.Len(t, data.Categories, 2)

	catMap := map[int64]string{}
	for _, c := range data.Categories {
		catMap[c.CategoryID] = c.Name
	}
	assert.Equal(t, "Ship", catMap[6])
	assert.Equal(t, "Blueprint", catMap[9])
}

func Test_SdeClient_ParseSDEWithTypeIDs(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"types.yaml": `
34:
  name:
    en: Tritanium
  volume: 0.01
  groupID: 18
  published: true
35:
  name:
    en: Pyerite
  volume: 0.01
  groupID: 18
  published: true
`,
	})
	defer os.Remove(zipPath)

	c := client.NewSdeClientWithBaseURL(nil, "https://test.example.com/")
	data, err := c.ParseSDE(zipPath)

	assert.NoError(t, err)
	assert.Len(t, data.Types, 2)

	typeMap := map[int64]string{}
	for _, t := range data.Types {
		typeMap[t.TypeID] = t.TypeName
	}
	assert.Equal(t, "Tritanium", typeMap[34])
	assert.Equal(t, "Pyerite", typeMap[35])
}

func Test_SdeClient_ParseSDEWithBlueprints(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"blueprints.yaml": `
681:
  maxProductionLimit: 10
  activities:
    manufacturing:
      time: 6000
      materials:
        - typeID: 34
          quantity: 100
      products:
        - typeID: 683
          quantity: 1
      skills:
        - typeID: 3380
          level: 1
`,
	})
	defer os.Remove(zipPath)

	c := client.NewSdeClientWithBaseURL(nil, "https://test.example.com/")
	data, err := c.ParseSDE(zipPath)

	assert.NoError(t, err)
	assert.Len(t, data.Blueprints, 1)
	assert.Equal(t, int64(681), data.Blueprints[0].BlueprintTypeID)
	assert.Len(t, data.BlueprintActivities, 1)
	assert.Equal(t, "manufacturing", data.BlueprintActivities[0].Activity)
	assert.Len(t, data.BlueprintMaterials, 1)
	assert.Equal(t, int64(34), data.BlueprintMaterials[0].TypeID)
	assert.Len(t, data.BlueprintProducts, 1)
	assert.Equal(t, int64(683), data.BlueprintProducts[0].TypeID)
	assert.Len(t, data.BlueprintSkills, 1)
	assert.Equal(t, int64(3380), data.BlueprintSkills[0].TypeID)
}

func Test_SdeClient_ParseSDEWithRegions(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"mapRegions.yaml": `
10000002:
  name:
    en: The Forge
10000032:
  name:
    en: Sinq Laison
`,
	})
	defer os.Remove(zipPath)

	c := client.NewSdeClientWithBaseURL(nil, "https://test.example.com/")
	data, err := c.ParseSDE(zipPath)

	assert.NoError(t, err)
	assert.Len(t, data.Regions, 2)

	regionMap := map[int64]string{}
	for _, r := range data.Regions {
		regionMap[r.ID] = r.Name
	}
	assert.Equal(t, "The Forge", regionMap[10000002])
	assert.Equal(t, "Sinq Laison", regionMap[10000032])
}

func Test_SdeClient_ParseSDEMissingFile(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"unknownFile.yaml": "data: true",
	})
	defer os.Remove(zipPath)

	c := client.NewSdeClientWithBaseURL(nil, "https://test.example.com/")
	data, err := c.ParseSDE(zipPath)

	assert.NoError(t, err)
	assert.Empty(t, data.Categories)
	assert.Empty(t, data.Types)
}

func Test_SdeClient_ParseSDEInvalidZip(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "bad-zip-*.zip")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, _ = tmpFile.Write([]byte("this is not a zip file"))
	tmpFile.Close()

	c := client.NewSdeClientWithBaseURL(nil, "https://test.example.com/")
	_, err = c.ParseSDE(tmpFile.Name())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open SDE ZIP")
}

func Test_SdeClient_ParseSDEInvalidYAML(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"categories.yaml": "{{{{ invalid yaml",
	})
	defer os.Remove(zipPath)

	c := client.NewSdeClientWithBaseURL(nil, "https://test.example.com/")
	_, err := c.ParseSDE(zipPath)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse categories.yaml")
}

func Test_SdeClient_NewSdeClientUsesDefaultBaseURL(t *testing.T) {
	c := client.NewSdeClient(nil)
	assert.NotNil(t, c)
}

// createTestZip creates a temporary ZIP file with the given files and returns its path.
func createTestZip(t *testing.T, files map[string]string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test-sde-*.zip")
	assert.NoError(t, err)

	w := zip.NewWriter(tmpFile)
	for name, content := range files {
		f, err := w.Create(name)
		assert.NoError(t, err)
		_, err = f.Write([]byte(content))
		assert.NoError(t, err)
	}
	assert.NoError(t, w.Close())
	assert.NoError(t, tmpFile.Close())

	return tmpFile.Name()
}
