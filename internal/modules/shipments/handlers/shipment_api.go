package handlers

import (
	authService "go-starter/internal/modules/auth/services"
	"go-starter/internal/modules/shipments/dto"
	shipmentServices "go-starter/internal/modules/shipments/services"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type shipmentAPIHandler struct {
	shipmentService    shipmentServices.ShipmentService
	safeCubeAPIService shipmentServices.SafeCubeAPIService
}

func NewShipmentAPIHandler(
	shipmentService shipmentServices.ShipmentService,
	safeCubeAPIService shipmentServices.SafeCubeAPIService,
) *shipmentAPIHandler {
	return &shipmentAPIHandler{
		shipmentService:    shipmentService,
		safeCubeAPIService: safeCubeAPIService,
	}
}

func (h *shipmentAPIHandler) AddShipment(c echo.Context) error {
	userID, err := authService.GetUserIDFromContext(c)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}

	var req dto.AddShipmentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	req.ShipmentNumber = strings.ToUpper(strings.TrimSpace(req.ShipmentNumber))
	req.ShipmentType = strings.ToUpper(strings.TrimSpace(req.ShipmentType))
	req.SealineCode = strings.ToUpper(strings.TrimSpace(req.SealineCode))

	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	shipment, err := h.shipmentService.AddShipment(
		c.Request().Context(),
		userID,
		&req,
	)

	if err != nil {
		if strings.Contains(err.Error(), "already tracking") {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "You are already tracking this shipment",
			})
		}
		if strings.Contains(err.Error(), "rate limit") {
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"error": "API rate limit exceeded. Please try again later",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to add shipment",
		})
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"message":  "Shipment added successfully",
		"shipment": shipment,
	})
}

func (h *shipmentAPIHandler) GetShipmentByNumber(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := authService.GetUserIDFromContext(c)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}

	shipmentNumber := c.Param("shipmentNumber")

	shipment, err := h.shipmentService.GetShipmentByNumber(ctx, userID, shipmentNumber)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to get shipment",
		})
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"message":  "success",
		"shipment": shipment,
	})
}

func (h *shipmentAPIHandler) GetShipmentByID(c echo.Context) error {
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

	shipment, err := h.shipmentService.GetShipmentByID(ctx, userID, shipmentID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to get shipment",
		})
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"message":  "success",
		"shipment": shipment,
	})
}

func (h *shipmentAPIHandler) RefreshShipment(c echo.Context) error {
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

	shipment, err := h.shipmentService.RefreshShipment(ctx, userID, shipmentID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message":  "success",
		"shipment": shipment,
	})

}

func (h *shipmentAPIHandler) GetShipmentDetails(c echo.Context) error {
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
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message":          "success",
		"shipment_details": shipmentDetails,
	})
}
