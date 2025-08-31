package repositories

import (
	"context"
	"errors"
	"fmt"
	"go-starter/internal/modules/shipments/dto"
	"go-starter/internal/modules/shipments/models"
	"go-starter/pkg/db"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ShipmentRepository interface {
	NewTx() *gorm.DB
	CreateShipment(ctx context.Context, userID uuid.UUID, shipment *models.Shipment) (*models.Shipment, error)
	GetShipmentByNumber(ctx context.Context, shipmentNumber string) (*models.Shipment, error)
	GetShipmentByID(ctx context.Context, id uuid.UUID) (*models.Shipment, error)
	CheckUserAlreadyTracking(ctx context.Context, userID uuid.UUID, shipmentNumber string) (bool, error)
	CheckUserOwnsShipment(ctx context.Context, userID, shipmentID uuid.UUID) (bool, error)
	CheckShipmentExists(ctx context.Context, shipmentNumber string) (bool, error)
	AddExistingShipmentToUser(ctx context.Context, userID uuid.UUID, shipmentNumber string) (*models.Shipment, error)
	UpdateShipment(ctx context.Context, id uuid.UUID, shipment *models.Shipment) (*models.Shipment, error)

	CreateLocation(ctx context.Context, shipmentID *uuid.UUID, location *models.Location) (*models.Location, error)
	FindLocationByLocode(ctx context.Context, locode string) (*models.Location, error)
	FindLocationByID(ctx context.Context, id uuid.UUID) (*models.Location, error)

	CreateRoute(ctx context.Context, route *models.ShipmentRoute) (*models.ShipmentRoute, error)

	CreateVessel(ctx context.Context, shipmentID *uuid.UUID, vessel *models.Vessel) (*models.Vessel, error)
	FindVesselByIMOAndMMSI(ctx context.Context, imo, mmsi int) (*models.Vessel, error)
	FindVesselByID(ctx context.Context, id *uuid.UUID) (*models.Vessel, error)

	CreateFacility(ctx context.Context, shipmentID *uuid.UUID, facility *models.Facility) (*models.Facility, error)
	FindFacilityByLocode(ctx context.Context, locode string) (*models.Facility, error)
	FindFacilityByID(ctx context.Context, id *uuid.UUID) (*models.Facility, error)

	CreateContainer(ctx context.Context, shipmentID *uuid.UUID, container *models.Container) (*models.Container, error)
	CreateContainerEvent(ctx context.Context, containerEvent *models.ContainerEvent) (*models.ContainerEvent, error)

	CreateRouteSegment(ctx context.Context, routeSegment *models.RouteSegment) (*models.RouteSegment, error)
	CreateRouteSegmentPoint(ctx context.Context, point *models.RouteSegmentPoint) (*models.RouteSegmentPoint, error)

	CreateCoordinate(ctx context.Context, coordinate *models.Coordinate) (*models.Coordinate, error)

	CreateAis(ctx context.Context, ais *models.Ais) (*models.Ais, error)
	GetShipmentAisData(ctx context.Context, shipmentID uuid.UUID) (*dto.ShipmentAisResponse, error)

	GetShipmentDetails(ctx context.Context, shipmentID uuid.UUID) (*dto.ShipmentDetailsResponse, error)
}

type shipmentRepository struct {
	db *db.Database
}

func NewShipmentRepository(db *db.Database) ShipmentRepository {
	if err := models.AutoMigrateShipments(db.DB); err != nil {
		log.Printf("failed to migrate shipments: %s", err)
	}

	return &shipmentRepository{
		db: db,
	}
}

func (r *shipmentRepository) NewTx() *gorm.DB {
	tx := r.db.DB.Begin()
	return tx
}

func (r *shipmentRepository) CreateShipment(
	ctx context.Context,
	userID uuid.UUID,
	shipment *models.Shipment,
) (*models.Shipment, error) {
	if err := r.db.DB.WithContext(ctx).Create(&shipment).Error; err != nil {
		return nil, fmt.Errorf("failed to create shipment: %w", err)
	}

	link := models.UserShipment{
		UserID:     userID,
		ShipmentID: shipment.ID,
	}
	if err := r.db.DB.WithContext(ctx).Create(&link).Error; err != nil {
		return nil, fmt.Errorf("failed to link shipment to user: %w", err)
	}

	return shipment, nil
}

func (r *shipmentRepository) GetShipmentByNumber(ctx context.Context, shipmentNumber string) (*models.Shipment, error) {
	var shipment models.Shipment
	err := r.db.DB.WithContext(ctx).Where("shipment_number = ?", shipmentNumber).First(&shipment).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get shipment with ShipmentNumber: %s", shipmentNumber)
	}
	return &shipment, nil
}

func (r *shipmentRepository) GetShipmentByID(ctx context.Context, id uuid.UUID) (*models.Shipment, error) {
	var shipment models.Shipment
	err := r.db.DB.WithContext(ctx).First(&shipment, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get shipment with ID: %s", id)
	}
	return &shipment, nil
}

func (r *shipmentRepository) CheckUserAlreadyTracking(ctx context.Context, userID uuid.UUID, shipmentNumber string) (bool, error) {
	var exists bool

	err := r.db.DB.WithContext(ctx).
		Model(&models.UserShipment{}).
		Select("count(*) > 0").
		Joins("JOIN shipments s on s.id = user_shipments.shipment_id").
		Where("user_shipments.user_id = ? AND s.shipment_number = ?", userID, shipmentNumber).
		Scan(&exists).Error
	if err != nil {
		return false, fmt.Errorf("failed to check user tracking: %w", err)
	}

	return exists, nil
}

func (r *shipmentRepository) CheckUserOwnsShipment(ctx context.Context, userID, shipmentID uuid.UUID) (bool, error) {
	var exists bool

	err := r.db.DB.WithContext(ctx).
		Model(&models.UserShipment{}).
		Select("count(*) > 0").
		Where("user_id = ? AND shipment_id = ?", userID, shipmentID).
		Scan(&exists).Error
	if err != nil {
		return false, fmt.Errorf("failed to check shipment ownership: %w", err)
	}

	return exists, nil
}

func (r *shipmentRepository) CheckShipmentExists(ctx context.Context, shipmentNumber string) (bool, error) {
	var exists bool

	err := r.db.DB.WithContext(ctx).
		Model(&models.Shipment{}).
		Select("count(*) > 0").
		Where("shipment_number = ?", shipmentNumber).
		Scan(&exists).Error
	if err != nil {
		return false, fmt.Errorf("failed to check shipment existence: %w", err)
	}

	return exists, nil
}

func (r *shipmentRepository) AddExistingShipmentToUser(ctx context.Context, userID uuid.UUID, shipmentNumber string) (*models.Shipment, error) {
	shipment, err := r.GetShipmentByNumber(ctx, shipmentNumber)
	if err != nil {
		return nil, err
	}

	link := &models.UserShipment{
		UserID:     userID,
		ShipmentID: shipment.ID,
	}
	if err := r.db.DB.WithContext(ctx).Create(&link).Error; err != nil {
		return nil, fmt.Errorf("failed to link shipment to user: %w", err)
	}

	return shipment, nil
}

func (r *shipmentRepository) UpdateShipment(ctx context.Context, id uuid.UUID, updates *models.Shipment) (*models.Shipment, error) {
	err := r.db.DB.WithContext(ctx).
		Model(&models.Shipment{}).
		Where("id = ?", id).
		Updates(&updates).Error
	if err != nil {
		return nil, fmt.Errorf("failed to update shipment: %s", id)
	}

	var updated models.Shipment
	err = r.db.DB.WithContext(ctx).First(&updated, "id = ?", id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated shipment: %s", id)
	}

	return &updated, nil
}

func (r *shipmentRepository) CreateLocation(ctx context.Context, shipmentID *uuid.UUID, location *models.Location) (*models.Location, error) {
	err := r.db.DB.WithContext(ctx).Where(location).FirstOrCreate(location).Error
	if err != nil {
		return nil, err
	}

	if shipmentID != nil {
		link := models.ShipmentLocation{
			ShipmentID: *shipmentID,
			LocationID: location.ID,
		}
		if err := r.db.DB.WithContext(ctx).Create(&link).Error; err != nil {
			return nil, fmt.Errorf("failed to link location to shipment: %w", err)
		}
	}

	return location, nil
}

func (r *shipmentRepository) FindLocationByLocode(ctx context.Context, locode string) (*models.Location, error) {
	var location models.Location
	err := r.db.DB.WithContext(ctx).Where("locode = ?", locode).First(&location).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("location not found")
		}
		return nil, fmt.Errorf("failed to get location with locode: %s", locode)
	}
	return &location, nil
}

func (r *shipmentRepository) FindLocationByID(ctx context.Context, id uuid.UUID) (*models.Location, error) {
	var location models.Location
	err := r.db.DB.WithContext(ctx).First(&location, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get location with ID: %s", id)
	}
	return &location, nil
}

func (r *shipmentRepository) CreateRoute(ctx context.Context, route *models.ShipmentRoute) (*models.ShipmentRoute, error) {
	err := r.db.DB.WithContext(ctx).Where(route).FirstOrCreate(route).Error
	if err != nil {
		return nil, err
	}
	return route, nil
}

func (r *shipmentRepository) CreateVessel(ctx context.Context, shipmentID *uuid.UUID, vessel *models.Vessel) (*models.Vessel, error) {
	err := r.db.DB.WithContext(ctx).Where(vessel).FirstOrCreate(vessel).Error
	if err != nil {
		return nil, err
	}

	if shipmentID != nil {
		link := models.ShipmentVessel{
			ShipmentID: *shipmentID,
			VesselID:   vessel.ID,
		}
		if err := r.db.DB.WithContext(ctx).Create(&link).Error; err != nil {
			return nil, fmt.Errorf("failed to link vessel to shipment: %w", err)
		}
	}

	return vessel, nil
}

func (r *shipmentRepository) FindVesselByIMOAndMMSI(ctx context.Context, imo, mmsi int) (*models.Vessel, error) {
	var vessel models.Vessel
	err := r.db.DB.WithContext(ctx).Where("imo = ? AND mmsi = ?", imo, mmsi).First(&vessel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get vessel with imo and mmsi: %d, %d", imo, mmsi)
	}
	return &vessel, nil
}

func (r *shipmentRepository) FindVesselByID(ctx context.Context, id *uuid.UUID) (*models.Vessel, error) {

	var vessel models.Vessel
	err := r.db.DB.WithContext(ctx).First(&vessel, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get vessel with ID: %s", id)
	}
	return &vessel, nil
}

func (r *shipmentRepository) CreateFacility(ctx context.Context, shipmentID *uuid.UUID, facility *models.Facility) (*models.Facility, error) {
	err := r.db.DB.WithContext(ctx).Where(facility).FirstOrCreate(facility).Error
	if err != nil {
		return nil, err
	}

	if shipmentID != nil {
		link := models.ShipmentFacility{
			ShipmentID: *shipmentID,
			FacilityID: facility.ID,
		}
		if err := r.db.DB.WithContext(ctx).Create(&link).Error; err != nil {
			return nil, fmt.Errorf("failed to link facility to shipment: %w", err)
		}
	}
	return facility, err
}

func (r *shipmentRepository) FindFacilityByLocode(ctx context.Context, locode string) (*models.Facility, error) {
	var facility models.Facility
	err := r.db.DB.WithContext(ctx).Where("locode = ?", locode).First(&facility).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("location not found")
		}
		return nil, fmt.Errorf("failed to get location with locode: %s", locode)
	}
	return &facility, nil
}

func (r *shipmentRepository) FindFacilityByID(ctx context.Context, id *uuid.UUID) (*models.Facility, error) {
	var facility models.Facility
	err := r.db.DB.WithContext(ctx).First(&facility, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get facility with ID: %s", id)
	}
	return &facility, nil
}

func (r *shipmentRepository) CreateContainer(ctx context.Context, shipmentID *uuid.UUID, container *models.Container) (*models.Container, error) {
	err := r.db.DB.WithContext(ctx).Where(container).FirstOrCreate(container).Error
	if err != nil {
		return nil, err
	}

	if shipmentID != nil {
		link := models.ShipmentContainer{
			ShipmentID:  *shipmentID,
			ContainerID: container.ID,
		}
		if err := r.db.DB.WithContext(ctx).Create(&link).Error; err != nil {
			return nil, fmt.Errorf("failed to link container to shipment: %w", err)
		}
	}
	return container, nil
}

func (r *shipmentRepository) CreateContainerEvent(ctx context.Context, containerEvent *models.ContainerEvent) (*models.ContainerEvent, error) {
	err := r.db.DB.WithContext(ctx).Where(containerEvent).FirstOrCreate(containerEvent).Error
	if err != nil {
		return nil, err
	}
	return containerEvent, nil
}

func (r *shipmentRepository) CreateRouteSegment(ctx context.Context, routeSegment *models.RouteSegment) (*models.RouteSegment, error) {
	err := r.db.DB.WithContext(ctx).Where(routeSegment).FirstOrCreate(routeSegment).Error
	if err != nil {
		return nil, err
	}
	return routeSegment, nil
}

func (r *shipmentRepository) CreateRouteSegmentPoint(ctx context.Context, point *models.RouteSegmentPoint) (*models.RouteSegmentPoint, error) {
	err := r.db.DB.WithContext(ctx).Where(point).FirstOrCreate(point).Error
	if err != nil {
		return nil, err
	}
	return point, nil
}

func (r *shipmentRepository) CreateCoordinate(ctx context.Context, coordinate *models.Coordinate) (*models.Coordinate, error) {
	err := r.db.DB.WithContext(ctx).Where(coordinate).FirstOrCreate(coordinate).Error
	if err != nil {
		return nil, err
	}
	return coordinate, nil
}

func (r *shipmentRepository) CreateAis(ctx context.Context, ais *models.Ais) (*models.Ais, error) {
	err := r.db.DB.WithContext(ctx).Where(ais).FirstOrCreate(ais).Error
	if err != nil {
		return nil, err
	}
	return ais, nil
}

func (r *shipmentRepository) GetShipmentAisData(ctx context.Context, shipmentID uuid.UUID) (*dto.ShipmentAisResponse, error) {
	var aisModel models.Ais

	err := r.db.DB.WithContext(ctx).
		Where("shipment_id = ?", shipmentID).
		Order("updated_at DESC"). // latest snapshot
		First(&aisModel).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// Convert to DTO
	aisResponse := dto.ShipmentAisResponse{
		Status:                   aisModel.Status,
		LastEventDescription:     aisModel.LastEventDescription,
		LastEventDate:            aisModel.LastEventDate,
		LastEventVoyage:          aisModel.LastEventVoyage,
		DischargePortName:        aisModel.DischargePortName,
		DischargePortCountryCode: aisModel.DischargePortCountryCode,
		DischargePortCode:        aisModel.DischargePortCode,
		DischargePortDate:        aisModel.DischargePortDate,
		DischargePortDateLabel:   aisModel.DischargePortDateLabel,
		DeparturePortName:        aisModel.DeparturePortName,
		DeparturePortCountryCode: aisModel.DeparturePortCountryCode,
		DeparturePortCode:        aisModel.DeparturePortCode,
		DeparturePortDate:        aisModel.DeparturePortDate,
		DeparturePortDateLabel:   aisModel.DeparturePortDateLabel,
		ArrivalPortName:          aisModel.ArrivalPortName,
		ArrivalPortCountryCode:   aisModel.ArrivalPortCountryCode,
		ArrivalPortCode:          aisModel.ArrivalPortCode,
		ArrivalPortDate:          aisModel.ArrivalPortDate,
		ArrivalPortDateLabel:     aisModel.ArrivalPortDateLabel,
		LastVesselPositionLat:    aisModel.LastVesselPositionLat,
		LastVesselPositionLng:    aisModel.LastVesselPositionLng,
		LastVesselPositionUpdate: aisModel.LastVesselPositionUpdate,
		UpdatedAt:                aisModel.UpdatedAt,
	}

	// Fetch vessel data if VesselID is present
	if aisModel.VesselID != nil {
		vessel, err := r.FindVesselByID(ctx, aisModel.VesselID)
		if err != nil {
			return nil, err
		}
		vesselResponse := r.convertVesselToDTO(*vessel)
		aisResponse.Vessel = &vesselResponse
	}

	return &aisResponse, nil
}

func (r *shipmentRepository) GetShipmentDetails(ctx context.Context, shipmentID uuid.UUID) (*dto.ShipmentDetailsResponse, error) {
	shipment, err := r.GetShipmentByID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}

	locations, err := r.getShipmentLocations(ctx, shipmentID)
	if err != nil {
		return nil, err
	}

	route, err := r.getShipmentRoute(ctx, shipmentID)
	if err != nil {
		return nil, err
	}

	vessels, err := r.getShipmentVessels(ctx, shipmentID)
	if err != nil {
		return nil, err
	}

	facilities, err := r.getShipmentFacilities(ctx, shipmentID)
	if err != nil {
		return nil, err
	}

	containers, err := r.getShipmentContainers(ctx, shipmentID)
	if err != nil {
		return nil, err
	}

	routeData, err := r.getShipmentRouteData(ctx, shipmentID, shipment.ID)
	if err != nil {
		return nil, err
	}

	return &dto.ShipmentDetailsResponse{
		ID:             shipment.ID,
		ShipmentType:   shipment.ShipmentType,
		ShipmentNumber: shipment.ShipmentNumber,
		SealineCode:    shipment.SealineCode,
		SealineName:    shipment.SealineName,
		ShippingStatus: shipment.ShippingStatus,
		CreatedAt:      shipment.CreatedAt,
		UpdatedAt:      shipment.UpdatedAt,
		Locations:      locations,
		Route:          route,
		Vessels:        vessels,
		Facilities:     facilities,
		Containers:     containers,
		RouteData:      routeData,
	}, nil
}

// getShipmentLocations fetches and converts shipment locations
func (r *shipmentRepository) getShipmentLocations(ctx context.Context, shipmentID uuid.UUID) ([]dto.ShipmentLocationResponse, error) {
	var locations []models.Location
	err := r.db.DB.WithContext(ctx).
		Joins("JOIN shipment_locations sl ON sl.location_id = locations.id").
		Where("sl.shipment_id = ?", shipmentID).
		Order("sl.added_at ASC").
		Find(&locations).Error
	if err != nil {
		return nil, err
	}

	return r.convertLocationsToDTO(locations), nil
}

// getShipmentRoute fetches and converts shipment route
func (r *shipmentRepository) getShipmentRoute(ctx context.Context, shipmentID uuid.UUID) (dto.ShipmentRouteResponse, error) {
	var routes []models.ShipmentRoute
	err := r.db.DB.WithContext(ctx).
		Preload("Location").
		Where("shipment_id = ?", shipmentID).
		Find(&routes).Error
	if err != nil {
		return dto.ShipmentRouteResponse{}, err
	}

	return r.convertRouteToDTO(routes), nil
}

// getShipmentVessels fetches and converts shipment vessels
func (r *shipmentRepository) getShipmentVessels(ctx context.Context, shipmentID uuid.UUID) ([]dto.ShipmentVesselResponse, error) {
	var vessels []models.Vessel
	err := r.db.DB.WithContext(ctx).
		Joins("JOIN shipment_vessels sv ON sv.vessel_id = vessels.id").
		Where("sv.shipment_id = ?", shipmentID).
		Order("sv.added_at ASC").
		Find(&vessels).Error
	if err != nil {
		return nil, err
	}

	return r.convertVesselsToDTO(vessels), nil
}

// getShipmentFacilities fetches and converts shipment facilities
func (r *shipmentRepository) getShipmentFacilities(ctx context.Context, shipmentID uuid.UUID) ([]dto.ShipmentFacilityResponse, error) {
	var facilities []models.Facility
	err := r.db.DB.WithContext(ctx).
		Joins("JOIN shipment_facilities sf ON sf.facility_id = facilities.id").
		Where("sf.shipment_id = ?", shipmentID).
		Order("sf.added_at ASC").
		Find(&facilities).Error
	if err != nil {
		return nil, err
	}

	return r.convertFacilitiesToDTO(facilities), nil
}

// getShipmentContainers fetches and converts shipment containers with events
func (r *shipmentRepository) getShipmentContainers(ctx context.Context, shipmentID uuid.UUID) ([]dto.ShipmentContainerResponse, error) {
	var containers []models.Container
	err := r.db.DB.WithContext(ctx).
		Joins("JOIN shipment_containers sc ON sc.container_id = containers.id").
		Where("sc.shipment_id = ?", shipmentID).
		Order("sc.added_at ASC").
		Find(&containers).Error
	if err != nil {
		return nil, err
	}

	containersResponse := make([]dto.ShipmentContainerResponse, len(containers))
	for i, container := range containers {
		containerEvents, err := r.getContainerEvents(ctx, container.ID)
		if err != nil {
			return nil, err
		}

		containersResponse[i] = dto.ShipmentContainerResponse{
			Number:   container.Number,
			IsoCode:  container.IsoCode,
			SizeType: container.SizeType,
			Status:   container.Status,
			Events:   containerEvents,
		}
	}

	return containersResponse, nil
}

// getContainerEvents fetches and converts container events
func (r *shipmentRepository) getContainerEvents(ctx context.Context, containerID uuid.UUID) ([]dto.ShipmentContainerEventResponse, error) {
	var containerEvents []models.ContainerEvent
	err := r.db.DB.WithContext(ctx).
		Where("container_id = ?", containerID).
		Find(&containerEvents).Error
	if err != nil {
		return nil, err
	}

	eventResponses := make([]dto.ShipmentContainerEventResponse, len(containerEvents))
	for i, event := range containerEvents {
		eventResponse, err := r.convertContainerEventToDTO(ctx, event)
		if err != nil {
			return nil, err
		}
		eventResponses[i] = eventResponse
	}

	return eventResponses, nil
}

// convertContainerEventToDTO converts a container event to DTO with related data
func (r *shipmentRepository) convertContainerEventToDTO(ctx context.Context, event models.ContainerEvent) (dto.ShipmentContainerEventResponse, error) {
	location, err := r.FindLocationByID(ctx, event.LocationID)
	if err != nil {
		return dto.ShipmentContainerEventResponse{}, err
	}

	eventResponse := dto.ShipmentContainerEventResponse{
		Location:          r.convertLocationToDTO(*location),
		Description:       event.Description,
		EventType:         event.EventType,
		EventCode:         event.EventCode,
		Status:            event.Status,
		Date:              event.Date,
		IsActual:          event.IsActual,
		IsAdditionalEvent: event.IsAdditionalEvent,
		RouteType:         event.RouteType,
		TransportType:     event.TransportType,
		Voyage:            event.Voyage,
	}

	if event.FacilityID != nil {
		facility, err := r.FindFacilityByID(ctx, event.FacilityID)
		if err != nil {
			return dto.ShipmentContainerEventResponse{}, err
		}
		facilityResponse := r.convertFacilityToDTO(*facility)
		eventResponse.Facility = &facilityResponse
	}

	if event.VesselID != nil {
		vessel, err := r.FindVesselByID(ctx, event.VesselID)
		if err != nil {
			return dto.ShipmentContainerEventResponse{}, err
		}
		vesselResponse := r.convertVesselToDTO(*vessel)
		eventResponse.Vessel = &vesselResponse
	}

	return eventResponse, nil
}

// getShipmentRouteData fetches and assembles route data including segments, coordinates, and AIS
func (r *shipmentRepository) getShipmentRouteData(ctx context.Context, shipmentID, shipmentDbID uuid.UUID) (dto.ShipmentRouteDataResponse, error) {
	routeSegments, err := r.getShipmentRouteSegments(ctx, shipmentID)
	if err != nil {
		return dto.ShipmentRouteDataResponse{}, err
	}

	coordinates, err := r.getShipmentCoordinates(ctx, shipmentDbID)
	if err != nil {
		return dto.ShipmentRouteDataResponse{}, err
	}

	aisResponse, err := r.GetShipmentAisData(ctx, shipmentID)
	if err != nil {
		return dto.ShipmentRouteDataResponse{}, err
	}
	if aisResponse == nil {
		return dto.ShipmentRouteDataResponse{}, fmt.Errorf("ais data not found for shipment: %s", shipmentID)
	}

	return dto.ShipmentRouteDataResponse{
		RouteSegments: routeSegments,
		Coordinates:   coordinates,
		Ais:           *aisResponse,
	}, nil
}

// getShipmentRouteSegments fetches and converts route segments
func (r *shipmentRepository) getShipmentRouteSegments(ctx context.Context, shipmentID uuid.UUID) ([]dto.ShipmentRouteSegmentResponse, error) {
	var routeSegments []models.RouteSegment
	err := r.db.DB.WithContext(ctx).
		Where("shipment_id = ?", shipmentID).
		Order("segment_order ASC").
		Find(&routeSegments).Error
	if err != nil {
		return nil, err
	}

	routeSegmentsResponse := make([]dto.ShipmentRouteSegmentResponse, 0, len(routeSegments))
	for _, segment := range routeSegments {
		segmentResponse, err := r.convertRouteSegmentToDTO(ctx, segment)
		if err != nil {
			return nil, err
		}
		routeSegmentsResponse = append(routeSegmentsResponse, segmentResponse)
	}

	return routeSegmentsResponse, nil
}

// convertRouteSegmentToDTO converts a route segment with its points to DTO
func (r *shipmentRepository) convertRouteSegmentToDTO(ctx context.Context, segment models.RouteSegment) (dto.ShipmentRouteSegmentResponse, error) {
	var routeSegmentPoints []models.RouteSegmentPoint
	err := r.db.DB.WithContext(ctx).
		Where("segment_id = ?", segment.ID).
		Order("point_order ASC").
		Find(&routeSegmentPoints).Error
	if err != nil {
		return dto.ShipmentRouteSegmentResponse{}, err
	}

	path := make([]dto.ShipmentRouteSegmentPointResponse, 0, len(routeSegmentPoints))
	for _, point := range routeSegmentPoints {
		path = append(path, dto.ShipmentRouteSegmentPointResponse{
			Latitude:   point.Latitude,
			Longitude:  point.Longitude,
			PointOrder: point.PointOrder,
		})
	}

	return dto.ShipmentRouteSegmentResponse{
		RouteType:    segment.RouteType,
		SegmentOrder: segment.SegmentOrder,
		Path:         path,
	}, nil
}

// getShipmentCoordinates fetches the latest coordinates for a shipment
func (r *shipmentRepository) getShipmentCoordinates(ctx context.Context, shipmentID uuid.UUID) (dto.ShipmentCoordinatesResponse, error) {
	var shipmentCoordinates *models.Coordinate
	err := r.db.DB.WithContext(ctx).
		Where("shipment_id = ?", shipmentID).
		Order("updated_at DESC").
		First(&shipmentCoordinates).Error
	if err != nil {
		return dto.ShipmentCoordinatesResponse{}, err
	}

	return dto.ShipmentCoordinatesResponse{
		Latitude:  shipmentCoordinates.Latitude,
		Longitude: shipmentCoordinates.Longitude,
		UpdatedAt: shipmentCoordinates.UpdatedAt,
	}, nil
}

// Helper methods for converting models to DTOs

func (r *shipmentRepository) convertLocationsToDTO(locations []models.Location) []dto.ShipmentLocationResponse {
	locationResponses := make([]dto.ShipmentLocationResponse, len(locations))
	for i, loc := range locations {
		locationResponses[i] = r.convertLocationToDTO(loc)
	}
	return locationResponses
}

func (r *shipmentRepository) convertLocationToDTO(location models.Location) dto.ShipmentLocationResponse {
	return dto.ShipmentLocationResponse{
		Name:        location.Name,
		State:       location.State,
		Country:     location.Country,
		CountryCode: location.CountryCode,
		Locode:      location.Locode,
		Latitude:    location.Latitude,
		Longitude:   location.Longitude,
		Timezone:    location.Timezone,
	}
}

func (r *shipmentRepository) convertRouteToDTO(routes []models.ShipmentRoute) dto.ShipmentRouteResponse {
	routeResponse := dto.ShipmentRouteResponse{}
	for _, route := range routes {
		item := &dto.ShipmentRoutePoint{
			Location:      r.convertLocationToDTO(route.Location),
			Date:          route.Date,
			Actual:        route.Actual,
			PredictiveETA: route.PredictiveETA,
		}
		switch route.RouteType {
		case "PREPOL":
			routeResponse.Prepol = item
		case "POL":
			routeResponse.Pol = item
		case "POD":
			routeResponse.Pod = item
		case "POSTPOD":
			routeResponse.Postpod = item
		}
	}
	return routeResponse
}

func (r *shipmentRepository) convertVesselsToDTO(vessels []models.Vessel) []dto.ShipmentVesselResponse {
	vesselResponses := make([]dto.ShipmentVesselResponse, len(vessels))
	for i, vessel := range vessels {
		vesselResponses[i] = r.convertVesselToDTO(vessel)
	}
	return vesselResponses
}

func (r *shipmentRepository) convertVesselToDTO(vessel models.Vessel) dto.ShipmentVesselResponse {
	return dto.ShipmentVesselResponse{
		Name:     vessel.Name,
		Imo:      vessel.Imo,
		Mmsi:     vessel.Mmsi,
		CallSign: vessel.CallSign,
		Flag:     vessel.Flag,
	}
}

func (r *shipmentRepository) convertFacilitiesToDTO(facilities []models.Facility) []dto.ShipmentFacilityResponse {
	facilityResponses := make([]dto.ShipmentFacilityResponse, len(facilities))
	for i, facility := range facilities {
		facilityResponses[i] = r.convertFacilityToDTO(facility)
	}
	return facilityResponses
}

func (r *shipmentRepository) convertFacilityToDTO(facility models.Facility) dto.ShipmentFacilityResponse {
	return dto.ShipmentFacilityResponse{
		Name:        facility.Name,
		CountryCode: facility.CountryCode,
		Locode:      facility.Locode,
		BicCode:     facility.BicCode,
		SmdgCode:    facility.SmdgCode,
		Latitude:    facility.Latitude,
		Longitude:   facility.Longitude,
	}
}
