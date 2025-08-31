package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Container struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Number    string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	IsoCode   string    `gorm:"type:varchar(10);not null"`
	SizeType  string    `gorm:"type:varchar(50);not null"`
	Status    string    `gorm:"type:varchar(50);not null"`
	CreatedAt time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
}

func (Container) TableName() string {
	return "containers"
}

func (c *Container) BeforeCreate(tx *gorm.DB) error {
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

func (c *Container) BeforeUpdate(tx *gorm.DB) error {
	c.UpdatedAt = time.Now()
	return nil
}

type ShipmentContainer struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	ShipmentID  uuid.UUID `gorm:"type:uuid;not null;index"`
	ContainerID uuid.UUID `gorm:"type:uuid;not null;index"`
	AddedAt     time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`

	Shipment  Shipment  `gorm:"foreignKey:ShipmentID;constraint:OnDelete:CASCADE"`
	Container Container `gorm:"foreignKey:ContainerID;constraint:OnDelete:CASCADE"`
}

func (ShipmentContainer) TableName() string {
	return "shipment_containers"
}

func (sc *ShipmentContainer) BeforeCreate(tx *gorm.DB) error {
	if sc.ID == uuid.Nil {
		sc.ID = uuid.New()
	}
	if sc.AddedAt.IsZero() {
		sc.AddedAt = time.Now()
	}
	return nil
}

type ContainerEvent struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ContainerID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	LocationID        uuid.UUID  `gorm:"type:uuid;index"`
	FacilityID        *uuid.UUID `gorm:"type:uuid;index"`
	Description       string     `gorm:"type:text"`
	EventType         *string    `gorm:"type:varchar(100)"`
	EventCode         *string    `gorm:"type:varchar(50)"`
	Status            string     `gorm:"type:varchar(50);not null"`
	Date              time.Time  `gorm:"type:timestamptz;not null"`
	IsActual          bool       `gorm:"type:boolean;default:false"`
	IsAdditionalEvent bool       `gorm:"type:boolean;default:false"`
	RouteType         string     `gorm:"type:varchar(50)"`
	TransportType     *string    `gorm:"type:varchar(50)"`
	VesselID          *uuid.UUID `gorm:"type:uuid;index"`
	Voyage            *string    `gorm:"type:varchar(255)"`
	CreatedAt         time.Time  `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt         time.Time  `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`

	Container Container `gorm:"foreignKey:ContainerID;constraint:OnDelete:CASCADE"`
	Location  Location  `gorm:"foreignKey:LocationID"`
	Facility  *Facility `gorm:"foreignKey:FacilityID"`
	Vessel    *Vessel   `gorm:"foreignKey:VesselID"`
}

func (ContainerEvent) TableName() string {
	return "container_events"
}

func (ce *ContainerEvent) BeforeCreate(tx *gorm.DB) error {
	if ce.ID == uuid.Nil {
		ce.ID = uuid.New()
	}
	now := time.Now()
	if ce.CreatedAt.IsZero() {
		ce.CreatedAt = now
	}
	if ce.UpdatedAt.IsZero() {
		ce.UpdatedAt = now
	}
	return nil
}

func (ce *ContainerEvent) BeforeUpdate(tx *gorm.DB) error {
	ce.UpdatedAt = time.Now()
	return nil
}
