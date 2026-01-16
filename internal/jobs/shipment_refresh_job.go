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

type ShipmentRefreshJob struct {
	shipmentRepo    repositories.ShipmentRepository
	shipmentService services.ShipmentService
	rateLimiter     *ratelimiter.SafeCubeAPIRateLimiter
	config          ShipmentRefreshConfig
}

type ShipmentRefreshConfig struct {
	RefreshInterval     time.Duration
	ConcurrentWorkers   int
	MaxShipmentsPerRun  int
	SkipRecentlyUpdated time.Duration
}

func DefaultShipmentRefreshConfig() ShipmentRefreshConfig {
	return ShipmentRefreshConfig{
		RefreshInterval:     3 * time.Hour,
		ConcurrentWorkers:   5,                // Conservative to respect rate limits
		MaxShipmentsPerRun:  0,                // No limit
		SkipRecentlyUpdated: 30 * time.Minute, // Don't refresh if updated in last 30 minutes
	}
}

type RefreshStats struct {
	StartTime         time.Time
	EndTime           time.Time
	TotalShipments    int
	SuccessfulRefresh int
	FailedRefresh     int
	SkippedShipments  int
	Errors            []RefreshError
}

type RefreshError struct {
	ShipmentID     uuid.UUID
	ShipmentNumber string
	Error          error
	Timestamp      time.Time
}

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

func (j *ShipmentRefreshJob) Start(ctx context.Context) {
	log.Printf("Starting shipment refresh job with %v interval", j.config.RefreshInterval)

	ticker := time.NewTicker(j.config.RefreshInterval)
	defer ticker.Stop()

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

func (j *ShipmentRefreshJob) RefreshAllShipments(ctx context.Context) RefreshStats {
	stats := RefreshStats{
		StartTime: time.Now(),
		Errors:    make([]RefreshError, 0),
	}

	log.Println("Starting bulk shipment refresh...")

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

	if j.config.MaxShipmentsPerRun > 0 && len(shipments) > j.config.MaxShipmentsPerRun {
		log.Printf("Limiting refresh to %d shipments (found %d)", j.config.MaxShipmentsPerRun, len(shipments))
		shipments = shipments[:j.config.MaxShipmentsPerRun]
		stats.TotalShipments = len(shipments)
	}

	shipmentChan := make(chan ShipmentForRefresh, len(shipments))
	resultChan := make(chan RefreshResult, len(shipments))

	for _, shipment := range shipments {
		shipmentChan <- shipment
	}
	close(shipmentChan)

	var wg sync.WaitGroup
	for i := 0; i < j.config.ConcurrentWorkers; i++ {
		wg.Add(1)
		go j.refreshWorker(ctx, shipmentChan, resultChan, &wg)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

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

type ShipmentForRefresh = repositories.ShipmentForRefresh

type RefreshResult struct {
	ShipmentID     uuid.UUID
	ShipmentNumber string
	Success        bool
	Error          error
	Duration       time.Duration
}

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

func (j *ShipmentRefreshJob) refreshSingleShipment(ctx context.Context, shipment ShipmentForRefresh) RefreshResult {
	startTime := time.Now()

	result := RefreshResult{
		ShipmentID:     shipment.ID,
		ShipmentNumber: shipment.ShipmentNumber,
		Success:        false,
	}

	if err := j.rateLimiter.Wait(ctx); err != nil {
		result.Error = fmt.Errorf("rate limiter error: %w", err)
		result.Duration = time.Since(startTime)
		return result
	}

	log.Printf("System refreshing shipment %s (ID: %s)",
		shipment.ShipmentNumber, shipment.ID)

	_, err := j.shipmentService.SystemRefreshShipment(ctx, shipment.ID)
	if err != nil {
		log.Printf("Failed to system refresh shipment %s: %v", shipment.ShipmentNumber, err)
		result.Error = err
	} else {
		result.Success = true
	}

	result.Duration = time.Since(startTime)
	return result
}

func (j *ShipmentRefreshJob) getShipmentsForRefresh(ctx context.Context) ([]ShipmentForRefresh, error) {
	return j.shipmentRepo.GetAllShipmentsForRefresh(ctx, j.config.SkipRecentlyUpdated)
}

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

	tokensAvailable := j.rateLimiter.TokensAvailable()
	log.Printf("Rate limiter tokens available: %d", tokensAvailable)
}

func (j *ShipmentRefreshJob) GetConfig() ShipmentRefreshConfig {
	return j.config
}

func (j *ShipmentRefreshJob) UpdateConfig(newConfig ShipmentRefreshConfig) {
	j.config = newConfig
	log.Printf("Shipment refresh job configuration updated: %+v", newConfig)
}
