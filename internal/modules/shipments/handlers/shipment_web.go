package handlers

import (
	"go-starter/internal/modules/shipments/views"

	"github.com/labstack/echo/v4"
)

type ShipmentWEBHandler struct{}

func NewShipmentWEBHandler() *ShipmentWEBHandler {
	return &ShipmentWEBHandler{}
}

func (h *ShipmentWEBHandler) ViewShipmentPage(c echo.Context) error {
	component := views.ShipmentPage()
	return component.Render(c.Request().Context(), c.Response().Writer)
}
