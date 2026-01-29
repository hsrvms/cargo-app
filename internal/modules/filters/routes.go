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

// Filter CRUD routes
func RegisterRoutes(api *echo.Group, database *db.Database, cfg *config.Config) {
	filterRepo := repositories.NewRepository(database)
	filterService := filterServices.NewFilterService(filterRepo)
	filterAPIHandler := handlers.NewFilterAPIHandler(filterService)

	jwtService := services.NewJWTService()

	filtersGroup := api.Group("/filters", middlewares.JWTMiddleware(jwtService))

	filtersGroup.POST("", filterAPIHandler.SaveFilter)
	filtersGroup.GET("", filterAPIHandler.GetAllFilters)
	filtersGroup.GET("/:id", filterAPIHandler.GetFilter)
	filtersGroup.PUT("/:id", filterAPIHandler.UpdateFilter)
	filtersGroup.DELETE("/:id", filterAPIHandler.DeleteFilter)
	filtersGroup.GET("/by-name/:name", filterAPIHandler.GetFilterByName)
	filtersGroup.DELETE("/by-name/:name", filterAPIHandler.DeleteFilterByName)
	filtersGroup.GET("/stats", filterAPIHandler.GetFilterStats)
}
