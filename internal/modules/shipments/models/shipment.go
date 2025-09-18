package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Shipment struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	ShipmentNumber string         `json:"shipment_number" gorm:"type:varchar(50);not null;uniqueIndex:idx_shipments_shipment_number"`
	ShipmentType   string         `json:"shipment_type" gorm:"type:varchar(10);not null"`
	SealineCode    string         `json:"sealine_code" gorm:"type:varchar(10)"`
	SealineName    string         `json:"sealine_name" gorm:"type:varchar(100)"`
	ShippingStatus string         `json:"shipping_status" gorm:"type:varchar(50);not null"`
	CreatedAt      time.Time      `json:"created_at" gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	Warnings       pq.StringArray `json:"warnings" gorm:"type:text[];default:'{}'"`

	// Shipment information
	Consignee        string `json:"consignee" gorm:"type:varchar(255)"`
	Recipient        string `json:"recipient" gorm:"type:varchar(255)"`
	AssignedTo       string `json:"assigned_to" gorm:"type:varchar(255)"`
	PlaceOfLoading   string `json:"place_of_loading" gorm:"type:varchar(255)"`
	PlaceOfDelivery  string `json:"place_of_delivery" gorm:"type:varchar(255)"`
	FinalDestination string `json:"final_destination" gorm:"type:text"`
	ContainerType    string `json:"container_type" gorm:"type:varchar(100)"`
	Shipper          string `json:"shipper" gorm:"type:varchar(255)"`
	InvoiceAmount    string `json:"invoice_amount" gorm:"type:varchar(100)"`
	Cost             string `json:"cost" gorm:"type:varchar(100)"`
	Customs          string `json:"customs" gorm:"type:varchar(255)"`
	MBL              string `json:"mbl" gorm:"type:varchar(100)"`
	Notes            string `json:"notes" gorm:"type:text"`

	// Boolean fields
	CustomsProcessed bool `json:"customs_processed" gorm:"type:boolean;default:false"`
	Invoiced         bool `json:"invoiced" gorm:"type:boolean;default:false"`
	PaymentReceived  bool `json:"payment_received" gorm:"type:boolean;default:false"`
}

// TableName specifies the table name for Shipment
func (Shipment) TableName() string {
	return "shipments"
}

func (s *Shipment) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	now := time.Now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = now
	}
	if len(s.Warnings) == 0 {
		s.Warnings = pq.StringArray{}
	}
	return nil
}

func (s *Shipment) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now()
	return nil
}

func AutoMigrateShipments(db *gorm.DB) error {
	return db.AutoMigrate(&Shipment{})
}

// Repository methods for common operations

// CreateShipment creates a new shipment
func CreateShipment(db *gorm.DB, shipment *Shipment) error {
	return db.Create(shipment).Error
}

// UpdateShipment updates an existing shipment
func UpdateShipment(db *gorm.DB, shipment *Shipment) error {
	return db.Save(shipment).Error
}

// UpdateShipmentAPIData updates API-related fields
func UpdateShipmentAPIData(db *gorm.DB, shipmentID uuid.UUID, status string, warnings pq.StringArray) error {
	return db.Model(&Shipment{}).
		Where("id = ?", shipmentID).
		Updates(map[string]any{
			"shipping_status": status,
			"last_api_fetch":  time.Now(),
			"api_warnings":    warnings,
			"updated_at":      time.Now(),
		}).Error
}

// GetShipmentsByUser retrieves all shipments for a specific user
func GetShipmentsByUser(db *gorm.DB, userID uuid.UUID) ([]Shipment, error) {
	var shipments []Shipment
	err := db.Joins("JOIN user_shipments ON user_shipments.shipment_id = shipments.id").
		Where("user_shipments.user_id = ?", userID).
		Find(&shipments).Error
	return shipments, err
}

// GetShipmentsWithUsers retrieves shipments with their associated users
func GetShipmentsWithUsers(db *gorm.DB) ([]Shipment, error) {
	var shipments []Shipment
	err := db.Preload("Users").Find(&shipments).Error
	return shipments, err
}

// DeleteShipment deletes a shipment (will cascade to user_shipments due to foreign key constraint)
func DeleteShipment(db *gorm.DB, id uuid.UUID) error {
	return db.Delete(&Shipment{}, id).Error
}

// ShipmentExists checks if a shipment exists by shipment number
func ShipmentExists(db *gorm.DB, shipmentNumber string) (bool, error) {
	var count int64
	err := db.Model(&Shipment{}).Where("shipment_number = ?", shipmentNumber).Count(&count).Error
	return count > 0, err
}

// GetShipmentsByStatus retrieves shipments by shipping status
func GetShipmentsByStatus(db *gorm.DB, status string) ([]Shipment, error) {
	var shipments []Shipment
	err := db.Where("shipping_status = ?", status).Find(&shipments).Error
	return shipments, err
}

// GetShipmentsNeedingAPIUpdate retrieves shipments that need API updates
func GetShipmentsNeedingAPIUpdate(db *gorm.DB, olderThan time.Time) ([]Shipment, error) {
	var shipments []Shipment
	err := db.Where("last_api_fetch IS NULL OR last_api_fetch < ?", olderThan).
		Find(&shipments).Error
	return shipments, err
}
