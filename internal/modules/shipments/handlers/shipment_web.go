package handlers

import (
	authService "go-starter/internal/modules/auth/services"
	shipmentServices "go-starter/internal/modules/shipments/services"
	"go-starter/internal/modules/shipments/views"
	"go-starter/internal/modules/shipments/views/components"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type shipmentWEBHandler struct {
	shipmentService shipmentServices.ShipmentService
}

func NewShipmentWEBHandler(shipmentService shipmentServices.ShipmentService) *shipmentWEBHandler {
	return &shipmentWEBHandler{
		shipmentService: shipmentService,
	}
}

func (h *shipmentWEBHandler) ViewShipmentPage(c echo.Context) error {
	component := views.ShipmentPage()
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *shipmentWEBHandler) GetShipmentDetailsHTML(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := authService.GetUserIDFromContext(c)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}

	idStr := c.Param("id")
	shipmentID, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid shipment id",
		})
	}

	shipmentDetails, err := h.shipmentService.GetShipmentDetails(ctx, userID, shipmentID)
	if err != nil {
		return c.String(http.StatusNotFound, "Shipment not found")
	}

	component := components.ShipmentDetails(*shipmentDetails)
	return component.Render(ctx, c.Response().Writer)
}
