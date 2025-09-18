package handlers

import (
	"go-starter/internal/modules/auth/services"
	"go-starter/internal/modules/filters/dto"
	filterServices "go-starter/internal/modules/filters/services"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type FilterAPIHandler struct {
	filterService *filterServices.FilterService
	validator     *validator.Validate
}

func NewFilterAPIHandler(filterService *filterServices.FilterService) *FilterAPIHandler {
	return &FilterAPIHandler{
		filterService: filterService,
		validator:     validator.New(),
	}
}

// SaveFilter handles POST /api/filters
func (h *FilterAPIHandler) SaveFilter(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := services.GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	var req dto.SaveFilterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := h.validator.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Validation failed: " + err.Error(),
		})
	}

	// Trim and validate filter name
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Filter name cannot be empty",
		})
	}

	response, err := h.filterService.SaveFilter(ctx, userID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to save filter: " + err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, response)
}

// GetFilter handles GET /api/filters/:id
func (h *FilterAPIHandler) GetFilter(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := services.GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	filterIDStr := c.Param("id")
	filterID, err := uuid.Parse(filterIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid filter ID",
		})
	}

	response, err := h.filterService.GetFilter(ctx, userID, filterID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Filter not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve filter: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

// GetFilterByName handles GET /api/filters/by-name/:name
func (h *FilterAPIHandler) GetFilterByName(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := services.GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	filterName := c.Param("name")
	if filterName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Filter name is required",
		})
	}

	response, err := h.filterService.GetFilterByName(ctx, userID, filterName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Filter not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve filter: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

// GetAllFilters handles GET /api/filters
func (h *FilterAPIHandler) GetAllFilters(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := services.GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	response, err := h.filterService.GetAllFilters(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve filters: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

// UpdateFilter handles PUT /api/filters/:id
func (h *FilterAPIHandler) UpdateFilter(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := services.GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	filterIDStr := c.Param("id")
	filterID, err := uuid.Parse(filterIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid filter ID",
		})
	}

	var req dto.UpdateFilterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := h.validator.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Validation failed: " + err.Error(),
		})
	}

	// Trim and validate filter name
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Filter name cannot be empty",
		})
	}

	response, err := h.filterService.UpdateFilter(ctx, userID, filterID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Filter not found",
			})
		}
		if strings.Contains(err.Error(), "already exists") {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update filter: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

// DeleteFilter handles DELETE /api/filters/:id
func (h *FilterAPIHandler) DeleteFilter(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := services.GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	filterIDStr := c.Param("id")
	filterID, err := uuid.Parse(filterIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid filter ID",
		})
	}

	err = h.filterService.DeleteFilter(ctx, userID, filterID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Filter not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete filter: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, dto.DeleteFilterResponse{
		Message: "Filter deleted successfully",
	})
}

// DeleteFilterByName handles DELETE /api/filters/by-name/:name
func (h *FilterAPIHandler) DeleteFilterByName(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := services.GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	filterName := c.Param("name")
	if filterName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Filter name is required",
		})
	}

	err = h.filterService.DeleteFilterByName(ctx, userID, filterName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Filter not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete filter: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, dto.DeleteFilterResponse{
		Message: "Filter deleted successfully",
	})
}

// GetFilterStats handles GET /api/filters/stats
func (h *FilterAPIHandler) GetFilterStats(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := services.GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	stats, err := h.filterService.GetFilterStats(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve filter stats: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, stats)
}
