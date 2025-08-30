package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ShipmentRoute struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	ShipmentID    uuid.UUID  `json:"shipment_id" gorm:"type:uuid;not null;index"`
	LocationID    uuid.UUID  `json:"location_id" gorm:"type:uuid;not null;uniqueIndex:idx_shipment_route"`
	RouteType     string     `json:"route_type" gorm:"type:varchar(10);not null;check:route_type IN ('PREPOL','POL','POD','POSTPOD');uniqueIndex:idx_shipment_route"`
	Date          *time.Time `gorm:"type:timestamptz"`
	Actual        *bool
	PredictiveETA *time.Time `gorm:"type:timestamptz"`
	CreatedAt     time.Time  `json:"created_at" gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`

	Shipment Shipment `gorm:"foreignKey:ShipmentID;constraint:OnDelete:CASCADE"`
	Location Location `gorm:"foreignKey:LocationID;constraint:OnDelete:CASCADE"`
}

func (ShipmentRoute) TableName() string {
	return "shipment_routes"
}

func (sr *ShipmentRoute) BeforeCreate(tx *gorm.DB) error {
	if sr.ID == uuid.Nil {
		sr.ID = uuid.New()
	}
	now := time.Now()
	if sr.CreatedAt.IsZero() {
		sr.CreatedAt = now
	}
	if sr.UpdatedAt.IsZero() {
		sr.UpdatedAt = now
	}
	return nil
}

func (sr *ShipmentRoute) BeforeUpdate(tx *gorm.DB) error {
	sr.UpdatedAt = time.Now()
	return nil
}

func AutoMigrateShipmentRoutes(db *gorm.DB) error {
	return db.AutoMigrate(&ShipmentRoute{})
}
