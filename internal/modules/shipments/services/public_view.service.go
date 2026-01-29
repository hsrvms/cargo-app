package services

import (
	"context"
	"go-starter/internal/modules/shipments/dto"
	"go-starter/internal/modules/shipments/models"
	"go-starter/internal/modules/shipments/repositories"

	"github.com/google/uuid"
)

type PublicViewService interface {
	Create(ctx context.Context, creatorID uuid.UUID, req *dto.PublicViewCreateRequest) (*models.PublicView, error)
	GetByID(ctx context.Context, publicViewID uuid.UUID) (*models.PublicView, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]models.PublicView, error)
	Delete(ctx context.Context, publicViewID, userID uuid.UUID) error
	GetPublicShipmentDetails(ctx context.Context, shipmentID uuid.UUID) (*models.Shipment, error)
}

type publicViewService struct {
	viewRepo     repositories.PublicViewRepository
	shipmentRepo repositories.ShipmentRepository
}

func NewPublicViewService(
	viewRepo repositories.PublicViewRepository,
	shipmentRepo repositories.ShipmentRepository,
) PublicViewService {
	return &publicViewService{
		viewRepo:     viewRepo,
		shipmentRepo: shipmentRepo,
	}
}

func (s *publicViewService) Create(
	ctx context.Context,
	creatorID uuid.UUID,
	req *dto.PublicViewCreateRequest,
) (*models.PublicView, error) {
	publicView, err := s.viewRepo.Create(ctx, creatorID, req.Name, req.ShipmentIDs)
	if err != nil {
		return nil, err
	}

	return publicView, nil
}

func (s *publicViewService) GetByID(ctx context.Context, publicViewID uuid.UUID) (*models.PublicView, error) {
	view, err := s.viewRepo.GetByID(ctx, publicViewID)
	if err != nil {
		return nil, err
	}

	return view, nil
}

func (s *publicViewService) GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]models.PublicView, error) {
	views, err := s.viewRepo.GetAllByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return views, nil
}

func (s *publicViewService) Delete(ctx context.Context, publicViewID, userID uuid.UUID) error {
	err := s.viewRepo.Delete(ctx, publicViewID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *publicViewService) GetPublicShipmentDetails(
	ctx context.Context,
	shipmentID uuid.UUID,
) (*models.Shipment, error) {
	shipmentDetails, err := s.shipmentRepo.GetShipmentWithRelations(ctx, shipmentID)
	if err != nil {
		return nil, err
	}

	return shipmentDetails, err
}
