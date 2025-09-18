package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FilterData represents the JSON structure for storing filter configurations
type FilterData struct {
	Filters           interface{} `json:"filters,omitempty"`
	Columns           interface{} `json:"columns,omitempty"`
	Sorting           interface{} `json:"sorting,omitempty"`
	ActiveFilterCount int         `json:"activeFilterCount,omitempty"`
	Version           string      `json:"version,omitempty"`
	UserAgent         string      `json:"userAgent,omitempty"`
}

// Scan implements the sql.Scanner interface for FilterData
func (fd *FilterData) Scan(value interface{}) error {
	if value == nil {
		*fd = FilterData{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into FilterData", value)
	}

	return json.Unmarshal(bytes, fd)
}

// Value implements the driver.Valuer interface for FilterData
func (fd FilterData) Value() (driver.Value, error) {
	if fd == (FilterData{}) {
		return nil, nil
	}
	return json.Marshal(fd)
}

// UserFilter represents a saved filter configuration for a specific user
type UserFilter struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	Name       string         `json:"name" gorm:"not null"`
	FilterData FilterData     `json:"filter_data" gorm:"type:jsonb;not null"`
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate is a GORM hook that runs before creating a user filter
func (uf *UserFilter) BeforeCreate(tx *gorm.DB) error {
	if uf.ID == uuid.Nil {
		uf.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name for the UserFilter model
func (UserFilter) TableName() string {
	return "user_filters"
}
