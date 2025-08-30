package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Facility struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"type:varchar(50);not null;uniqueIndex:idx_facilities_name"`
	CountryCode string    `gorm:"type:varchar(10);not null"`
	Locode      string    `gorm:"type:varchar(50)"`
	BicCode     *string   `gorm:"type:varchar(50)"`
	SmdgCode    *string   `gorm:"type:varchar(50)"`
	Latitude    float64   `gorm:"type:decimal(10,8)"`
	Longitude   float64   `gorm:"type:decimal(11,8)"`
	CreatedAt   time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
}

func (Facility) TableName() string {
	return "facilities"
}

func (f *Facility) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	now := time.Now()
	if f.CreatedAt.IsZero() {
		f.CreatedAt = now
	}
	if f.UpdatedAt.IsZero() {
		f.UpdatedAt = now
	}
	return nil
}

func (f *Facility) BeforeUpdate(tx *gorm.DB) error {
	f.UpdatedAt = time.Now()
	return nil
}

type ShipmentFacility struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	ShipmentID uuid.UUID `gorm:"type:uuid;not null;index"`
	FacilityID uuid.UUID `gorm:"type:uuid;not null;index"`
	AddedAt    time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`

	Shipment Shipment `gorm:"foreignKey:ShipmentID;constraint:OnDelete:CASCADE"`
	Facility Facility `gorm:"foreignKey:FacilityID;constraint:OnDelete:CASCADE"`
}

func (ShipmentFacility) TableName() string {
	return "shipment_facilities"
}

func (sf *ShipmentFacility) BeforeCreate(tx *gorm.DB) error {
	if sf.ID == uuid.Nil {
		sf.ID = uuid.New()
	}
	if sf.AddedAt.IsZero() {
		sf.AddedAt = time.Now()
	}
	return nil
}
