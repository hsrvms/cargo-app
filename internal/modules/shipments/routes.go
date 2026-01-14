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
	"log"
	"os"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, api *echo.Group, database *db.Database, cfg *config.Config) {
	jwtService := authServices.NewJWTService()

	var safeCubeAPIService shipmentServices.SafeCubeAPIService

	if os.Getenv("USE_DUMMY_API") == "true" {
		safeCubeAPIService = shipmentServices.NewMockSafeCubeAPIService("pkg/db/dummy_response_data.json")
		log.Println("⚠️ USING MOCK SAFECUBE API")
	} else {
		rateLimiter := ratelimiter.NewSafeCubeAPIRateLimiter()
		safeCubeAPIService = shipmentServices.NewSafeCubeAPIService(
			cfg.SafeCubeAPI.BaseURL,
			cfg.SafeCubeAPI.APIKey,
			rateLimiter,
		)

	}

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
	shipmentsAPI.PATCH("/:id/update-info", shipmentAPIHandler.UpdateUserShipmentInfo)
	shipmentsAPI.DELETE("/:id", shipmentAPIHandler.DeleteUserShipment)
	shipmentsAPI.DELETE("/bulk-delete", shipmentAPIHandler.BulkDeleteUserShipments)

	e.GET("/shipments", shipmentWEBHandler.ViewShipmentPage, middlewares.WebJWTMiddleware(jwtService))
	e.GET("/shipments/grid", shipmentWEBHandler.ViewGridOnlyPage, middlewares.WebJWTMiddleware(jwtService))
	e.GET("/map", shipmentWEBHandler.ViewMapPage, middlewares.WebJWTMiddleware(jwtService))
}
