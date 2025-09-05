package jobs

import (
	"net/http"
	"time"

	"go-starter/internal/jobs"

	"github.com/labstack/echo/v4"
)

// JobHandler handles job management API endpoints
type JobHandler struct {
	scheduler *jobs.JobScheduler
}

// NewJobHandler creates a new job handler
func NewJobHandler(scheduler *jobs.JobScheduler) *JobHandler {
	return &JobHandler{
		scheduler: scheduler,
	}
}

// JobStatus represents the status response for jobs
type JobStatus struct {
	IsRunning       bool      `json:"is_running"`
	Uptime          string    `json:"uptime"`
	UptimeSeconds   float64   `json:"uptime_seconds"`
	RegisteredJobs  []string  `json:"registered_jobs"`
	TotalJobs       int       `json:"total_jobs"`
	LastStatusCheck time.Time `json:"last_status_check"`
}

// GetJobsStatus returns the current status of all background jobs
func (h *JobHandler) GetJobsStatus(c echo.Context) error {
	uptime := h.scheduler.GetUptime()
	registeredJobs := h.scheduler.GetRegisteredJobs()

	status := JobStatus{
		IsRunning:       h.scheduler.IsRunning(),
		Uptime:          uptime.String(),
		UptimeSeconds:   uptime.Seconds(),
		RegisteredJobs:  registeredJobs,
		TotalJobs:       len(registeredJobs),
		LastStatusCheck: time.Now(),
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   status,
	})
}

// HealthCheck provides a simple health check for job scheduler
func (h *JobHandler) HealthCheck(c echo.Context) error {
	isHealthy := h.scheduler.IsRunning()
	status := "healthy"
	httpStatus := http.StatusOK

	if !isHealthy {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	return c.JSON(httpStatus, map[string]interface{}{
		"status":            status,
		"timestamp":         time.Now(),
		"scheduler_running": isHealthy,
	})
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// sendErrorResponse sends a standardized error response
func (h *JobHandler) sendErrorResponse(c echo.Context, statusCode int, message string) error {
	return c.JSON(statusCode, ErrorResponse{
		Status:  "error",
		Message: message,
		Code:    statusCode,
	})
}

// sendSuccessResponse sends a standardized success response
func (h *JobHandler) sendSuccessResponse(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusOK, SuccessResponse{
		Status:    "success",
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	})
}
