package repositories

import (
	"context"
	"fmt"
	"go-starter/internal/modules/shipments/models"
	"go-starter/pkg/db"
	"log"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PublicViewRepository interface {
	GetDB() *db.Database

	Create(ctx context.Context, creatorID uuid.UUID, name string, shipmentIDs []uuid.UUID) (*models.PublicView, error)
	GetByID(ctx context.Context, publicViewID uuid.UUID) (*models.PublicView, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]models.PublicView, error)
	Delete(ctx context.Context, publicViewID, userID uuid.UUID) error
}

type publicViewRepository struct {
	db *db.Database
}

func NewPublicViewRepository(db *db.Database) PublicViewRepository {
	if err := models.AutoMegratePublicViews(db.DB); err != nil {
		log.Printf("failed to migrate public_views: %s", err)
	}

	if err := models.FixShipmentRouteConstraint(db.DB); err != nil {
		log.Printf("failed to fix shipment route constraint: %s", err)
	}

	return &publicViewRepository{
		db: db,
	}
}

func (pv *publicViewRepository) GetDB() *db.Database {
	return pv.db
}

func (r *publicViewRepository) Create(
	ctx context.Context,
	creatorID uuid.UUID,
	name string,
	shipmentIDs []uuid.UUID,
) (*models.PublicView, error) {
	if len(shipmentIDs) == 0 {
		return nil, fmt.Errorf("cannot create public view: no shipment IDs provided")
	}

	db := r.getDBFromContext(ctx)

	var shipments []models.Shipment
	if err := db.Find(&shipments, shipmentIDs).Error; err != nil {
		return nil, err
	}

	view := models.PublicView{
		Name:        name,
		CreatedByID: creatorID,
		Shipments:   shipments,
	}

	if err := db.Create(&view).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return nil, fmt.Errorf("you already have a view named '%s'", name)
		}
		return nil, err
	}

	return &view, nil
}

func (r *publicViewRepository) GetByID(ctx context.Context, publicViewID uuid.UUID) (*models.PublicView, error) {
	var view models.PublicView
	db := r.getDBFromContext(ctx)

	err := db.Preload("Shipments").
		Preload("Shipments.Locations").
		Preload("Shipments.Routes.Location").
		Preload("Shipments.Vessels").
		Preload("Shipments.Facilities").
		Preload("Shipments.Containers").
		Preload("Shipments.Containers.Events.Location").
		Preload("Shipments.Containers.Events.Facility").
		Preload("Shipments.Containers.Events.Vessel").
		Preload("Shipments.RouteSegments").
		Preload("Shipments.RouteSegments.Points").
		Preload("Shipments.Coordinates").
		Preload("Shipments.Ais").
		First(&view, "id = ?", publicViewID).Error

	if err != nil {
		return nil, err
	}

	return &view, nil
}

func (r *publicViewRepository) GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]models.PublicView, error) {
	db := r.getDBFromContext(ctx)

	var publicViews []models.PublicView

	err := db.WithContext(ctx).
		Where("created_by_id = ?", userID).
		Preload("Shipments").
		Preload("Shipments.Locations").
		Preload("Shipments.Routes.Location").
		Preload("Shipments.Vessels").
		Preload("Shipments.Facilities").
		Preload("Shipments.Containers").
		Preload("Shipments.Containers.Events.Location").
		Preload("Shipments.Containers.Events.Facility").
		Preload("Shipments.Containers.Events.Vessel").
		Preload("Shipments.RouteSegments").
		Preload("Shipments.RouteSegments.Points").
		Preload("Shipments.Coordinates").
		Preload("Shipments.Ais").
		Find(&publicViews).Error

	if err != nil {
		return nil, fmt.Errorf("Failed to fetch shipments for grid: %w", err)
	}

	return publicViews, nil
}

func (r *publicViewRepository) Delete(ctx context.Context, publicViewID, userID uuid.UUID) error {
	db := r.getDBFromContext(ctx)

	var view models.PublicView
	if err := db.Where("id = ? AND created_by_id = ?", publicViewID, userID).First(&view).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("public view not found")
		}
		return fmt.Errorf("failed to find public view: %w", err)
	}

	if err := db.Model(&view).Association("Shipments").Clear(); err != nil {
		return fmt.Errorf("failed to clear shipment associations: %w", err)
	}

	if err := db.Delete(&view).Error; err != nil {
		return fmt.Errorf("failed to delete public view: %w", err)
	}

	return nil
}

func (r *publicViewRepository) getDBFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok {
		return tx
	}
	return r.db.DB
}
