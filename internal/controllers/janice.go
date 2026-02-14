package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type Janice struct {
	httpClient *http.Client
}

func NewJanice(router Routerer) *Janice {
	controller := &Janice{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	log.Info("registering janice controller", "endpoint", "/v1/janice/appraisal")
	router.RegisterRestAPIRoute("/v1/janice/appraisal", web.AuthAccessUser, controller.CreateAppraisal, "POST")

	return controller
}

type JaniceAppraisalRequest struct {
	Items string `json:"items"`
}

type JaniceAppraisalResponse struct {
	Code string `json:"code"`
}

func (c *Janice) CreateAppraisal(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	var req JaniceAppraisalRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		log.Error("failed to decode janice request", "error", err)
		return nil, &web.HttpError{
			StatusCode: 400,
			Error:      errors.Wrap(err, "failed to decode request body"),
		}
	}

	log.Info("creating janice appraisal", "items_length", len(req.Items))

	// Janice API expects plain text items, not JSON
	// Build URL with query parameters
	url := "https://janice.e-351.com/api/rest/v2/appraisal?market=2&price_percentage=100"

	log.Info("calling janice api", "url", url, "items", req.Items)

	// Create HTTP request with plain text body
	httpReq, err := http.NewRequest("POST", url, bytes.NewBufferString(req.Items))
	if err != nil {
		log.Error("failed to create janice request", "error", err)
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to create janice request"),
		}
	}

	// Janice expects text/plain content type and X-ApiKey header
	httpReq.Header.Set("Content-Type", "text/plain")
	// Using sample API key from Janice docs - you may want to make this configurable
	httpReq.Header.Set("X-ApiKey", "G9KwKq3465588VPd6747t95Zh94q3W2E")

	// Call Janice API
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Error("janice api call failed", "error", err)
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to call janice api"),
		}
	}
	defer resp.Body.Close()

	log.Info("janice api responded", "status", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Error("janice api returned error", "status", resp.StatusCode, "body", string(body))
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Errorf("janice api returned status %d: %s", resp.StatusCode, string(body)),
		}
	}

	// Parse Janice response
	var janiceResp JaniceAppraisalResponse
	if err := json.NewDecoder(resp.Body).Decode(&janiceResp); err != nil {
		log.Error("failed to decode janice response", "error", err)
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to decode janice response"),
		}
	}

	log.Info("janice appraisal created", "code", janiceResp.Code)

	return janiceResp, nil
}
