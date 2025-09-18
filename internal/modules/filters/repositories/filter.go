package repositories

import (
	"context"
	"go-starter/internal/modules/filters/models"
	"go-starter/pkg/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FilterRepository struct {
	db *db.Database
}

func NewRepository(database *db.Database) *FilterRepository {
	return &FilterRepository{
		db: database,
	}
}

// CreateFilter creates a new filter for a user
func (r *FilterRepository) CreateFilter(ctx context.Context, filter *models.UserFilter) error {
	return r.db.DB.WithContext(ctx).Create(filter).Error
}

// GetFilterByID retrieves a filter by its ID and user ID
func (r *FilterRepository) GetFilterByID(ctx context.Context, filterID, userID uuid.UUID) (*models.UserFilter, error) {
	var filter models.UserFilter
	err := r.db.DB.WithContext(ctx).
		Where("id = ? AND user_id = ?", filterID, userID).
		First(&filter).Error

	if err != nil {
		return nil, err
	}

	return &filter, nil
}

// GetFilterByName retrieves a filter by its name and user ID
func (r *FilterRepository) GetFilterByName(ctx context.Context, name string, userID uuid.UUID) (*models.UserFilter, error) {
	var filter models.UserFilter
	err := r.db.DB.WithContext(ctx).
		Where("name = ? AND user_id = ?", name, userID).
		First(&filter).Error

	if err != nil {
		return nil, err
	}

	return &filter, nil
}

// GetFiltersByUserID retrieves all filters for a user
func (r *FilterRepository) GetFiltersByUserID(ctx context.Context, userID uuid.UUID) ([]models.UserFilter, error) {
	var filters []models.UserFilter
	err := r.db.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&filters).Error

	if err != nil {
		return nil, err
	}

	return filters, nil
}

// UpdateFilter updates an existing filter
func (r *FilterRepository) UpdateFilter(ctx context.Context, filter *models.UserFilter) error {
	return r.db.DB.WithContext(ctx).
		Model(filter).
		Where("id = ? AND user_id = ?", filter.ID, filter.UserID).
		Updates(map[string]interface{}{
			"name":        filter.Name,
			"filter_data": filter.FilterData,
			"updated_at":  gorm.Expr("NOW()"),
		}).Error
}

// DeleteFilter deletes a filter by its ID and user ID
func (r *FilterRepository) DeleteFilter(ctx context.Context, filterID, userID uuid.UUID) error {
	result := r.db.DB.WithContext(ctx).
		Where("id = ? AND user_id = ?", filterID, userID).
		Delete(&models.UserFilter{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// DeleteFilterByName deletes a filter by its name and user ID
func (r *FilterRepository) DeleteFilterByName(ctx context.Context, name string, userID uuid.UUID) error {
	result := r.db.DB.WithContext(ctx).
		Where("name = ? AND user_id = ?", name, userID).
		Delete(&models.UserFilter{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// FilterExists checks if a filter with the given name exists for a user
func (r *FilterRepository) FilterExists(ctx context.Context, name string, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.DB.WithContext(ctx).
		Model(&models.UserFilter{}).
		Where("name = ? AND user_id = ?", name, userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetFilterCount returns the total number of filters for a user
func (r *FilterRepository) GetFilterCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.DB.WithContext(ctx).
		Model(&models.UserFilter{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	return count, err
}
