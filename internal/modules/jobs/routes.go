package jobs

import (
	"go-starter/internal/jobs"
	"go-starter/pkg/config"
	"go-starter/pkg/db"

	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers job management routes
func RegisterRoutes(e *echo.Echo, api *echo.Group, database *db.Database, cfg *config.Config, jobScheduler *jobs.JobScheduler) {
	jobHandler := NewJobHandler(jobScheduler)

	// Public health check endpoint
	api.GET("/jobs/health", jobHandler.HealthCheck)

	// Protected job management endpoints
	jobsAPI := api.Group("/jobs")

	// Note: You may want to add authentication middleware here
	// jobsAPI.Use(middlewares.JWTMiddleware(jwtService))

	jobsAPI.GET("/status", jobHandler.GetJobsStatus)
}
