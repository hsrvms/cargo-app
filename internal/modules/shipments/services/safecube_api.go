package services

import (
	"context"
	"encoding/json"
	"fmt"
	shipmentsDto "go-starter/internal/modules/shipments/dto"
	"go-starter/pkg/ratelimiter"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type mockSafeCubeAPIService struct {
	dummyFilePath string
}

func NewMockSafeCubeAPIService(filePath string) SafeCubeAPIService {
	return &mockSafeCubeAPIService{
		dummyFilePath: filePath,
	}
}

func (m *mockSafeCubeAPIService) GetShipmentDetails(
	ctx context.Context,
	shipmentNumber string,
	shipmentType string,
	sealine string,
) (*shipmentsDto.SafeCubeAPIShipmentResponse, error) {
	log.Printf("Mock SafeCube API: Reading dummy data from %s for shipment %s", m.dummyFilePath, shipmentNumber)

	jsonFile, err := os.Open(m.dummyFilePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open dummy data file: %w", err)
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read dummy data file: %w", err)
	}

	var shipmentList []shipmentsDto.SafeCubeAPIShipmentResponse
	err = json.Unmarshal(byteValue, &shipmentList)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal dummy JSON: %w", err)
	}

	shipmentNumber = strings.ToUpper(strings.TrimSpace(shipmentNumber))
	for _, shipment := range shipmentList {
		if strings.ToUpper(strings.TrimSpace(shipment.Metadata.ShipmentNumber)) == shipmentNumber {
			log.Printf("Mock SafeCube API: Found shipment %s in dummy data", shipmentNumber)
			return &shipment, nil
		}
	}

	log.Printf("Mock SafeCube API: Shipment %s not found in dummy data", shipmentNumber)
	return nil, fmt.Errorf("Shipment not found: %s", shipmentNumber)
}

func (m *mockSafeCubeAPIService) GetRawShipmentData(ctx context.Context, shipmentNumber, shipmentType, sealine string) (json.RawMessage, error) {
	return nil, nil
}

func (m *mockSafeCubeAPIService) GetRawShipmentData(ctx context.Context, shipmentNumber, shipmentType, sealine string) (json.RawMessage, error) {
	return nil, nil
}

type safeCubeAPIService struct {
	httpClient  *http.Client
	baseUrl     string
	apiKey      string
	rateLimiter *ratelimiter.SafeCubeAPIRateLimiter
}

func NewSafeCubeAPIService(baseUrl, apiKey string, rateLimiter *ratelimiter.SafeCubeAPIRateLimiter) SafeCubeAPIService {
	return &safeCubeAPIService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseUrl:     baseUrl,
		apiKey:      apiKey,
		rateLimiter: rateLimiter,
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

	// Wait for rate limiter before making the request
	log.Printf("SafeCube API: Waiting for rate limiter...")
	if err := s.rateLimiter.Wait(ctx); err != nil {
		log.Printf("SafeCube API: Rate limiter error: %v", err)
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	log.Printf("SafeCube API: Rate limiter cleared, making HTTP call...")
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

func (s *safeCubeAPIService) GetRawShipmentData(
	ctx context.Context,
	shipmentNumber string,
	shipmentType string,
	sealine string,
) (json.RawMessage, error) {
	log.Printf("SafeCube API: Requesting raw shipment data for %s (type: %s, sealine: %s)", shipmentNumber, shipmentType, sealine)

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

	log.Printf("SafeCube API: Waiting for rate limiter...")
	if err := s.rateLimiter.Wait(ctx); err != nil {
		log.Printf("SafeCube API: Rate limiter error: %v", err)
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	log.Printf("Safecube API: Rate limiter cleared, making HTTP call...")
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

	var temp any
	if err := json.Unmarshal(body, &temp); err != nil {
		log.Printf("SafeCube API: Invalid JSON response: %v. Response body: %s", err, string(body))
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return json.RawMessage(body), nil
}
