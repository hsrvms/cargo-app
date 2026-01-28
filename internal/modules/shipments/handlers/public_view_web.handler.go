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

type publicViewWEBHandler struct {
	publicViewService shipmentServices.PublicViewService
	shipmentService   shipmentServices.ShipmentService
}

func NewPublicViewWebHandler(
	publicViewService shipmentServices.PublicViewService,
	shipmentService shipmentServices.ShipmentService,
) *publicViewWEBHandler {
	return &publicViewWEBHandler{
		publicViewService: publicViewService,
		shipmentService:   shipmentService,
	}
}

func (h *publicViewWEBHandler) ViewPublicGridPage(c echo.Context) error {
	idStr := c.Param("id")
	viewID, err := uuid.Parse(idStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid Link")
	}

	view, err := h.publicViewService.GetByID(c.Request().Context(), viewID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Public View Not Found")
	}

	dto := PublicViewToReadResponse(view)

	component := views.PublicShipmentsGrid(dto)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *publicViewWEBHandler) GetPublicShipmentDetailsHTML(c echo.Context) error {
	ctx := c.Request().Context()

	idStr := c.Param("id")
	shipmentID, err := uuid.Parse(idStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid shipment id")
	}

	shipmentDetails, err := h.publicViewService.GetPublicShipmentDetails(ctx, shipmentID)
	if err != nil {
		return c.String(http.StatusNotFound, "Shipment not found")
	}

	shipmentDetailsDTO := ShipmentToDetailsResponse(*shipmentDetails)

	component := components.PublicShipmentDetails(shipmentDetailsDTO)
	return component.Render(ctx, c.Response().Writer)
}

func (h *publicViewWEBHandler) GetUserViewsList(c echo.Context) error {
	userID, err := authService.GetUserIDFromContext(c)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}

	views, err := h.publicViewService.GetAllByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error fetching views")
	}

	component := components.PublicViewList(views)
	return component.Render(c.Request().Context(), c.Response().Writer)
}
