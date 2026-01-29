package models

import (
	authModels "go-starter/internal/modules/auth/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PublicView struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Name      string    `json:"name" gorm:"type:varchar(50);not nul;uniqueIndex:idx_user_view_name"`
	ExpiresAt time.Time `json:"expires_at" gorm:"type:timestamptz"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`

	CreatedByID uuid.UUID       `json:"created_by_id" gorm:"uniqueIndex:idx_user_view_name"`
	CreatedBy   authModels.User `json:"created_by" gorm:"foreignKey:CreatedByID"`

	Shipments []Shipment `json:"shipments" gorm:"many2many:public_view_shipments;"`
}

func (PublicView) TableName() string {
	return "public_view"
}

func (pv *PublicView) BeforeCreate(tx *gorm.DB) error {
	if pv.ID == uuid.Nil {
		pv.ID = uuid.New()
	}

	if pv.ExpiresAt.IsZero() {
		pv.ExpiresAt = time.Now().AddDate(0, 1, 0)
	}

	return nil
}

func (pv *PublicView) BeforeUpdate(tx *gorm.DB) error {
	pv.UpdatedAt = time.Now()
	return nil
}

func AutoMegratePublicViews(db *gorm.DB) error {
	return db.AutoMigrate(&PublicView{})
}
