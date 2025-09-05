package server

import (
	"go-starter/internal/modules/auth"
	"go-starter/internal/modules/dashboard"
	"go-starter/internal/modules/jobs"
	"go-starter/internal/modules/shipments"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) initRoutes() {
	api := s.Echo.Group("/api")

	api.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	auth.RegisterRoutes(s.Echo, api, s.DB)
	dashboard.RegisterRoutes(s.Echo, api, s.DB)
	shipments.RegisterRoutes(s.Echo, api, s.DB, s.Config)
	jobs.RegisterRoutes(s.Echo, api, s.DB, s.Config, s.JobScheduler)

}
