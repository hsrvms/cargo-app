package handlers

import (
	"fmt"
	authService "go-starter/internal/modules/auth/services"
	"go-starter/internal/modules/shipments/dto"
	shipmentServices "go-starter/internal/modules/shipments/services"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type publicViewAPIHandler struct {
	publicViewService shipmentServices.PublicViewService
}

func NewPublicViewAPIHandler(publicViewService shipmentServices.PublicViewService) *publicViewAPIHandler {
	return &publicViewAPIHandler{
		publicViewService: publicViewService,
	}
}

func (h *publicViewAPIHandler) GetByID(c echo.Context) error {
	idParam := c.Param("id")
	viewID, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid link ID"})
	}

	view, err := h.publicViewService.GetByID(c.Request().Context(), viewID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	response := PublicViewToReadResponse(view)

	return c.JSON(http.StatusOK, response)
}

func (h *publicViewAPIHandler) GetAllByUserID(c echo.Context) error {
	userID, err := authService.GetUserIDFromContext(c)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}

	views, err := h.publicViewService.GetAllByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	response := make([]dto.PublicViewReadResponse, 0, len(views))
	for _, v := range views {
		response = append(response, PublicViewToReadResponse(&v))
	}

	return c.JSON(http.StatusOK, response)
}

func (h *publicViewAPIHandler) CreatePublicView(c echo.Context) error {
	userID, err := authService.GetUserIDFromContext(c)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}

	var req dto.PublicViewCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	req.Name = strings.TrimSpace(req.Name)

	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	publicView, err := h.publicViewService.Create(
		c.Request().Context(),
		userID,
		&req,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	baseUrl := os.Getenv("APP_BASE_URL")
	if baseUrl == "" {
		baseUrl = "http://localhost:8080"
	}

	shareLink := fmt.Sprintf("%s/shipments/public/%s", baseUrl, publicView.ID)

	response := dto.PublicViewCreateResponse{
		ID:        publicView.ID,
		Name:      publicView.Name,
		ExpiresAt: publicView.ExpiresAt,
		ShareLink: shareLink,
	}

	return c.JSON(http.StatusCreated, response)
}

func (h *publicViewAPIHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := authService.GetUserIDFromContext(c)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}

	idStr := c.Param("id")
	publicViewID, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid public view id"})
	}

	err = h.publicViewService.Delete(ctx, publicViewID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.NoContent(http.StatusOK)
}
