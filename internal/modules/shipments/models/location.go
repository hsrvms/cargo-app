package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Location struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Name        string    `json:"name" gorm:"type:varchar(50);not null;uniqueIndex:idx_locations_name"`
	State       *string   `json:"state" gorm:"type:varchar(50)"`
	Country     string    `json:"country" gorm:"type:varchar(50);not null"`
	CountryCode string    `json:"countryCode" gorm:"type:varchar(10);not null"`
	Locode      string    `json:"locode" gorm:"type:varchar(10);uniqueIndex"`
	Latitude    float64   `json:"latitude" gorm:"type:decimal(10,8)"`
	Longitude   float64   `json:"longitude" gorm:"type:decimal(11,8)"`
	Timezone    string    `json:"timezone" gorm:"type:varchar(50)"`
	CreatedAt   time.Time `json:"created_at" gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
}

func (Location) TableName() string {
	return "locations"
}

func (l *Location) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	if l.CreatedAt.IsZero() {
		l.CreatedAt = now
	}
	if l.UpdatedAt.IsZero() {
		l.UpdatedAt = now
	}
	return nil
}

func (l *Location) BeforeUpdate(tx *gorm.DB) error {
	l.UpdatedAt = time.Now()
	return nil
}

func AutoMigrateLocations(db *gorm.DB) error {
	return db.AutoMigrate(&Location{})
}

type ShipmentLocation struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	ShipmentID uuid.UUID `json:"shipment_id" gorm:"type:uuid;not null;index"`
	LocationID uuid.UUID `json:"location_id" gorm:"type:uuid;not null;index"`
	AddedAt    time.Time `json:"added_at" gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`

	Shipment Shipment `gorm:"foreignKey:ShipmentID;constraint:OnDelete:CASCADE"`
	Location Location `gorm:"foreignKey:LocationID;constraint:OnDelete:CASCADE"`
}

func (ShipmentLocation) TableName() string {
	return "shipment_locations"
}

func (sl *ShipmentLocation) BeforeCreate(tx *gorm.DB) error {
	if sl.ID == uuid.Nil {
		sl.ID = uuid.New()
	}
	if sl.AddedAt.IsZero() {
		sl.AddedAt = time.Now()
	}
	return nil
}
