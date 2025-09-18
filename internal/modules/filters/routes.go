package filters

import (
	"go-starter/internal/modules/auth/middlewares"
	"go-starter/internal/modules/auth/services"
	"go-starter/internal/modules/filters/handlers"
	"go-starter/internal/modules/filters/repositories"
	filterServices "go-starter/internal/modules/filters/services"
	"go-starter/pkg/config"
	"go-starter/pkg/db"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(api *echo.Group, database *db.Database, cfg *config.Config) {
	// Initialize dependencies
	filterRepo := repositories.NewRepository(database)
	filterService := filterServices.NewFilterService(filterRepo)
	filterAPIHandler := handlers.NewFilterAPIHandler(filterService)

	// Create JWT service for middleware
	jwtService := services.NewJWTService()

	// Create filters group with JWT middleware
	filtersGroup := api.Group("/filters", middlewares.JWTMiddleware(jwtService))

	// Filter CRUD routes
	filtersGroup.POST("", filterAPIHandler.SaveFilter)                         // POST /api/filters
	filtersGroup.GET("", filterAPIHandler.GetAllFilters)                       // GET /api/filters
	filtersGroup.GET("/:id", filterAPIHandler.GetFilter)                       // GET /api/filters/:id
	filtersGroup.PUT("/:id", filterAPIHandler.UpdateFilter)                    // PUT /api/filters/:id
	filtersGroup.DELETE("/:id", filterAPIHandler.DeleteFilter)                 // DELETE /api/filters/:id
	filtersGroup.GET("/by-name/:name", filterAPIHandler.GetFilterByName)       // GET /api/filters/by-name/:name
	filtersGroup.DELETE("/by-name/:name", filterAPIHandler.DeleteFilterByName) // DELETE /api/filters/by-name/:name
	filtersGroup.GET("/stats", filterAPIHandler.GetFilterStats)                // GET /api/filters/stats
}
