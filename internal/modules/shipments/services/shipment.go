package services

import (
	"context"
	"fmt"
	"go-starter/internal/modules/shipments/dto"
	"go-starter/internal/modules/shipments/models"
	"go-starter/internal/modules/shipments/repositories"
	"log"

	"github.com/google/uuid"
)

type ShipmentService interface {
	AddShipment(ctx context.Context, userID uuid.UUID, req *dto.AddShipmentRequest) (*models.Shipment, error)
	GetShipmentByNumber(ctx context.Context, userID uuid.UUID, shipmentNumber string) (*models.Shipment, error)
	GetShipmentByID(ctx context.Context, userID, shipmentID uuid.UUID) (*models.Shipment, error)
	GetShipmentDetails(ctx context.Context, userID, shipmentID uuid.UUID) (*dto.ShipmentDetailsResponse, error)
	SyncShipment(ctx context.Context, shipmentID uuid.UUID) (*models.Shipment, error)
	RefreshShipment(ctx context.Context, userID, shipmentID uuid.UUID) (*models.Shipment, error)
}

type shipmentService struct {
	repo               repositories.ShipmentRepository
	safeCubeAPIService SafeCubeAPIService
}

func NewShipmentService(repo repositories.ShipmentRepository, safeCubeAPIService SafeCubeAPIService) ShipmentService {
	return &shipmentService{
		repo:               repo,
		safeCubeAPIService: safeCubeAPIService,
	}
}

func (s *shipmentService) AddShipment(
	ctx context.Context,
	userID uuid.UUID,
	req *dto.AddShipmentRequest,
) (*models.Shipment, error) {

	alreadyTracking, err := s.repo.CheckUserAlreadyTracking(ctx, userID, req.ShipmentNumber)
	if err != nil {
		return nil, err
	}
	if alreadyTracking {
		return nil, fmt.Errorf("you are already tracking this shipment")
	}

	exists, err := s.repo.CheckShipmentExists(ctx, req.ShipmentNumber)
	if err != nil {
		return nil, err
	}
	if exists {
		shipment, err := s.repo.AddExistingShipmentToUser(ctx, userID, req.ShipmentNumber)
		if err != nil {
			return nil, err
		}

		shipment, err = s.SyncShipment(ctx, shipment.ID)
		if err != nil {
			return nil, err
		}

		return shipment, nil
	}

	return s.createNewShipmentFromSafeCubeAPI(ctx, userID, req)
}

func (s *shipmentService) createNewShipmentFromSafeCubeAPI(
	ctx context.Context,
	userID uuid.UUID,
	req *dto.AddShipmentRequest,
) (*models.Shipment, error) {
	apiResponse, err := s.safeCubeAPIService.GetShipmentDetails(
		ctx,
		req.ShipmentNumber,
		req.ShipmentType,
		req.SealineCode,
	)
	if err != nil {
		return nil, err
	}

	shipmentModel := &models.Shipment{
		ShipmentNumber: apiResponse.Metadata.ShipmentNumber,
		ShipmentType:   apiResponse.Metadata.ShipmentType,
		SealineCode:    apiResponse.Metadata.Sealine,
		SealineName:    apiResponse.Metadata.SealineName,
		ShippingStatus: apiResponse.Metadata.ShippingStatus,
		Warnings:       apiResponse.Metadata.Warnings,
	}

	shipment, err := s.repo.CreateShipment(ctx, userID, shipmentModel)
	if err != nil {
		return nil, err
	}

	locations := make([]models.Location, 0, len(apiResponse.Locations))
	for _, loc := range apiResponse.Locations {
		location := models.Location{
			Name:        loc.Name,
			State:       loc.State,
			Country:     loc.Country,
			CountryCode: loc.CountryCode,
			Locode:      loc.Locode,
			Latitude:    loc.Coordinates.Lat,
			Longitude:   loc.Coordinates.Lng,
			Timezone:    loc.Timezone,
		}
		locations = append(locations, location)
	}

	for _, location := range locations {
		location, err := s.repo.CreateLocation(ctx, &shipment.ID, &location)
		if err != nil {
			return nil, err
		}
		locations = append(locations, *location)
	}

	routes := map[string]*dto.SafeCubeRoutePoint{
		"PREPOL":  &apiResponse.Route.Prepol,
		"POL":     &apiResponse.Route.Pol,
		"POD":     &apiResponse.Route.Pod,
		"POSTPOD": &apiResponse.Route.Postpod,
	}

	for routeType, r := range routes {
		if r == nil {
			continue
		}

		loc := &models.Location{
			Name:        r.Location.Name,
			State:       r.Location.State,
			Country:     r.Location.Country,
			CountryCode: r.Location.CountryCode,
			Locode:      r.Location.Locode,
			Latitude:    r.Location.Coordinates.Lat,
			Longitude:   r.Location.Coordinates.Lng,
			Timezone:    r.Location.Timezone,
		}
		loc, err := s.repo.CreateLocation(ctx, nil, loc)
		if err != nil {
			return nil, err
		}

		route := &models.ShipmentRoute{
			ShipmentID:    shipment.ID,
			LocationID:    loc.ID,
			RouteType:     routeType,
			Date:          r.Date,
			Actual:        r.Actual,
			PredictiveETA: r.PredictiveEta,
		}

		route, err = s.repo.CreateRoute(ctx, route)
		if err != nil {
			return nil, err
		}
	}

	vessels := make([]models.Vessel, 0, len(apiResponse.Vessels))
	for _, v := range apiResponse.Vessels {
		vessel := models.Vessel{
			Name:     v.Name,
			Imo:      v.Imo,
			Mmsi:     v.Mmsi,
			CallSign: v.CallSign,
			Flag:     v.Flag,
		}
		createdVessel, err := s.repo.CreateVessel(ctx, &shipment.ID, &vessel)
		if err != nil {
			return nil, err
		}
		vessels = append(vessels, *createdVessel)
	}

	facilities := make([]models.Facility, 0, len(apiResponse.Facilities))
	for _, f := range apiResponse.Facilities {
		facility := models.Facility{
			Name:        f.Name,
			CountryCode: f.CountryCode,
			Locode:      f.Locode,
			BicCode:     f.BicCode,
			SmdgCode:    f.SmdgCode,
			Latitude:    f.Coordinates.Lat,
			Longitude:   f.Coordinates.Lng,
		}
		createdFacility, err := s.repo.CreateFacility(ctx, &shipment.ID, &facility)
		if err != nil {
			return nil, err
		}
		facilities = append(facilities, *createdFacility)
	}

	containers := make([]models.Container, 0, len(apiResponse.Containers))
	for _, c := range apiResponse.Containers {
		container := models.Container{
			Number:   c.Number,
			IsoCode:  c.IsoCode,
			SizeType: c.SizeType,
			Status:   c.Status,
		}
		createdContainer, err := s.repo.CreateContainer(ctx, &shipment.ID, &container)
		if err != nil {
			return nil, err
		}
		containers = append(containers, *createdContainer)
	}

	log.Printf("Created shipment %s in database with ID: %s", req.ShipmentNumber, shipment.ID)

	return shipment, nil
}

func (s *shipmentService) GetShipmentByNumber(
	ctx context.Context,
	userID uuid.UUID,
	shipmentNumber string,
) (*models.Shipment, error) {
	shipment, err := s.repo.GetShipmentByNumber(ctx, shipmentNumber)
	if err != nil {
		return nil, err
	}

	return shipment, nil
}

func (s *shipmentService) GetShipmentByID(
	ctx context.Context,
	userID uuid.UUID,
	shipmentID uuid.UUID,
) (*models.Shipment, error) {
	owns, err := s.repo.CheckUserOwnsShipment(ctx, userID, shipmentID)
	if err != nil {
		return nil, err
	}
	if !owns {
		return nil, fmt.Errorf("shipment not found or access denied")
	}

	shipment, err := s.repo.GetShipmentByID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	return shipment, nil
}

func (s *shipmentService) SyncShipment(ctx context.Context, shipmentID uuid.UUID) (*models.Shipment, error) {
	existingShipment, err := s.repo.GetShipmentByID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}

	apiResponse, err := s.safeCubeAPIService.GetShipmentDetails(
		ctx,
		existingShipment.ShipmentNumber,
		existingShipment.ShipmentType,
		existingShipment.SealineCode,
	)
	if err != nil {
		return nil, err
	}

	shipmentModel := &models.Shipment{
		ShipmentNumber: apiResponse.Metadata.ShipmentNumber,
		ShipmentType:   apiResponse.Metadata.ShipmentType,
		SealineCode:    apiResponse.Metadata.Sealine,
		SealineName:    apiResponse.Metadata.SealineName,
		ShippingStatus: apiResponse.Metadata.ShippingStatus,
		Warnings:       apiResponse.Metadata.Warnings,
	}

	shipment, err := s.repo.UpdateShipment(ctx, shipmentID, shipmentModel)
	if err != nil {
		return nil, err
	}

	log.Printf("Updated shipment %s in database with ID: %s", shipment.ShipmentNumber, shipment.ID)

	return shipment, nil
}

func (s *shipmentService) RefreshShipment(ctx context.Context, userID, shipmentID uuid.UUID) (*models.Shipment, error) {
	owns, err := s.repo.CheckUserOwnsShipment(ctx, userID, shipmentID)
	if err != nil {
		return nil, err
	}
	if !owns {
		return nil, fmt.Errorf("shipment not found or access denied")
	}

	shipment, err := s.SyncShipment(ctx, shipmentID)
	if err != nil {
		return nil, err
	}

	return shipment, nil
}

func (s *shipmentService) GetShipmentDetails(ctx context.Context, userID, shipmentID uuid.UUID) (*dto.ShipmentDetailsResponse, error) {
	owns, err := s.repo.CheckUserOwnsShipment(ctx, userID, shipmentID)
	if err != nil {
		return nil, err
	}
	if !owns {
		return nil, fmt.Errorf("shipment not found or access denied")
	}

	shipmentDetails, err := s.repo.GetShipmentDetails(ctx, shipmentID)
	if err != nil {
		return nil, err
	}

	return shipmentDetails, nil

}
