package shipments

import (
	"go-starter/internal/modules/auth/middlewares"
	authServices "go-starter/internal/modules/auth/services"
	"go-starter/internal/modules/shipments/handlers"
	shipmentRespositories "go-starter/internal/modules/shipments/repositories"
	shipmentServices "go-starter/internal/modules/shipments/services"
	"go-starter/pkg/db"
	"os"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, api *echo.Group, database *db.Database) {
	jwtService := authServices.NewJWTService()

	safeCubeBaseURL := os.Getenv("SAFECUBE_API_BASE_URL")
	safeCubeAPIKey := os.Getenv("SAFECUBE_API_KEY")

	safeCubeAPIService := shipmentServices.NewSafeCubeAPIService(
		safeCubeBaseURL,
		safeCubeAPIKey,
	)
	shipmentRepository := shipmentRespositories.NewShipmentRepository(database)
	shipmentService := shipmentServices.NewShipmentService(shipmentRepository, safeCubeAPIService)
	shipmentAPIHandler := handlers.NewShipmentAPIHandler(shipmentService, safeCubeAPIService)

	shipmentWEBHandler := handlers.NewShipmentWEBHandler()

	shipmentsAPI := api.Group("/shipments")
	shipmentsAPI.Use(middlewares.JWTMiddleware(jwtService))

	shipmentsAPI.POST("", shipmentAPIHandler.AddShipment)
	shipmentsAPI.GET("/:id/details", shipmentAPIHandler.GetShipmentDetails)
	shipmentsAPI.GET("/:id/refresh", shipmentAPIHandler.RefreshShipment)
	shipmentsAPI.GET("/:id", shipmentAPIHandler.GetShipmentByID)
	shipmentsAPI.DELETE("/:id", shipmentAPIHandler.DeleteUserShipment)
	shipmentsAPI.DELETE("/bulk-delete", shipmentAPIHandler.BulkDeleteUserShipments)

	e.GET("/shipments", shipmentWEBHandler.ViewShipmentPage)
}
