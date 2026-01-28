package repositories

import (
	"context"
	"errors"
	"fmt"
	"go-starter/internal/modules/shipments/dto"
	"go-starter/internal/modules/shipments/models"
	"go-starter/pkg/db"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ShipmentRepository interface {
	GetDB() *db.Database

	CreateShipment(ctx context.Context, userID uuid.UUID, shipment *models.Shipment) (*models.Shipment, error)
	GetShipmentByNumber(ctx context.Context, shipmentNumber string) (*models.Shipment, error)
	GetShipmentByID(ctx context.Context, userID, shipmentID uuid.UUID) (*models.Shipment, error)
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

	GetShipmentWithRelations(ctx context.Context, shipmentID uuid.UUID) (*models.Shipment, error)
	GetShipmentDetails(ctx context.Context, userID, shipmentID uuid.UUID) (*dto.ShipmentDetailsResponse, error)
	UpdateShipmentInfo(ctx context.Context, userID, shipmentID uuid.UUID, req *dto.UpdateShipmentInfoRequest) error
	UpdateShipmentInfoPartial(ctx context.Context, userID, shipmentID uuid.UUID, updates map[string]interface{}) error

	DeleteShipmentLocations(ctx context.Context, shipmentID uuid.UUID) error
	DeleteShipmentRoutes(ctx context.Context, shipmentID uuid.UUID) error
	DeleteShipmentVessels(ctx context.Context, shipmentID uuid.UUID) error
	DeleteShipmentFacilities(ctx context.Context, shipmentID uuid.UUID) error
	DeleteShipmentContainers(ctx context.Context, shipmentID uuid.UUID) error
	DeleteRouteSegments(ctx context.Context, shipmentID uuid.UUID) error
	DeleteShipmentCoordinates(ctx context.Context, shipmentID uuid.UUID) error
	DeleteShipmentAis(ctx context.Context, shipmentID uuid.UUID) error
	DeleteAllShipmentRelatedData(ctx context.Context, shipmentID uuid.UUID) error
	GetShipmentDataSummary(ctx context.Context, shipmentID uuid.UUID) (*ShipmentDataSummary, error)

	GetShipmentsForGrid(ctx context.Context, userID uuid.UUID) ([]models.Shipment, error)
	DeleteUserShipment(ctx context.Context, userID, shipmentID uuid.UUID) error
	BulkDeleteUserShipments(ctx context.Context, userID uuid.UUID, shipmentIDs []uuid.UUID) error
	GetAllShipmentsForRefresh(ctx context.Context, skipRecentlyUpdated time.Duration) ([]ShipmentForRefresh, error)
	GetShipmentsWithDetailsForGrid(ctx context.Context, userID uuid.UUID) ([]models.Shipment, error)

	ConvertContainerEventToDTO(ctx context.Context, event models.ContainerEvent) (dto.ShipmentContainerEventResponse, error)
	ConvertRouteToDTO(routes []models.ShipmentRoute) dto.ShipmentRouteResponse
	ConvertRouteSegmentToDTO(ctx context.Context, segment models.RouteSegment) (dto.ShipmentRouteSegmentResponse, error)
	ConvertLocationsToDTO(locations []models.Location) []dto.ShipmentLocationResponse
	ConvertLocationToDTO(location models.Location) dto.ShipmentLocationResponse
	ConvertVesselsToDTO(vessels []models.Vessel) []dto.ShipmentVesselResponse
	ConvertVesselToDTO(vessel models.Vessel) dto.ShipmentVesselResponse
	ConvertFacilitiesToDTO(facilities []models.Facility) []dto.ShipmentFacilityResponse
	ConvertFacilityToDTO(facility models.Facility) dto.ShipmentFacilityResponse

	GetShipmentDate(containers []dto.ShipmentContainerResponse) (*time.Time, error)
	MapShipmentToDTO(ctx context.Context, shipment models.Shipment) dto.ShipmentDetailsResponse
}

type shipmentRepository struct {
	db *db.Database
}

func NewShipmentRepository(db *db.Database) ShipmentRepository {
	if err := models.AutoMigrateShipments(db.DB); err != nil {
		log.Printf("failed to migrate shipments: %s", err)
	}

	if err := models.FixShipmentRouteConstraint(db.DB); err != nil {
		log.Printf("failed to fix shipment route constraint: %s", err)
	}

	return &shipmentRepository{
		db: db,
	}
}

func (r *shipmentRepository) CreateShipment(
	ctx context.Context,
	userID uuid.UUID,
	shipment *models.Shipment,
) (*models.Shipment, error) {
	db := r.getDBFromContext(ctx)
	if err := db.WithContext(ctx).Create(&shipment).Error; err != nil {
		return nil, fmt.Errorf("failed to create shipment: %w", err)
	}

	link := models.UserShipment{
		UserID:     userID,
		ShipmentID: shipment.ID,
	}

	if err := db.WithContext(ctx).Create(&link).Error; err != nil {
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

func (r *shipmentRepository) GetShipmentByID(ctx context.Context, userID, shipmentID uuid.UUID) (*models.Shipment, error) {
	var shipment models.Shipment
	err := r.db.DB.WithContext(ctx).
		Joins("JOIN user_shipments us ON us.shipment_id = shipments.id").
		Where("us.user_id = ? AND shipments.id = ?", userID, shipmentID).
		First(&shipment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("shipment not found or access denied")
		}
		return nil, fmt.Errorf("failed to get shipment: %w", err)
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
	db := r.getDBFromContext(ctx)

	var existingLocation models.Location
	err := db.WithContext(ctx).Where("locode = ?", location.Locode).First(&existingLocation).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = db.WithContext(ctx).Create(location).Error
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		existingLocation.Name = location.Name
		existingLocation.State = location.State
		existingLocation.Country = location.Country
		existingLocation.CountryCode = location.CountryCode
		existingLocation.Latitude = location.Latitude
		existingLocation.Longitude = location.Longitude
		existingLocation.Timezone = location.Timezone

		err = db.WithContext(ctx).Save(&existingLocation).Error
		if err != nil {
			return nil, err
		}
		*location = existingLocation
	}

	if shipmentID != nil {
		link := models.ShipmentLocation{
			ShipmentID: *shipmentID,
			LocationID: location.ID,
		}

		err := db.WithContext(ctx).Where(&link).FirstOrCreate(&link).Error
		if err != nil {
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
	db := r.getDBFromContext(ctx)

	var existingRoute models.ShipmentRoute
	err := db.WithContext(ctx).Where("shipment_id = ? AND location_id = ? AND route_type = ?",
		route.ShipmentID, route.LocationID, route.RouteType).First(&existingRoute).Error

	if err == nil {
		existingRoute.Date = route.Date
		existingRoute.Actual = route.Actual
		existingRoute.PredictiveETA = route.PredictiveETA
		err = db.WithContext(ctx).Save(&existingRoute).Error
		if err != nil {
			return nil, fmt.Errorf("failed to update existing route: %w", err)
		}
		return &existingRoute, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check for existing route: %w", err)
	}

	err = db.WithContext(ctx).Create(route).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create route: %w", err)
	}
	return route, nil
}

func (r *shipmentRepository) CreateVessel(ctx context.Context, shipmentID *uuid.UUID, vessel *models.Vessel) (*models.Vessel, error) {
	db := r.getDBFromContext(ctx)

	var existingVessel models.Vessel
	err := db.WithContext(ctx).Where("imo = ?", vessel.Imo).First(&existingVessel).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = db.WithContext(ctx).Create(vessel).Error
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		existingVessel.Name = vessel.Name
		existingVessel.Mmsi = vessel.Mmsi
		existingVessel.CallSign = vessel.CallSign
		existingVessel.Flag = vessel.Flag

		err = db.WithContext(ctx).Save(&existingVessel).Error
		if err != nil {
			return nil, err
		}
		*vessel = existingVessel
	}

	if shipmentID != nil {
		link := models.ShipmentVessel{
			ShipmentID: *shipmentID,
			VesselID:   vessel.ID,
		}

		err := db.WithContext(ctx).Where(&link).FirstOrCreate(&link).Error
		if err != nil {
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
	db := r.getDBFromContext(ctx)

	var existingFacility models.Facility
	err := db.WithContext(ctx).Where("name = ?", facility.Name).First(&existingFacility).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = db.WithContext(ctx).Create(facility).Error
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		existingFacility.CountryCode = facility.CountryCode
		existingFacility.Locode = facility.Locode
		existingFacility.BicCode = facility.BicCode
		existingFacility.SmdgCode = facility.SmdgCode
		existingFacility.Latitude = facility.Latitude
		existingFacility.Longitude = facility.Longitude

		err = db.WithContext(ctx).Save(&existingFacility).Error
		if err != nil {
			return nil, err
		}
		*facility = existingFacility
	}

	if shipmentID != nil {
		link := models.ShipmentFacility{
			ShipmentID: *shipmentID,
			FacilityID: facility.ID,
		}

		err := db.WithContext(ctx).Where(&link).FirstOrCreate(&link).Error
		if err != nil {
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
	db := r.getDBFromContext(ctx)

	var existingContainer models.Container
	err := db.WithContext(ctx).Where("number = ?", container.Number).First(&existingContainer).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = db.WithContext(ctx).Create(container).Error
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		existingContainer.IsoCode = container.IsoCode
		existingContainer.SizeType = container.SizeType
		existingContainer.Status = container.Status

		err = db.WithContext(ctx).Save(&existingContainer).Error
		if err != nil {
			return nil, err
		}
		*container = existingContainer
	}

	if shipmentID != nil {
		link := models.ShipmentContainer{
			ShipmentID:  *shipmentID,
			ContainerID: container.ID,
		}

		err := db.WithContext(ctx).Where(&link).FirstOrCreate(&link).Error
		if err != nil {
			return nil, fmt.Errorf("failed to link container to shipment: %w", err)
		}
	}
	return container, nil
}

func (r *shipmentRepository) CreateContainerEvent(ctx context.Context, containerEvent *models.ContainerEvent) (*models.ContainerEvent, error) {
	db := r.getDBFromContext(ctx)
	err := db.WithContext(ctx).Where(containerEvent).FirstOrCreate(containerEvent).Error
	if err != nil {
		return nil, err
	}
	return containerEvent, nil
}

func (r *shipmentRepository) CreateRouteSegment(ctx context.Context, routeSegment *models.RouteSegment) (*models.RouteSegment, error) {
	db := r.getDBFromContext(ctx)
	err := db.WithContext(ctx).Where(routeSegment).FirstOrCreate(routeSegment).Error
	if err != nil {
		return nil, err
	}
	return routeSegment, nil
}

func (r *shipmentRepository) CreateRouteSegmentPoint(ctx context.Context, point *models.RouteSegmentPoint) (*models.RouteSegmentPoint, error) {
	db := r.getDBFromContext(ctx)
	err := db.WithContext(ctx).Where(point).FirstOrCreate(point).Error
	if err != nil {
		return nil, err
	}
	return point, nil
}

func (r *shipmentRepository) CreateCoordinate(ctx context.Context, coordinate *models.Coordinate) (*models.Coordinate, error) {
	db := r.getDBFromContext(ctx)
	err := db.WithContext(ctx).Where(coordinate).FirstOrCreate(coordinate).Error
	if err != nil {
		return nil, err
	}
	return coordinate, nil
}

func (r *shipmentRepository) CreateAis(ctx context.Context, ais *models.Ais) (*models.Ais, error) {
	db := r.getDBFromContext(ctx)
	err := db.WithContext(ctx).Where(ais).FirstOrCreate(ais).Error
	if err != nil {
		return nil, err
	}
	return ais, nil
}

func (r *shipmentRepository) GetShipmentAisData(ctx context.Context, shipmentID uuid.UUID) (*dto.ShipmentAisResponse, error) {
	var aisModel models.Ais

	err := r.db.DB.WithContext(ctx).
		Where("shipment_id = ?", shipmentID).
		Order("updated_at DESC").
		First(&aisModel).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

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

	if aisModel.VesselID != nil {
		vessel, err := r.FindVesselByID(ctx, aisModel.VesselID)
		if err != nil {
			return nil, err
		}
		vesselResponse := r.ConvertVesselToDTO(*vessel)
		aisResponse.Vessel = &vesselResponse
	}

	return &aisResponse, nil
}

func (r *shipmentRepository) GetShipmentWithRelations(ctx context.Context, shipmentID uuid.UUID) (*models.Shipment, error) {
	db := r.getDBFromContext(ctx)
	var shipment models.Shipment

	err := db.WithContext(ctx).
		Preload("Locations").
		Preload("Routes.Location").
		Preload("Vessels").
		Preload("Facilities").
		Preload("Containers").
		Preload("Containers.Events.Location").
		Preload("Containers.Events.Facility").
		Preload("Containers.Events.Vessel").
		Preload("RouteSegments").
		Preload("RouteSegments.Points").
		Preload("Coordinates").
		Preload("Ais").
		First(&shipment, "id = ?", shipmentID).Error

	if err != nil {
		return nil, err
	}

	return &shipment, err
}

func (r *shipmentRepository) GetShipmentDetails(ctx context.Context, userID, shipmentID uuid.UUID) (*dto.ShipmentDetailsResponse, error) {
	shipment, err := r.GetShipmentByID(ctx, userID, shipmentID)
	if err != nil {
		return nil, err
	}

	var userShipment models.UserShipment
	if err := r.db.DB.WithContext(ctx).
		Where("user_id = ? AND shipment_id = ?", userID, shipmentID).
		First(&userShipment).Error; err != nil {
		return nil, fmt.Errorf("failed to get user shipment info: %w", err)
	}

	return r.buildShipmentResponse(ctx, shipment)
}

func (r *shipmentRepository) buildShipmentResponse(ctx context.Context, shipment *models.Shipment) (*dto.ShipmentDetailsResponse, error) {
	locations, err := r.getShipmentLocations(ctx, shipment.ID)
	if err != nil {
		return nil, err
	}

	route, err := r.getShipmentRoute(ctx, shipment.ID)
	if err != nil {
		return nil, err
	}

	vessels, err := r.getShipmentVessels(ctx, shipment.ID)
	if err != nil {
		return nil, err
	}

	facilities, err := r.getShipmentFacilities(ctx, shipment.ID)
	if err != nil {
		return nil, err
	}

	containers, err := r.getShipmentContainers(ctx, shipment.ID)
	if err != nil {
		return nil, err
	}

	routeData, err := r.getShipmentRouteData(ctx, shipment.ID, shipment.ID)
	if err != nil {
		return nil, err
	}

	response := &dto.ShipmentDetailsResponse{
		ID:             shipment.ID,
		ShipmentType:   shipment.ShipmentType,
		ShipmentNumber: shipment.ShipmentNumber,
		SealineCode:    shipment.SealineCode,
		SealineName:    shipment.SealineName,
		ShippingStatus: shipment.ShippingStatus,
		CreatedAt:      shipment.CreatedAt,
		UpdatedAt:      shipment.UpdatedAt,

		Consignee:        &shipment.Consignee,
		Recipient:        &shipment.Recipient,
		AssignedTo:       &shipment.AssignedTo,
		PlaceOfLoading:   &shipment.PlaceOfLoading,
		PlaceOfDelivery:  &shipment.PlaceOfDelivery,
		FinalDestination: &shipment.FinalDestination,
		ContainerType:    &shipment.ContainerType,
		Shipper:          &shipment.Shipper,
		InvoiceAmount:    &shipment.InvoiceAmount,
		Cost:             &shipment.Cost,
		Customs:          &shipment.Customs,
		MBL:              &shipment.MBL,
		Notes:            &shipment.Notes,
		CustomsProcessed: shipment.CustomsProcessed,
		Invoiced:         shipment.Invoiced,
		PaymentReceived:  shipment.PaymentReceived,

		Locations:  locations,
		Route:      route,
		Vessels:    vessels,
		Facilities: facilities,
		Containers: containers,
		RouteData:  routeData,
	}

	shipmentDate, err := r.GetShipmentDate(containers)
	if err == nil {
		response.ShipmentDate = shipmentDate
	}

	return response, nil
}

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

	return r.ConvertLocationsToDTO(locations), nil
}

func (r *shipmentRepository) getShipmentRoute(ctx context.Context, shipmentID uuid.UUID) (dto.ShipmentRouteResponse, error) {
	var routes []models.ShipmentRoute
	err := r.db.DB.WithContext(ctx).
		Preload("Location").
		Where("shipment_id = ?", shipmentID).
		Find(&routes).Error
	if err != nil {
		return dto.ShipmentRouteResponse{}, err
	}

	return r.ConvertRouteToDTO(routes), nil
}

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

	return r.ConvertVesselsToDTO(vessels), nil
}

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

	return r.ConvertFacilitiesToDTO(facilities), nil
}

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
		eventResponse, err := r.ConvertContainerEventToDTO(ctx, event)
		if err != nil {
			return nil, err
		}
		eventResponses[i] = eventResponse
	}

	return eventResponses, nil
}

func (r *shipmentRepository) ConvertContainerEventToDTO(ctx context.Context, event models.ContainerEvent) (dto.ShipmentContainerEventResponse, error) {

	var locationResponse dto.ShipmentLocationResponse

	if event.LocationID != uuid.Nil {
		if event.Location.ID != uuid.Nil {
			locationResponse = r.ConvertLocationToDTO(event.Location)
		} else {
			location, err := r.FindLocationByID(ctx, event.LocationID)
			if err != nil {
				return dto.ShipmentContainerEventResponse{}, err
			}
			locationResponse = r.ConvertLocationToDTO(*location)
		}
	}

	eventResponse := dto.ShipmentContainerEventResponse{
		Location:          locationResponse,
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
		if event.Facility != nil && event.Facility.ID != uuid.Nil {
			facilityResponse := r.ConvertFacilityToDTO(*event.Facility)
			eventResponse.Facility = &facilityResponse
		} else {
			fac, err := r.FindFacilityByID(ctx, event.FacilityID)
			if err == nil {
				fDto := r.ConvertFacilityToDTO(*fac)
				eventResponse.Facility = &fDto
			}
		}
	}

	if event.VesselID != nil {
		if event.Vessel != nil && event.Vessel.ID != uuid.Nil {
			vesselResponse := r.ConvertVesselToDTO(*event.Vessel)
			eventResponse.Vessel = &vesselResponse
		} else {
			ves, err := r.FindVesselByID(ctx, event.VesselID)
			if err == nil {
				vDto := r.ConvertVesselToDTO(*ves)
				eventResponse.Vessel = &vDto
			}
		}
	}

	return eventResponse, nil

}

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
		segmentResponse, err := r.ConvertRouteSegmentToDTO(ctx, segment)
		if err != nil {
			return nil, err
		}
		routeSegmentsResponse = append(routeSegmentsResponse, segmentResponse)
	}

	return routeSegmentsResponse, nil
}

func (r *shipmentRepository) ConvertRouteSegmentToDTO(ctx context.Context, segment models.RouteSegment) (dto.ShipmentRouteSegmentResponse, error) {
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

func (r *shipmentRepository) ConvertLocationsToDTO(locations []models.Location) []dto.ShipmentLocationResponse {
	locationResponses := make([]dto.ShipmentLocationResponse, len(locations))
	for i, loc := range locations {
		locationResponses[i] = r.ConvertLocationToDTO(loc)
	}
	return locationResponses
}

func (r *shipmentRepository) ConvertLocationToDTO(location models.Location) dto.ShipmentLocationResponse {
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

func (r *shipmentRepository) ConvertRouteToDTO(routes []models.ShipmentRoute) dto.ShipmentRouteResponse {
	routeResponse := dto.ShipmentRouteResponse{}
	for _, route := range routes {
		item := &dto.ShipmentRoutePoint{
			Location:      r.ConvertLocationToDTO(route.Location),
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

func (r *shipmentRepository) ConvertVesselsToDTO(vessels []models.Vessel) []dto.ShipmentVesselResponse {
	vesselResponses := make([]dto.ShipmentVesselResponse, len(vessels))
	for i, vessel := range vessels {
		vesselResponses[i] = r.ConvertVesselToDTO(vessel)
	}
	return vesselResponses
}

func (r *shipmentRepository) ConvertVesselToDTO(vessel models.Vessel) dto.ShipmentVesselResponse {
	return dto.ShipmentVesselResponse{
		Name:     vessel.Name,
		Imo:      vessel.Imo,
		Mmsi:     vessel.Mmsi,
		CallSign: vessel.CallSign,
		Flag:     vessel.Flag,
	}
}

func (r *shipmentRepository) ConvertFacilitiesToDTO(facilities []models.Facility) []dto.ShipmentFacilityResponse {
	facilityResponses := make([]dto.ShipmentFacilityResponse, len(facilities))
	for i, facility := range facilities {
		facilityResponses[i] = r.ConvertFacilityToDTO(facility)
	}
	return facilityResponses
}

func (r *shipmentRepository) ConvertFacilityToDTO(facility models.Facility) dto.ShipmentFacilityResponse {
	return dto.ShipmentFacilityResponse{
		Name:        facility.Name,
		CountryCode: facility.CountryCode,
		Locode:      facility.Locode,
		BicCode:     facility.BicCode,
		SmdgCode:    facility.SmdgCode,
		Latitude:    *facility.Latitude,
		Longitude:   *facility.Longitude,
	}
}

func (r *shipmentRepository) DeleteShipmentLocations(ctx context.Context, shipmentID uuid.UUID) error {
	if shipmentID == uuid.Nil {
		return fmt.Errorf("invalid shipment ID: cannot be nil")
	}

	db := r.getDBFromContext(ctx)
	result := db.WithContext(ctx).
		Where("shipment_id = ?", shipmentID).
		Delete(&models.ShipmentLocation{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete shipment locations: %w", result.Error)
	}

	log.Printf("Deleted %d location relationships for shipment %s", result.RowsAffected, shipmentID)
	return nil
}

func (r *shipmentRepository) DeleteShipmentRoutes(ctx context.Context, shipmentID uuid.UUID) error {
	if shipmentID == uuid.Nil {
		return fmt.Errorf("invalid shipment ID: cannot be nil")
	}

	db := r.getDBFromContext(ctx)
	result := db.WithContext(ctx).
		Where("shipment_id = ?", shipmentID).
		Delete(&models.ShipmentRoute{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete shipment routes: %w", result.Error)
	}

	log.Printf("Deleted %d routes for shipment %s", result.RowsAffected, shipmentID)
	return nil
}

func (r *shipmentRepository) DeleteShipmentVessels(ctx context.Context, shipmentID uuid.UUID) error {
	if shipmentID == uuid.Nil {
		return fmt.Errorf("invalid shipment ID: cannot be nil")
	}

	db := r.getDBFromContext(ctx)
	result := db.WithContext(ctx).
		Where("shipment_id = ?", shipmentID).
		Delete(&models.ShipmentVessel{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete shipment vessels: %w", result.Error)
	}

	log.Printf("Deleted %d vessel relationships for shipment %s", result.RowsAffected, shipmentID)
	return nil
}

func (r *shipmentRepository) DeleteShipmentFacilities(ctx context.Context, shipmentID uuid.UUID) error {
	if shipmentID == uuid.Nil {
		return fmt.Errorf("invalid shipment ID: cannot be nil")
	}

	db := r.getDBFromContext(ctx)
	result := db.WithContext(ctx).
		Where("shipment_id = ?", shipmentID).
		Delete(&models.ShipmentFacility{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete shipment facilities: %w", result.Error)
	}

	log.Printf("Deleted %d facility relationships for shipment %s", result.RowsAffected, shipmentID)
	return nil
}

func (r *shipmentRepository) DeleteShipmentContainers(ctx context.Context, shipmentID uuid.UUID) error {
	if shipmentID == uuid.Nil {
		return fmt.Errorf("invalid shipment ID: cannot be nil")
	}

	db := r.getDBFromContext(ctx)
	result := db.WithContext(ctx).
		Where("shipment_id = ?", shipmentID).
		Delete(&models.ShipmentContainer{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete shipment containers: %w", result.Error)
	}

	log.Printf("Deleted %d container relationships for shipment %s", result.RowsAffected, shipmentID)
	return nil
}

func (r *shipmentRepository) DeleteRouteSegments(ctx context.Context, shipmentID uuid.UUID) error {
	if shipmentID == uuid.Nil {
		return fmt.Errorf("invalid shipment ID: cannot be nil")
	}

	db := r.getDBFromContext(ctx)
	result := db.WithContext(ctx).
		Where("shipment_id = ?", shipmentID).
		Delete(&models.RouteSegment{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete route segments: %w", result.Error)
	}

	log.Printf("Deleted %d route segments for shipment %s", result.RowsAffected, shipmentID)
	return nil
}

func (r *shipmentRepository) DeleteShipmentCoordinates(ctx context.Context, shipmentID uuid.UUID) error {
	if shipmentID == uuid.Nil {
		return fmt.Errorf("invalid shipment ID: cannot be nil")
	}

	db := r.getDBFromContext(ctx)
	result := db.WithContext(ctx).
		Where("shipment_id = ?", shipmentID).
		Delete(&models.Coordinate{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete coordinates: %w", result.Error)
	}

	log.Printf("Deleted %d coordinates for shipment %s", result.RowsAffected, shipmentID)
	return nil
}

func (r *shipmentRepository) DeleteShipmentAis(ctx context.Context, shipmentID uuid.UUID) error {
	if shipmentID == uuid.Nil {
		return fmt.Errorf("invalid shipment ID: cannot be nil")
	}

	db := r.getDBFromContext(ctx)
	result := db.WithContext(ctx).
		Where("shipment_id = ?", shipmentID).
		Delete(&models.Ais{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete AIS data: %w", result.Error)
	}

	log.Printf("Deleted %d AIS records for shipment %s", result.RowsAffected, shipmentID)
	return nil
}

func (r *shipmentRepository) DeleteAllShipmentRelatedData(ctx context.Context, shipmentID uuid.UUID) error {
	if shipmentID == uuid.Nil {
		return fmt.Errorf("invalid shipment ID: cannot be nil")
	}

	log.Printf("Starting cleanup of all related data for shipment ID: %s", shipmentID)

	log.Printf("Deleting AIS data for shipment %s", shipmentID)
	if err := r.DeleteShipmentAis(ctx, shipmentID); err != nil {
		log.Printf("Failed to delete AIS data for shipment %s: %v", shipmentID, err)
		return fmt.Errorf("failed to delete AIS data: %w", err)
	}

	log.Printf("Deleting coordinates for shipment %s", shipmentID)
	if err := r.DeleteShipmentCoordinates(ctx, shipmentID); err != nil {
		log.Printf("Failed to delete coordinates for shipment %s: %v", shipmentID, err)
		return fmt.Errorf("failed to delete coordinates: %w", err)
	}

	log.Printf("Deleting route segments for shipment %s", shipmentID)
	if err := r.DeleteRouteSegments(ctx, shipmentID); err != nil {
		log.Printf("Failed to delete route segments for shipment %s: %v", shipmentID, err)
		return fmt.Errorf("failed to delete route segments: %w", err)
	}

	log.Printf("Deleting container relationships for shipment %s", shipmentID)
	if err := r.DeleteShipmentContainers(ctx, shipmentID); err != nil {
		log.Printf("Failed to delete containers for shipment %s: %v", shipmentID, err)
		return fmt.Errorf("failed to delete containers: %w", err)
	}

	log.Printf("Deleting facility relationships for shipment %s", shipmentID)
	if err := r.DeleteShipmentFacilities(ctx, shipmentID); err != nil {
		log.Printf("Failed to delete facilities for shipment %s: %v", shipmentID, err)
		return fmt.Errorf("failed to delete facilities: %w", err)
	}

	log.Printf("Deleting vessel relationships for shipment %s", shipmentID)
	if err := r.DeleteShipmentVessels(ctx, shipmentID); err != nil {
		log.Printf("Failed to delete vessels for shipment %s: %v", shipmentID, err)
		return fmt.Errorf("failed to delete vessels: %w", err)
	}

	log.Printf("Deleting routes for shipment %s", shipmentID)
	if err := r.DeleteShipmentRoutes(ctx, shipmentID); err != nil {
		log.Printf("Failed to delete routes for shipment %s: %v", shipmentID, err)
		return fmt.Errorf("failed to delete routes: %w", err)
	}

	log.Printf("Deleting location relationships for shipment %s", shipmentID)
	if err := r.DeleteShipmentLocations(ctx, shipmentID); err != nil {
		log.Printf("Failed to delete locations for shipment %s: %v", shipmentID, err)
		return fmt.Errorf("failed to delete locations: %w", err)
	}

	log.Printf("Successfully deleted all related data for shipment %s", shipmentID)
	return nil
}

func (r *shipmentRepository) GetDB() *db.Database {
	return r.db
}

func (r *shipmentRepository) getDBFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok {
		return tx
	}
	return r.db.DB
}

type ShipmentDataSummary struct {
	ShipmentID           uuid.UUID `json:"shipmentId"`
	LocationsCount       int64     `json:"locationsCount"`
	RoutesCount          int64     `json:"routesCount"`
	VesselsCount         int64     `json:"vesselsCount"`
	FacilitiesCount      int64     `json:"facilitiesCount"`
	ContainersCount      int64     `json:"containersCount"`
	ContainerEventsCount int64     `json:"containerEventsCount"`
	RouteSegmentsCount   int64     `json:"routeSegmentsCount"`
	CoordinatesCount     int64     `json:"coordinatesCount"`
	AisCount             int64     `json:"aisCount"`
}

func (r *shipmentRepository) GetShipmentDataSummary(ctx context.Context, shipmentID uuid.UUID) (*ShipmentDataSummary, error) {
	if shipmentID == uuid.Nil {
		return nil, fmt.Errorf("invalid shipment ID: cannot be nil")
	}

	db := r.getDBFromContext(ctx)
	summary := &ShipmentDataSummary{ShipmentID: shipmentID}

	if err := db.WithContext(ctx).Model(&models.ShipmentLocation{}).Where("shipment_id = ?", shipmentID).Count(&summary.LocationsCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count locations: %w", err)
	}

	if err := db.WithContext(ctx).Model(&models.ShipmentRoute{}).Where("shipment_id = ?", shipmentID).Count(&summary.RoutesCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count routes: %w", err)
	}

	if err := db.WithContext(ctx).Model(&models.ShipmentVessel{}).Where("shipment_id = ?", shipmentID).Count(&summary.VesselsCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count vessels: %w", err)
	}

	if err := db.WithContext(ctx).Model(&models.ShipmentFacility{}).Where("shipment_id = ?", shipmentID).Count(&summary.FacilitiesCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count facilities: %w", err)
	}

	if err := db.WithContext(ctx).Model(&models.ShipmentContainer{}).Where("shipment_id = ?", shipmentID).Count(&summary.ContainersCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count containers: %w", err)
	}

	var containerIDs []uuid.UUID
	if err := db.WithContext(ctx).Model(&models.ShipmentContainer{}).Where("shipment_id = ?", shipmentID).Pluck("container_id", &containerIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to get container IDs: %w", err)
	}
	if len(containerIDs) > 0 {
		if err := db.WithContext(ctx).Model(&models.ContainerEvent{}).Where("container_id IN ?", containerIDs).Count(&summary.ContainerEventsCount).Error; err != nil {
			return nil, fmt.Errorf("failed to count container events: %w", err)
		}
	}

	if err := db.WithContext(ctx).Model(&models.RouteSegment{}).Where("shipment_id = ?", shipmentID).Count(&summary.RouteSegmentsCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count route segments: %w", err)
	}

	if err := db.WithContext(ctx).Model(&models.Coordinate{}).Where("shipment_id = ?", shipmentID).Count(&summary.CoordinatesCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count coordinates: %w", err)
	}

	if err := db.WithContext(ctx).Model(&models.Ais{}).Where("shipment_id = ?", shipmentID).Count(&summary.AisCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count AIS records: %w", err)
	}

	return summary, nil
}

func (r *shipmentRepository) GetShipmentsForGrid(ctx context.Context, userID uuid.UUID) ([]models.Shipment, error) {
	db := r.db.DB.WithContext(ctx)

	query := db.Model(&models.Shipment{}).
		Joins("JOIN user_shipments us ON us.shipment_id = shipments.id").
		Where("us.user_id = ?", userID)

	var shipments []models.Shipment
	if err := query.Find(&shipments).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch shipments: %w", err)
	}

	return shipments, nil
}

func (r *shipmentRepository) UpdateShipmentInfo(ctx context.Context, userID, shipmentID uuid.UUID, req *dto.UpdateShipmentInfoRequest) error {
	db := r.db.DB.WithContext(ctx)

	result := db.Model(&models.Shipment{}).
		Where("id = ?", shipmentID).
		Updates(map[string]interface{}{
			"consignee":         req.Consignee,
			"recipient":         req.Recipient,
			"assigned_to":       req.AssignedTo,
			"place_of_loading":  req.PlaceOfLoading,
			"place_of_delivery": req.PlaceOfDelivery,
			"final_destination": req.FinalDestination,
			"container_type":    req.ContainerType,
			"shipper":           req.Shipper,
			"invoice_amount":    req.InvoiceAmount,
			"cost":              req.Cost,
			"customs":           req.Customs,
			"mbl":               req.MBL,
			"notes":             req.Notes,
			"customs_processed": req.CustomsProcessed,
			"invoiced":          req.Invoiced,
			"payment_received":  req.PaymentReceived,
			"updated_at":        time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update shipment info: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("shipment not found")
	}

	return nil
}

func (r *shipmentRepository) UpdateShipmentInfoPartial(ctx context.Context, userID, shipmentID uuid.UUID, updates map[string]interface{}) error {
	db := r.db.DB.WithContext(ctx)

	validUpdates := make(map[string]any)

	fieldMappings := map[string]string{
		"consignee":        "consignee",
		"recipient":        "recipient",
		"assignedTo":       "assigned_to",
		"placeOfLoading":   "place_of_loading",
		"placeOfDelivery":  "place_of_delivery",
		"finalDestination": "final_destination",
		"containerType":    "container_type",
		"shipper":          "shipper",
		"invoiceAmount":    "invoice_amount",
		"cost":             "cost",
		"customs":          "customs",
		"mbl":              "mbl",
		"notes":            "notes",
		"customsProcessed": "customs_processed",
		"invoiced":         "invoiced",
		"paymentReceived":  "payment_received",
	}

	for jsonField, value := range updates {
		if dbField, exists := fieldMappings[jsonField]; exists {
			validUpdates[dbField] = value
		}
	}

	validUpdates["updated_at"] = time.Now()

	if len(validUpdates) == 1 {
		return fmt.Errorf("no valid fields to update")
	}

	result := db.Model(&models.Shipment{}).
		Where("id = ?", shipmentID).
		Updates(validUpdates)

	if result.Error != nil {
		return fmt.Errorf("failed to update shipment info: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("shipment not found")
	}

	return nil
}

func (r *shipmentRepository) DeleteUserShipment(ctx context.Context, userID, shipmentID uuid.UUID) error {
	db := r.getDBFromContext(ctx)

	result := db.WithContext(ctx).
		Where("user_id = ? AND shipment_id = ?", userID, shipmentID).
		Delete(&models.UserShipment{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete user shipment: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("shipment not found")
	}

	return nil
}

func (r *shipmentRepository) BulkDeleteUserShipments(ctx context.Context, userID uuid.UUID, shipmentIDs []uuid.UUID) error {
	if len(shipmentIDs) == 0 {
		return fmt.Errorf("no shipment IDs provided")
	}

	db := r.getDBFromContext(ctx)

	result := db.WithContext(ctx).
		Where("user_id = ? AND shipment_id IN ?", userID, shipmentIDs).
		Delete(&models.UserShipment{})

	if result.Error != nil {
		return fmt.Errorf("failed to bulk delete user shipments: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no shipments found to delete")
	}

	return nil
}

type ShipmentForRefresh struct {
	models.Shipment
}

func (r *shipmentRepository) GetAllShipmentsForRefresh(ctx context.Context, skipRecentlyUpdated time.Duration) ([]ShipmentForRefresh, error) {
	db := r.getDBFromContext(ctx)

	var results []ShipmentForRefresh

	query := db.Table("shipments").
		Where("shipping_status != ?", "DELIVERED")

	if skipRecentlyUpdated > 0 {
		cutoffTime := time.Now().Add(-skipRecentlyUpdated)
		query = query.Where("updated_at < ?", cutoffTime)
	}

	if err := query.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch shipments for refresh: %w", err)
	}

	return results, nil
}

func (r *shipmentRepository) GetShipmentsWithDetailsForGrid(ctx context.Context, userID uuid.UUID) ([]models.Shipment, error) {
	var shipments []models.Shipment

	err := r.db.DB.WithContext(ctx).
		Joins("JOIN user_shipments us ON us.shipment_id = shipments.id").
		Where("us.user_id = ?", userID).
		Preload("Locations").
		Preload("Routes.Location").
		Preload("Vessels").
		Preload("Facilities").
		Preload("Containers").
		Preload("Containers.Events.Location").
		Preload("Containers.Events.Facility").
		Preload("Containers.Events.Vessel").
		Preload("RouteSegments").
		Preload("RouteSegments.Points").
		Preload("Coordinates").
		Preload("Ais").
		Find(&shipments).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch shipments for grid: %w", err)
	}

	return shipments, nil
}

func (r *shipmentRepository) GetShipmentDate(containers []dto.ShipmentContainerResponse) (*time.Time, error) {
	var minDate *time.Time

	for _, container := range containers {
		for _, event := range container.Events {
			if event.Date.IsZero() {
				continue
			}

			currentDate := event.Date

			if minDate == nil || currentDate.Before(*minDate) {
				t := currentDate
				minDate = &t
			}
		}
	}

	if minDate != nil {
		return minDate, nil
	}

	return nil, fmt.Errorf("no pickup (PICK) event found in shipment data")
}

func (r *shipmentRepository) MapShipmentToDTO(ctx context.Context, shipment models.Shipment) dto.ShipmentDetailsResponse {

	containerdtos := make([]dto.ShipmentContainerResponse, len(shipment.Containers))
	for j, cont := range shipment.Containers {
		eventDtos := make([]dto.ShipmentContainerEventResponse, len(cont.Events))
		for k, evt := range cont.Events {
			mappedEvent, _ := r.ConvertContainerEventToDTO(ctx, evt)
			eventDtos[k] = mappedEvent
		}
		containerdtos[j] = dto.ShipmentContainerResponse{
			Number:   cont.Number,
			IsoCode:  cont.IsoCode,
			SizeType: cont.SizeType,
			Status:   cont.Status,
			Events:   eventDtos,
		}
	}

	routeSegDTOs := make([]dto.ShipmentRouteSegmentResponse, len(shipment.RouteSegments))
	for k, seg := range shipment.RouteSegments {
		pathDTOs := make([]dto.ShipmentRouteSegmentPointResponse, len(seg.Points))
		for m, pt := range seg.Points {
			pathDTOs[m] = dto.ShipmentRouteSegmentPointResponse{
				Latitude:   pt.Latitude,
				Longitude:  pt.Longitude,
				PointOrder: pt.PointOrder,
			}
		}

		routeSegDTOs[k] = dto.ShipmentRouteSegmentResponse{
			RouteType:    seg.RouteType,
			SegmentOrder: seg.SegmentOrder,
			Path:         pathDTOs,
		}
	}

	var coordDTO dto.ShipmentCoordinatesResponse
	if len(shipment.Coordinates) > 0 {
		latestCoord := shipment.Coordinates[0]
		for _, c := range shipment.Coordinates {
			if c.UpdatedAt.After(latestCoord.UpdatedAt) {
				latestCoord = c
			}
		}
		coordDTO = dto.ShipmentCoordinatesResponse{
			Latitude:  latestCoord.Latitude,
			Longitude: latestCoord.Longitude,
			UpdatedAt: latestCoord.UpdatedAt,
		}
	}

	var aisDTO dto.ShipmentAisResponse
	if shipment.Ais != nil {
		aisDTO = dto.ShipmentAisResponse{
			Status:                   shipment.Ais.Status,
			LastEventDescription:     shipment.Ais.LastEventDescription,
			LastEventDate:            shipment.Ais.LastEventDate,
			LastEventVoyage:          shipment.Ais.LastEventVoyage,
			DischargePortName:        shipment.Ais.DischargePortName,
			DischargePortCountryCode: shipment.Ais.DischargePortCountryCode,
			DischargePortCode:        shipment.Ais.DischargePortCode,
			DischargePortDate:        shipment.Ais.DischargePortDate,
			DischargePortDateLabel:   shipment.Ais.DischargePortDateLabel,
			DeparturePortName:        shipment.Ais.DeparturePortName,
			DeparturePortCountryCode: shipment.Ais.DeparturePortCountryCode,
			DeparturePortCode:        shipment.Ais.DeparturePortCode,
			DeparturePortDate:        shipment.Ais.DeparturePortDate,
			DeparturePortDateLabel:   shipment.Ais.DeparturePortDateLabel,
			ArrivalPortName:          shipment.Ais.ArrivalPortName,
			ArrivalPortCountryCode:   shipment.Ais.ArrivalPortCountryCode,
			ArrivalPortCode:          shipment.Ais.ArrivalPortCode,
			ArrivalPortDate:          shipment.Ais.ArrivalPortDate,
			ArrivalPortDateLabel:     shipment.Ais.ArrivalPortDateLabel,
			LastVesselPositionLat:    shipment.Ais.LastVesselPositionLat,
			LastVesselPositionLng:    shipment.Ais.LastVesselPositionLng,
			LastVesselPositionUpdate: shipment.Ais.LastVesselPositionUpdate,
			UpdatedAt:                shipment.Ais.UpdatedAt,
		}
		if shipment.Ais.Vessel != nil {
			v := r.ConvertVesselToDTO(*shipment.Ais.Vessel)
			aisDTO.Vessel = &v
		}
	}

	response := dto.ShipmentDetailsResponse{
		ID:             shipment.ID,
		ShipmentNumber: shipment.ShipmentNumber,
		ShipmentType:   shipment.ShipmentType,
		SealineCode:    shipment.SealineCode,
		SealineName:    shipment.SealineName,
		ShippingStatus: shipment.ShippingStatus,
		CreatedAt:      shipment.CreatedAt,
		UpdatedAt:      shipment.UpdatedAt,

		Consignee:        &shipment.Consignee,
		Recipient:        &shipment.Recipient,
		AssignedTo:       &shipment.AssignedTo,
		PlaceOfLoading:   &shipment.PlaceOfLoading,
		PlaceOfDelivery:  &shipment.PlaceOfDelivery,
		FinalDestination: &shipment.FinalDestination,
		ContainerType:    &shipment.ContainerType,
		Shipper:          &shipment.Shipper,
		InvoiceAmount:    &shipment.InvoiceAmount,
		Cost:             &shipment.Cost,
		Customs:          &shipment.Customs,
		MBL:              &shipment.MBL,
		Notes:            &shipment.Notes,

		CustomsProcessed: shipment.CustomsProcessed,
		Invoiced:         shipment.Invoiced,
		PaymentReceived:  shipment.PaymentReceived,

		Locations:  r.ConvertLocationsToDTO(shipment.Locations),
		Route:      r.ConvertRouteToDTO(shipment.Routes),
		Vessels:    r.ConvertVesselsToDTO(shipment.Vessels),
		Facilities: r.ConvertFacilitiesToDTO(shipment.Facilities),
		Containers: containerdtos,

		RouteData: dto.ShipmentRouteDataResponse{
			RouteSegments: routeSegDTOs,
			Coordinates:   coordDTO,
			Ais:           aisDTO,
		},
	}

	if date, err := r.GetShipmentDate(containerdtos); err == nil {
		response.ShipmentDate = date
	}

	return response
}
