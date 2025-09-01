package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Ais struct {
	ID                       uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ShipmentID               uuid.UUID  `gorm:"type:uuid;not null;index"`
	Status                   string     `gorm:"type:varchar(255)"`
	LastEventDescription     *string    `gorm:"type:text"`
	LastEventDate            *time.Time `gorm:"type:timestamptz"`
	LastEventVoyage          *string    `gorm:"type:varchar(255)"`
	DischargePortName        *string    `gorm:"type:varchar(255)"`
	DischargePortCountryCode *string    `gorm:"type:varchar(10)"`
	DischargePortCode        *string    `gorm:"type:varchar(50)"`
	DischargePortDate        *time.Time `gorm:"type:timestamptz"`
	DischargePortDateLabel   *string    `gorm:"type:varchar(255)"`
	DeparturePortName        *string    `gorm:"type:varchar(255)"`
	DeparturePortCountryCode *string    `gorm:"type:varchar(10)"`
	DeparturePortCode        *string    `gorm:"type:varchar(50)"`
	DeparturePortDate        *time.Time `gorm:"type:timestamptz"`
	DeparturePortDateLabel   *string    `gorm:"type:varchar(255)"`
	ArrivalPortName          *string    `gorm:"type:varchar(255)"`
	ArrivalPortCountryCode   *string    `gorm:"type:varchar(10)"`
	ArrivalPortCode          *string    `gorm:"type:varchar(50)"`
	ArrivalPortDate          *time.Time `gorm:"type:timestamptz"`
	ArrivalPortDateLabel     *string    `gorm:"type:varchar(255)"`
	VesselID                 *uuid.UUID `gorm:"type:uuid"`
	LastVesselPositionLat    *float64   `gorm:"type:decimal(10,8)"`
	LastVesselPositionLng    *float64   `gorm:"type:decimal(11,8)"`
	LastVesselPositionUpdate *time.Time `gorm:"type:timestamptz"`

	CreatedAt time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
}

func (Ais) TableName() string {
	return "ais"
}

func (a *Ais) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	now := time.Now()
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}
	if a.UpdatedAt.IsZero() {
		a.UpdatedAt = now
	}
	return nil
}

func (a *Ais) BeforeUpdate(tx *gorm.DB) error {
	a.UpdatedAt = time.Now()
	return nil
}
