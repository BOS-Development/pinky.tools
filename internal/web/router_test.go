package web_test

import (
	"testing"

	"github.com/annymsMthd/industry-tool/internal/web"
)

func Test_RouterConstructor(t *testing.T) {
	router := web.NewRouter(8080, "test-key")

	if router == nil {
		t.Error("Router should not be nil")
	}
}
