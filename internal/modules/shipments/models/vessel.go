package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Vessel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"not null"`
	Imo       int       `gorm:"uniqueIndex"`
	Mmsi      int       `gorm:"uniqueIndex"`
	CallSign  string
	Flag      string
	CreatedAt time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
}

func (Vessel) TableName() string {
	return "vessels"
}

func (v *Vessel) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	now := time.Now()
	if v.CreatedAt.IsZero() {
		v.CreatedAt = now
	}
	if v.UpdatedAt.IsZero() {
		v.UpdatedAt = now
	}
	return nil
}

func (v *Vessel) BeforeUpdate(tx *gorm.DB) error {
	v.UpdatedAt = time.Now()
	return nil
}

func AutoMigrateVessels(db *gorm.DB) error {
	return db.AutoMigrate(&Vessel{})
}

type ShipmentVessel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	ShipmentID uuid.UUID `gorm:"type:uuid;not null;index"`
	VesselID   uuid.UUID `gorm:"type:uuid;not null;index"`
	AddedAt    time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`

	Shipment Shipment `gorm:"foreignKey:ShipmentID;constraint:OnDelete:CASCADE"`
	Vessel   Vessel   `gorm:"foreignKey:VesselID;constraint:OnDelete:CASCADE"`
}

func (ShipmentVessel) TableName() string {
	return "shipment_vessels"
}

func (sv *ShipmentVessel) BeforeCreate(tx *gorm.DB) error {
	if sv.ID == uuid.Nil {
		sv.ID = uuid.New()
	}
	if sv.AddedAt.IsZero() {
		sv.AddedAt = time.Now()
	}
	return nil
}
