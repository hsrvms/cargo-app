package services

import (
	"context"
	"go-starter/internal/modules/shipments/dto"
	shipmentsDto "go-starter/internal/modules/shipments/dto"
	"go-starter/internal/modules/shipments/models"

	"github.com/google/uuid"
)

type ShipmentService interface {
	AddShipment(ctx context.Context, userID uuid.UUID, req *dto.AddShipmentRequest) (*models.Shipment, error)
	GetShipmentByNumber(ctx context.Context, userID uuid.UUID, shipmentNumber string) (*models.Shipment, error)
	GetShipmentByID(ctx context.Context, userID, shipmentID uuid.UUID) (*models.Shipment, error)
	GetShipmentDetails(ctx context.Context, userID, shipmentID uuid.UUID) (*dto.ShipmentDetailsResponse, error)
	UpdateShipmentInfo(ctx context.Context, userID, shipmentID uuid.UUID, recipient, address, notes string) error
	SyncShipment(ctx context.Context, userID, shipmentID uuid.UUID) (*models.Shipment, error)
	RefreshShipment(ctx context.Context, userID, shipmentID uuid.UUID) (*models.Shipment, error)
	SystemRefreshShipment(ctx context.Context, shipmentID uuid.UUID) (*models.Shipment, error)
	GetShipmentsForGrid(ctx context.Context, userID uuid.UUID) (*dto.GridDataResponse, error)
	DeleteUserShipment(ctx context.Context, userID, shipmentID uuid.UUID) error
	BulkDeleteUserShipments(ctx context.Context, userID uuid.UUID, shipmentIDs []uuid.UUID) error
}

type SafeCubeAPIService interface {
	GetShipmentDetails(ctx context.Context, shipmentNumber, shipmentType, sealine string) (*shipmentsDto.SafeCubeAPIShipmentResponse, error)
}
