package shipments

import (
	"go-starter/internal/modules/auth/middlewares"
	authServices "go-starter/internal/modules/auth/services"
	"go-starter/internal/modules/shipments/handlers"
	shipmentRespositories "go-starter/internal/modules/shipments/repositories"
	shipmentServices "go-starter/internal/modules/shipments/services"
	"go-starter/pkg/config"
	"go-starter/pkg/db"
	"go-starter/pkg/ratelimiter"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, api *echo.Group, database *db.Database, cfg *config.Config) {
	jwtService := authServices.NewJWTService()

	// Create rate limiter for SafeCube API
	rateLimiter := ratelimiter.NewSafeCubeAPIRateLimiter()

	safeCubeAPIService := shipmentServices.NewSafeCubeAPIService(
		cfg.SafeCubeAPI.BaseURL,
		cfg.SafeCubeAPI.APIKey,
		rateLimiter,
	)
	shipmentRepository := shipmentRespositories.NewShipmentRepository(database)
	shipmentService := shipmentServices.NewShipmentService(shipmentRepository, safeCubeAPIService)
	shipmentAPIHandler := handlers.NewShipmentAPIHandler(shipmentService, safeCubeAPIService)

	shipmentWEBHandler := handlers.NewShipmentWEBHandler(shipmentService)

	shipmentsAPI := api.Group("/shipments")
	shipmentsAPI.Use(middlewares.JWTMiddleware(jwtService))

	shipmentsAPI.POST("", shipmentAPIHandler.AddShipment)
	shipmentsAPI.GET("/grid-data", shipmentAPIHandler.GetShipmentsForGrid)
	shipmentsAPI.GET("/:id/details", shipmentAPIHandler.GetShipmentDetails)
	shipmentsAPI.GET("/:id/details-html", shipmentWEBHandler.GetShipmentDetailsHTML)
	shipmentsAPI.GET("/:id", shipmentAPIHandler.GetShipmentByID)
	shipmentsAPI.POST("/:id/refresh", shipmentAPIHandler.RefreshShipment)
	shipmentsAPI.DELETE("/:id", shipmentAPIHandler.DeleteUserShipment)
	shipmentsAPI.DELETE("/bulk-delete", shipmentAPIHandler.BulkDeleteUserShipments)

	e.GET("/shipments", shipmentWEBHandler.ViewShipmentPage)
	e.GET("/map", shipmentWEBHandler.ViewMapPage)
}
