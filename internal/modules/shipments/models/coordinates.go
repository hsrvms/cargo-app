package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Coordinate struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	ShipmentID uuid.UUID `gorm:"type:uuid;not null;index"`
	Latitude   float64   `gorm:"not null"`
	Longitude  float64   `gorm:"not null"`
	CreatedAt  time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
}

func (Coordinate) TableName() string {
	return "coordinates"
}

func (c *Coordinate) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	now := time.Now()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = now
	}
	return nil
}

func (c *Coordinate) BeforeUpdate(tx *gorm.DB) error {
	c.UpdatedAt = time.Now()
	return nil
}
