package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RouteSegment struct {
	ID           uuid.UUID           `gorm:"type:uuid;primaryKey"`
	ShipmentID   uuid.UUID           `gorm:"type:uuid;not null;index"`
	RouteType    string              `gorm:"type:varchar(10);not null"` // SEA, LAND, etc.
	SegmentOrder int                 `gorm:"not null"`
	Points       []RouteSegmentPoint `gorm:"foreignKey:SegmentID;constraint:OnDelete:CASCADE"`
	CreatedAt    time.Time           `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time           `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
}

func (RouteSegment) TableName() string {
	return "route_segments"
}

func (rs *RouteSegment) BeforeCreate(tx *gorm.DB) error {
	if rs.ID == uuid.Nil {
		rs.ID = uuid.New()
	}
	now := time.Now()
	if rs.CreatedAt.IsZero() {
		rs.CreatedAt = now
	}
	if rs.UpdatedAt.IsZero() {
		rs.UpdatedAt = now
	}
	return nil
}

func (rs *RouteSegment) BeforeUpdate(tx *gorm.DB) error {
	rs.UpdatedAt = time.Now()
	return nil
}

type RouteSegmentPoint struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	SegmentID  uuid.UUID `gorm:"type:uuid;not null;index"`
	Latitude   float64   `gorm:"not null"`
	Longitude  float64   `gorm:"not null"`
	PointOrder int       `gorm:"not null"`
}

func (RouteSegmentPoint) TableName() string {
	return "route_segment_points"
}

func (rsp *RouteSegmentPoint) BeforeCreate(tx *gorm.DB) error {
	if rsp.ID == uuid.Nil {
		rsp.ID = uuid.New()
	}
	return nil
}
