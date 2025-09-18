package dto

import (
	"go-starter/internal/modules/filters/models"
	"time"

	"github.com/google/uuid"
)

// SaveFilterRequest represents the request to save a filter
type SaveFilterRequest struct {
	Name       string            `json:"name" validate:"required,min=1,max=100"`
	FilterData models.FilterData `json:"filter_data" validate:"required"`
}

// UpdateFilterRequest represents the request to update a filter
type UpdateFilterRequest struct {
	Name       string            `json:"name" validate:"required,min=1,max=100"`
	FilterData models.FilterData `json:"filter_data" validate:"required"`
}

// FilterResponse represents the response when returning a filter
type FilterResponse struct {
	ID         uuid.UUID         `json:"id"`
	UserID     uuid.UUID         `json:"user_id"`
	Name       string            `json:"name"`
	FilterData models.FilterData `json:"filter_data"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// FiltersListResponse represents the response when listing filters
type FiltersListResponse struct {
	Filters []FilterResponse `json:"filters"`
	Total   int              `json:"total"`
}

// SaveFilterResponse represents the response after saving a filter
type SaveFilterResponse struct {
	ID      uuid.UUID `json:"id"`
	Message string    `json:"message"`
}

// DeleteFilterResponse represents the response after deleting a filter
type DeleteFilterResponse struct {
	Message string `json:"message"`
}

// ToFilterResponse converts a UserFilter model to FilterResponse DTO
func ToFilterResponse(filter *models.UserFilter) FilterResponse {
	return FilterResponse{
		ID:         filter.ID,
		UserID:     filter.UserID,
		Name:       filter.Name,
		FilterData: filter.FilterData,
		CreatedAt:  filter.CreatedAt,
		UpdatedAt:  filter.UpdatedAt,
	}
}

// ToFiltersListResponse converts a slice of UserFilter models to FiltersListResponse DTO
func ToFiltersListResponse(filters []models.UserFilter) FiltersListResponse {
	filterResponses := make([]FilterResponse, len(filters))
	for i, filter := range filters {
		filterResponses[i] = ToFilterResponse(&filter)
	}

	return FiltersListResponse{
		Filters: filterResponses,
		Total:   len(filterResponses),
	}
}
