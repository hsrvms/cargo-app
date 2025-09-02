package services

import (
	"context"
	"encoding/json"
	"fmt"
	shipmentsDto "go-starter/internal/modules/shipments/dto"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type safeCubeAPIService struct {
	httpClient *http.Client
	baseUrl    string
	apiKey     string
}

func NewSafeCubeAPIService(baseUrl, apiKey string) SafeCubeAPIService {
	return &safeCubeAPIService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseUrl: baseUrl,
		apiKey:  apiKey,
	}
}

func (s *safeCubeAPIService) GetShipmentDetails(
	ctx context.Context,
	shipmentNumber string,
	shipmentType string,
	sealine string,
) (*shipmentsDto.SafeCubeAPIShipmentResponse, error) {
	log.Printf("SafeCube API: Requesting shipment details for %s (type: %s, sealine: %s)", shipmentNumber, shipmentType, sealine)

	apiUrl, err := url.Parse(s.baseUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	apiUrl.Path = apiUrl.Path + "/shipment"

	params := url.Values{}
	params.Add("shipmentNumber", shipmentNumber)
	if shipmentType != "" {
		params.Add("shipmentType", shipmentType)
	}
	if sealine != "" {
		params.Add("sealine", sealine)
	}

	apiUrl.RawQuery = params.Encode()
	log.Printf("SafeCube API: Making request to URL: %s", apiUrl.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiUrl.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("API_KEY", s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	log.Printf("SafeCube API: Request headers set, making HTTP call...")
	startTime := time.Now()
	resp, err := s.httpClient.Do(req)
	duration := time.Since(startTime)
	if err != nil {
		log.Printf("SafeCube API: HTTP request failed after %v: %v", duration, err)
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}

	log.Printf("SafeCube API: HTTP request completed in %v, status: %d", duration, resp.StatusCode)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("SafeCube API: Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	log.Printf("SafeCube API: Response body size: %d bytes", len(body))

	if resp.StatusCode != http.StatusOK {
		log.Printf("SafeCube API: Non-200 status code %d, response: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var safeCubeShipmentResponse shipmentsDto.SafeCubeAPIShipmentResponse
	err = json.Unmarshal(body, &safeCubeShipmentResponse)
	if err != nil {
		log.Printf("SafeCube API: Failed to unmarshal JSON response: %v. Response body: %s", err, string(body))
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return &safeCubeShipmentResponse, nil
}
