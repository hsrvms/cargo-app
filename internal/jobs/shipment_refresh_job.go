package jobs

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go-starter/internal/modules/shipments/repositories"
	"go-starter/internal/modules/shipments/services"
	"go-starter/pkg/ratelimiter"

	"github.com/google/uuid"
)

// ShipmentRefreshJob handles background refreshing of shipments
type ShipmentRefreshJob struct {
	shipmentRepo    repositories.ShipmentRepository
	shipmentService services.ShipmentService
	rateLimiter     *ratelimiter.SafeCubeAPIRateLimiter
	config          ShipmentRefreshConfig
}

// ShipmentRefreshConfig contains configuration for the refresh job
type ShipmentRefreshConfig struct {
	// RefreshInterval is how often to run the full refresh cycle
	RefreshInterval time.Duration
	// ConcurrentWorkers is the number of goroutines to use for parallel processing
	ConcurrentWorkers int
	// MaxShipmentsPerRun limits how many shipments to process in one cycle (0 = no limit)
	MaxShipmentsPerRun int
	// SkipRecentlyUpdated skips shipments updated within this duration
	SkipRecentlyUpdated time.Duration
}

// DefaultShipmentRefreshConfig returns a default configuration
func DefaultShipmentRefreshConfig() ShipmentRefreshConfig {
	return ShipmentRefreshConfig{
		RefreshInterval:     3 * time.Hour,
		ConcurrentWorkers:   5,                // Conservative to respect rate limits
		MaxShipmentsPerRun:  0,                // No limit
		SkipRecentlyUpdated: 30 * time.Minute, // Don't refresh if updated in last 30 minutes
	}
}

// RefreshStats contains statistics about a refresh run
type RefreshStats struct {
	StartTime         time.Time
	EndTime           time.Time
	TotalShipments    int
	SuccessfulRefresh int
	FailedRefresh     int
	SkippedShipments  int
	Errors            []RefreshError
}

// RefreshError represents an error during refresh
type RefreshError struct {
	ShipmentID     uuid.UUID
	ShipmentNumber string
	Error          error
	Timestamp      time.Time
}

// NewShipmentRefreshJob creates a new shipment refresh job
func NewShipmentRefreshJob(
	shipmentRepo repositories.ShipmentRepository,
	shipmentService services.ShipmentService,
	rateLimiter *ratelimiter.SafeCubeAPIRateLimiter,
	config ShipmentRefreshConfig,
) *ShipmentRefreshJob {
	return &ShipmentRefreshJob{
		shipmentRepo:    shipmentRepo,
		shipmentService: shipmentService,
		rateLimiter:     rateLimiter,
		config:          config,
	}
}

// Start begins the background refresh job with the configured interval
func (j *ShipmentRefreshJob) Start(ctx context.Context) {
	log.Printf("Starting shipment refresh job with %v interval", j.config.RefreshInterval)

	ticker := time.NewTicker(j.config.RefreshInterval)
	defer ticker.Stop()

	// Run once immediately
	go func() {
		stats := j.RefreshAllShipments(ctx)
		j.logRefreshStats(stats)
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("Shipment refresh job stopped")
			return
		case <-ticker.C:
			go func() {
				stats := j.RefreshAllShipments(ctx)
				j.logRefreshStats(stats)
			}()
		}
	}
}

// RefreshAllShipments refreshes all shipments in the system
func (j *ShipmentRefreshJob) RefreshAllShipments(ctx context.Context) RefreshStats {
	stats := RefreshStats{
		StartTime: time.Now(),
		Errors:    make([]RefreshError, 0),
	}

	log.Println("Starting bulk shipment refresh...")

	// Get all shipments that need refreshing
	shipments, err := j.getShipmentsForRefresh(ctx)
	if err != nil {
		log.Printf("Failed to get shipments for refresh: %v", err)
		stats.EndTime = time.Now()
		return stats
	}

	stats.TotalShipments = len(shipments)
	log.Printf("Found %d shipments to refresh", stats.TotalShipments)

	if stats.TotalShipments == 0 {
		stats.EndTime = time.Now()
		return stats
	}

	// Limit shipments if configured
	if j.config.MaxShipmentsPerRun > 0 && len(shipments) > j.config.MaxShipmentsPerRun {
		log.Printf("Limiting refresh to %d shipments (found %d)", j.config.MaxShipmentsPerRun, len(shipments))
		shipments = shipments[:j.config.MaxShipmentsPerRun]
		stats.TotalShipments = len(shipments)
	}

	// Create channels for work distribution
	shipmentChan := make(chan ShipmentForRefresh, len(shipments))
	resultChan := make(chan RefreshResult, len(shipments))

	// Send all shipments to the work channel
	for _, shipment := range shipments {
		shipmentChan <- shipment
	}
	close(shipmentChan)

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < j.config.ConcurrentWorkers; i++ {
		wg.Add(1)
		go j.refreshWorker(ctx, shipmentChan, resultChan, &wg)
	}

	// Close result channel when all workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for result := range resultChan {
		if result.Success {
			stats.SuccessfulRefresh++
		} else {
			stats.FailedRefresh++
			stats.Errors = append(stats.Errors, RefreshError{
				ShipmentID:     result.ShipmentID,
				ShipmentNumber: result.ShipmentNumber,
				Error:          result.Error,
				Timestamp:      time.Now(),
			})
		}
	}

	stats.EndTime = time.Now()
	return stats
}

// ShipmentForRefresh represents a shipment for background processing
type ShipmentForRefresh = repositories.ShipmentForRefresh

// RefreshResult represents the result of refreshing a single shipment
type RefreshResult struct {
	ShipmentID     uuid.UUID
	ShipmentNumber string
	Success        bool
	Error          error
	Duration       time.Duration
}

// refreshWorker is a worker goroutine that processes shipments
func (j *ShipmentRefreshJob) refreshWorker(
	ctx context.Context,
	shipmentChan <-chan ShipmentForRefresh,
	resultChan chan<- RefreshResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for shipment := range shipmentChan {
		result := j.refreshSingleShipment(ctx, shipment)
		resultChan <- result
	}
}

// refreshSingleShipment refreshes a single shipment
func (j *ShipmentRefreshJob) refreshSingleShipment(ctx context.Context, shipment ShipmentForRefresh) RefreshResult {
	startTime := time.Now()

	result := RefreshResult{
		ShipmentID:     shipment.ID,
		ShipmentNumber: shipment.ShipmentNumber,
		Success:        false,
	}

	// Wait for rate limiter
	if err := j.rateLimiter.Wait(ctx); err != nil {
		result.Error = fmt.Errorf("rate limiter error: %w", err)
		result.Duration = time.Since(startTime)
		return result
	}

	log.Printf("System refreshing shipment %s (ID: %s)",
		shipment.ShipmentNumber, shipment.ID)

	// Use the system refresh service (no user authentication required)
	_, err := j.shipmentService.SystemRefreshShipment(ctx, shipment.ID)
	if err != nil {
		log.Printf("Failed to system refresh shipment %s: %v", shipment.ShipmentNumber, err)
		result.Error = err
	} else {
		log.Printf("Successfully system refreshed shipment %s", shipment.ShipmentNumber)
		result.Success = true
	}

	result.Duration = time.Since(startTime)
	return result
}

// getShipmentsForRefresh gets all shipments that need to be refreshed
func (j *ShipmentRefreshJob) getShipmentsForRefresh(ctx context.Context) ([]ShipmentForRefresh, error) {
	return j.shipmentRepo.GetAllShipmentsForRefresh(ctx, j.config.SkipRecentlyUpdated)
}

// logRefreshStats logs the statistics from a refresh run
func (j *ShipmentRefreshJob) logRefreshStats(stats RefreshStats) {
	duration := stats.EndTime.Sub(stats.StartTime)

	log.Printf("System shipment refresh completed in %v", duration)
	log.Printf("Total: %d, Successful: %d, Failed: %d, Skipped: %d",
		stats.TotalShipments, stats.SuccessfulRefresh, stats.FailedRefresh, stats.SkippedShipments)

	if len(stats.Errors) > 0 {
		log.Printf("System refresh errors (%d):", len(stats.Errors))
		for _, err := range stats.Errors {
			log.Printf("  - Shipment %s (%s): %v",
				err.ShipmentNumber, err.ShipmentID, err.Error)
		}
	}

	// Log rate limiter status
	tokensAvailable := j.rateLimiter.TokensAvailable()
	log.Printf("Rate limiter tokens available: %d", tokensAvailable)
}

// GetRefreshStats returns a copy of the current refresh configuration
func (j *ShipmentRefreshJob) GetConfig() ShipmentRefreshConfig {
	return j.config
}

// UpdateConfig allows updating the job configuration at runtime
func (j *ShipmentRefreshJob) UpdateConfig(newConfig ShipmentRefreshConfig) {
	j.config = newConfig
	log.Printf("Shipment refresh job configuration updated: %+v", newConfig)
}
