package server

import (
	"context"
	"fmt"
	"go-starter/internal/jobs"
	shipmentRepositories "go-starter/internal/modules/shipments/repositories"
	shipmentServices "go-starter/internal/modules/shipments/services"
	"go-starter/pkg/config"
	"go-starter/pkg/db"
	"go-starter/pkg/ratelimiter"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	prettylogger "github.com/rdbell/echo-pretty-logger"
)

type Server struct {
	Echo         *echo.Echo
	DB           *db.Database
	Config       *config.Config
	JobScheduler *jobs.JobScheduler
}

func New(cfg *config.Config, database *db.Database) *Server {
	e := echo.New()

	e.Use(prettylogger.Logger)
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.Static("/assets", "web/assets")
	e.Static("/scripts", "web/scripts")
	e.Static("/config", "web/config")

	// Initialize job scheduler
	jobScheduler := jobs.NewJobScheduler()

	server := &Server{
		Echo:         e,
		DB:           database,
		Config:       cfg,
		JobScheduler: jobScheduler,
	}

	server.initRoutes()
	server.initBackgroundJobs()

	return server
}

func (s *Server) Start() {
	addr := fmt.Sprintf(":%d", s.Config.Server.Port)

	httpServer := &http.Server{
		Addr:         addr,
		ReadTimeout:  s.Config.Server.ReadTimeout,
		WriteTimeout: s.Config.Server.WriteTimeout,
		IdleTimeout:  s.Config.Server.IdleTimeout,
	}

	go func() {
		if err := s.Echo.StartServer(httpServer); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Start background jobs
	if err := s.JobScheduler.Start(); err != nil {
		log.Fatalf("Failed to start job scheduler: %v", err)
	}

	log.Printf("Server started on %s", addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	// Stop background jobs first
	if err := s.JobScheduler.Stop(); err != nil {
		log.Printf("Error stopping job scheduler: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Echo.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to gracefully shutdown server: %v", err)
	}

	log.Println("Server stopped")
}

// initBackgroundJobs initializes and registers background jobs
func (s *Server) initBackgroundJobs() {
	// Initialize SafeCube API rate limiter
	rateLimiter := ratelimiter.NewSafeCubeAPIRateLimiter()

	// Initialize services for background jobs
	safeCubeAPIService := shipmentServices.NewSafeCubeAPIService(
		s.Config.SafeCubeAPI.BaseURL,
		s.Config.SafeCubeAPI.APIKey,
		rateLimiter,
	)
	shipmentRepository := shipmentRepositories.NewShipmentRepository(s.DB)
	shipmentService := shipmentServices.NewShipmentService(shipmentRepository, safeCubeAPIService)

	// Configure shipment refresh job
	refreshConfig := jobs.ShipmentRefreshConfig{
		RefreshInterval:     s.Config.BackgroundJobs.ShipmentRefreshInterval,
		ConcurrentWorkers:   s.Config.BackgroundJobs.ShipmentRefreshWorkers,
		MaxShipmentsPerRun:  s.Config.BackgroundJobs.ShipmentMaxPerRun,
		SkipRecentlyUpdated: s.Config.BackgroundJobs.ShipmentSkipRecentlyUpdated,
	}

	// Create and register shipment refresh job
	shipmentRefreshJob := jobs.NewShipmentRefreshJob(
		shipmentRepository,
		shipmentService,
		rateLimiter,
		refreshConfig,
	)

	if err := s.JobScheduler.RegisterJob("shipment_refresh", &ShipmentRefreshJobWrapper{job: shipmentRefreshJob}); err != nil {
		log.Fatalf("Failed to register shipment refresh job: %v", err)
	}

	log.Printf("Background jobs initialized successfully")
	log.Printf("Shipment refresh configured: interval=%v, workers=%d, max_per_run=%d, skip_recently_updated=%v",
		refreshConfig.RefreshInterval,
		refreshConfig.ConcurrentWorkers,
		refreshConfig.MaxShipmentsPerRun,
		refreshConfig.SkipRecentlyUpdated)
}

// ShipmentRefreshJobWrapper adapts ShipmentRefreshJob to implement the Job interface
type ShipmentRefreshJobWrapper struct {
	job *jobs.ShipmentRefreshJob
}

func (w *ShipmentRefreshJobWrapper) Start(ctx context.Context) {
	w.job.Start(ctx)
}

func (w *ShipmentRefreshJobWrapper) GetName() string {
	return "shipment_refresh"
}
