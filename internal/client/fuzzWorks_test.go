package client_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/dsnet/compress/bzip2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/annymsMthd/industry-tool/internal/client"
)

func Test_ItemType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	httpClient := NewMockHttpGetter(ctrl)

	// Create CSV test data
	csvData := "typeID,typeName,volume,iconID\n34,Tritanium,0.01,22\n35,Pyerite,0.01,23\n"

	// Compress with bzip2
	var compressed bytes.Buffer
	writer, err := bzip2.NewWriter(&compressed, &bzip2.WriterConfig{Level: bzip2.DefaultCompression})
	assert.NoError(t, err)
	_, err = writer.Write([]byte(csvData))
	assert.NoError(t, err)
	err = writer.Close()
	assert.NoError(t, err)

	// Create a mock HTTP response with compressed data
	mockResponse := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(compressed.Bytes())),
	}

	// Set expectation for HTTP GET call
	httpClient.EXPECT().
		Get("https://www.fuzzwork.co.uk/dump/latest//invTypes.csv.bz2").
		Return(mockResponse, nil)

	update := client.NewFuzzWorks(httpClient)

	items, err := update.GetInventoryTypes(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(items))
	assert.Equal(t, int64(34), items[0].TypeID)
	assert.Equal(t, "Tritanium", items[0].TypeName)
}
