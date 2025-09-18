package services

import (
	"context"
	"fmt"
	"go-starter/internal/modules/filters/dto"
	"go-starter/internal/modules/filters/models"
	"go-starter/internal/modules/filters/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FilterService struct {
	filterRepo *repositories.FilterRepository
}

func NewFilterService(filterRepo *repositories.FilterRepository) *FilterService {
	return &FilterService{
		filterRepo: filterRepo,
	}
}

// SaveFilter saves a new filter or updates an existing one
func (s *FilterService) SaveFilter(ctx context.Context, userID uuid.UUID, req *dto.SaveFilterRequest) (*dto.SaveFilterResponse, error) {
	// Check if filter with this name already exists
	existingFilter, err := s.filterRepo.GetFilterByName(ctx, req.Name, userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("error checking existing filter: %w", err)
	}

	// If filter exists, update it
	if existingFilter != nil {
		existingFilter.FilterData = req.FilterData
		if err := s.filterRepo.UpdateFilter(ctx, existingFilter); err != nil {
			return nil, fmt.Errorf("error updating filter: %w", err)
		}

		return &dto.SaveFilterResponse{
			ID:      existingFilter.ID,
			Message: fmt.Sprintf("Filter '%s' updated successfully", req.Name),
		}, nil
	}

	// Create new filter
	filter := &models.UserFilter{
		UserID:     userID,
		Name:       req.Name,
		FilterData: req.FilterData,
	}

	if err := s.filterRepo.CreateFilter(ctx, filter); err != nil {
		return nil, fmt.Errorf("error creating filter: %w", err)
	}

	return &dto.SaveFilterResponse{
		ID:      filter.ID,
		Message: fmt.Sprintf("Filter '%s' saved successfully", req.Name),
	}, nil
}

// GetFilter retrieves a filter by ID
func (s *FilterService) GetFilter(ctx context.Context, userID, filterID uuid.UUID) (*dto.FilterResponse, error) {
	filter, err := s.filterRepo.GetFilterByID(ctx, filterID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("filter not found")
		}
		return nil, fmt.Errorf("error retrieving filter: %w", err)
	}

	response := dto.ToFilterResponse(filter)
	return &response, nil
}

// GetFilterByName retrieves a filter by name
func (s *FilterService) GetFilterByName(ctx context.Context, userID uuid.UUID, name string) (*dto.FilterResponse, error) {
	filter, err := s.filterRepo.GetFilterByName(ctx, name, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("filter not found")
		}
		return nil, fmt.Errorf("error retrieving filter: %w", err)
	}

	response := dto.ToFilterResponse(filter)
	return &response, nil
}

// GetAllFilters retrieves all filters for a user
func (s *FilterService) GetAllFilters(ctx context.Context, userID uuid.UUID) (*dto.FiltersListResponse, error) {
	filters, err := s.filterRepo.GetFiltersByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving filters: %w", err)
	}

	response := dto.ToFiltersListResponse(filters)
	return &response, nil
}

// UpdateFilter updates an existing filter
func (s *FilterService) UpdateFilter(ctx context.Context, userID, filterID uuid.UUID, req *dto.UpdateFilterRequest) (*dto.FilterResponse, error) {
	// First, get the existing filter to ensure it belongs to the user
	existingFilter, err := s.filterRepo.GetFilterByID(ctx, filterID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("filter not found")
		}
		return nil, fmt.Errorf("error retrieving filter: %w", err)
	}

	// Check if another filter with the same name exists (but different ID)
	if existingFilter.Name != req.Name {
		nameExists, err := s.filterRepo.FilterExists(ctx, req.Name, userID)
		if err != nil {
			return nil, fmt.Errorf("error checking filter name: %w", err)
		}
		if nameExists {
			return nil, fmt.Errorf("filter with name '%s' already exists", req.Name)
		}
	}

	// Update the filter
	existingFilter.Name = req.Name
	existingFilter.FilterData = req.FilterData

	if err := s.filterRepo.UpdateFilter(ctx, existingFilter); err != nil {
		return nil, fmt.Errorf("error updating filter: %w", err)
	}

	response := dto.ToFilterResponse(existingFilter)
	return &response, nil
}

// DeleteFilter deletes a filter by ID
func (s *FilterService) DeleteFilter(ctx context.Context, userID, filterID uuid.UUID) error {
	if err := s.filterRepo.DeleteFilter(ctx, filterID, userID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("filter not found")
		}
		return fmt.Errorf("error deleting filter: %w", err)
	}

	return nil
}

// DeleteFilterByName deletes a filter by name
func (s *FilterService) DeleteFilterByName(ctx context.Context, userID uuid.UUID, name string) error {
	if err := s.filterRepo.DeleteFilterByName(ctx, name, userID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("filter not found")
		}
		return fmt.Errorf("error deleting filter: %w", err)
	}

	return nil
}

// GetFilterStats returns statistics about user's filters
func (s *FilterService) GetFilterStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	count, err := s.filterRepo.GetFilterCount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting filter count: %w", err)
	}

	return map[string]interface{}{
		"total_filters": count,
	}, nil
}
