package models

import (
	authModels "go-starter/internal/modules/auth/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserShipment represents the many-to-many relationship between users and shipments
type UserShipment struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	ShipmentID uuid.UUID `json:"shipment_id" gorm:"type:uuid;not null;index"`
	AddedAt    time.Time `json:"added_at" gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`

	// User-specific shipment information
	Recipient string `json:"recipient" gorm:"type:varchar(255)"`
	Address   string `json:"address" gorm:"type:text"`
	Notes     string `json:"notes" gorm:"type:text"`

	// Foreign key relationships
	User     authModels.User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"user"`
	Shipment Shipment        `gorm:"foreignKey:ShipmentID;references:ID;constraint:OnDelete:CASCADE" json:"shipment"`
}

// TableName specifies the table name for UserShipment
func (UserShipment) TableName() string {
	return "user_shipments"
}

// BeforeCreate hook to set UUID if not provided
func (us *UserShipment) BeforeCreate(tx *gorm.DB) error {
	if us.ID == uuid.Nil {
		us.ID = uuid.New()
	}
	if us.AddedAt.IsZero() {
		us.AddedAt = time.Now()
	}
	return nil
}

func AutoMigrateUserShipments(db *gorm.DB) error {
	return db.AutoMigrate(&UserShipment{})
}

func AddUniqueConstraint(db *gorm.DB) error {
	return db.Exec(`
        ALTER TABLE user_shipments
        ADD CONSTRAINT IF NOT EXISTS user_shipments_unique
        UNIQUE (user_id, shipment_id);
    `).Error
}

// Usage examples:

// Create a user-shipment relationship
func CreateUserShipmentRelation(db *gorm.DB, userID, shipmentID uuid.UUID, recipient, address, notes string) error {
	userShipment := &UserShipment{
		UserID:     userID,
		ShipmentID: shipmentID,
		Recipient:  recipient,
		Address:    address,
		Notes:      notes,
	}

	return db.Create(userShipment).Error
}

// Get all shipments for a user
func GetUserShipments(db *gorm.DB, userID uuid.UUID) ([]UserShipment, error) {
	var userShipments []UserShipment
	err := db.Where("user_id = ?", userID).
		Preload("Shipment").
		Find(&userShipments).Error
	return userShipments, err
}

// Get all users for a shipment
func GetShipmentUsers(db *gorm.DB, shipmentID uuid.UUID) ([]UserShipment, error) {
	var userShipments []UserShipment
	err := db.Where("shipment_id = ?", shipmentID).
		Preload("User").
		Find(&userShipments).Error
	return userShipments, err
}

// Check if user-shipment relationship exists
func UserShipmentExists(db *gorm.DB, userID, shipmentID uuid.UUID) (bool, error) {
	var count int64
	err := db.Model(&UserShipment{}).
		Where("user_id = ? AND shipment_id = ?", userID, shipmentID).
		Count(&count).Error
	return count > 0, err
}

// Remove user-shipment relationship
func RemoveUserShipment(db *gorm.DB, userID, shipmentID uuid.UUID) error {
	return db.Where("user_id = ? AND shipment_id = ?", userID, shipmentID).
		Delete(&UserShipment{}).Error
}
