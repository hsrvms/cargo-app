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

func (r *shipmentRepository) GetShipmentDetails(ctx context.Context, shipmentID uuid.UUID) (*dto.ShipmentDetailsResponse, error) {
	shipment, err := r.GetShipmentByID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}

	var locations []models.Location
	err = r.db.DB.WithContext(ctx).
		Joins("JOIN shipment_locations sl ON sl.location_id = locations.id").
		Where("sl.shipment_id = ?", shipmentID).
		Order("sl.added_at ASC").
		Find(&locations).Error

	if err != nil {
		return nil, err
	}

	// Convert models.Location to dto.LocationResponse
	locationResponses := make([]dto.ShipmentLocationResponse, len(locations))
	for i, loc := range locations {
		locationResponses[i] = dto.ShipmentLocationResponse{
			Name:        loc.Name,
			State:       loc.State,
			Country:     loc.Country,
			CountryCode: loc.CountryCode,
			Locode:      loc.Locode,
			Latitude:    loc.Latitude,
			Longitude:   loc.Longitude,
			Timezone:    loc.Timezone,
		}
	}

	var routes []models.ShipmentRoute
	err = r.db.DB.WithContext(ctx).
		Preload("Location").
		Where("shipment_id = ?", shipmentID).
		Find(&routes).Error
	if err != nil {
		return nil, err
	}

	routeResponse := dto.ShipmentRouteResponse{}
	for _, r := range routes {
		item := &dto.ShipmentRoutePoint{
			Location: dto.ShipmentLocationResponse{
				Name:        r.Location.Name,
				State:       r.Location.State,
				Country:     r.Location.Country,
				CountryCode: r.Location.CountryCode,
				Locode:      r.Location.Locode,
				Latitude:    r.Location.Latitude,
				Longitude:   r.Location.Longitude,
				Timezone:    r.Location.Timezone,
			},
			Date:          r.Date,
			Actual:        r.Actual,
			PredictiveETA: r.PredictiveETA,
		}
		switch r.RouteType {
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

	var vessels []models.Vessel
	err = r.db.DB.WithContext(ctx).
		Joins("JOIN shipment_vessels sv ON sv.vessel_id = vessels.id").
		Where("sv.shipment_id = ?", shipmentID).
		Order("sv.added_at ASC").
		Find(&vessels).Error
	if err != nil {
		return nil, err
	}

	vesselResponses := make([]dto.ShipmentVesselResponse, len(vessels))
	for i, v := range vessels {
		vesselResponses[i] = dto.ShipmentVesselResponse{
			Name:     v.Name,
			Imo:      v.Imo,
			Mmsi:     v.Mmsi,
			CallSign: v.CallSign,
			Flag:     v.Flag,
		}
	}

	var facilities []models.Facility
	err = r.db.DB.WithContext(ctx).
		Joins("JOIN shipment_facilities sf ON sf.facility_id = facilities.id").
		Where("sf.shipment_id = ?", shipmentID).
		Order("sf.added_at ASC").
		Find(&facilities).Error
	if err != nil {
		return nil, err
	}

	facilityResponses := make([]dto.ShipmentFacilityResponse, len(facilities))
	for i, f := range facilities {
		facilityResponses[i] = dto.ShipmentFacilityResponse{
			Name:        f.Name,
			CountryCode: f.CountryCode,
			Locode:      f.Locode,
			BicCode:     f.BicCode,
			SmdgCode:    f.SmdgCode,
			Latitude:    f.Latitude,
			Longitude:   f.Longitude,
		}
	}

	var containers []models.Container
	err = r.db.DB.WithContext(ctx).
		Joins("JOIN shipment_containers sc ON sc.container_id = containers.id").
		Where("sc.shipment_id = ?", shipmentID).
		Order("sc.added_at ASC").
		Find(&containers).Error
	if err != nil {
		return nil, err
	}

	containersResponse := make([]dto.ShipmentContainerResponse, len(containers))
	for i, c := range containers {
		containersResponse[i] = dto.ShipmentContainerResponse{
			Number:   c.Number,
			IsoCode:  c.IsoCode,
			SizeType: c.SizeType,
			Status:   c.Status,
		}

		var containerEvents []models.ContainerEvent
		err = r.db.DB.WithContext(ctx).
			Where("container_id = ?", containers[i].ID).
			Find(&containerEvents).Error
		if err != nil {
			return nil, err
		}

		containerEventsResponse := make([]dto.ShipmentContainerEventResponse, len(containerEvents))
		for i, ce := range containerEvents {
			location, err := r.FindLocationByID(ctx, ce.LocationID)
			if err != nil {
				return nil, err
			}
			locationResponse := dto.ShipmentLocationResponse{
				Name:        location.Name,
				State:       location.State,
				Country:     location.Country,
				CountryCode: location.CountryCode,
				Locode:      location.Locode,
				Latitude:    location.Latitude,
				Longitude:   location.Longitude,
				Timezone:    location.Timezone,
			}

			containerEventsResponse[i] = dto.ShipmentContainerEventResponse{
				Location: locationResponse,
				// Facility: ce.FacilityID,
				Description:       ce.Description,
				EventType:         ce.EventType,
				EventCode:         ce.EventCode,
				Status:            ce.Status,
				Date:              ce.Date,
				IsActual:          ce.IsActual,
				IsAdditionalEvent: ce.IsAdditionalEvent,
				RouteType:         ce.RouteType,
				TransportType:     ce.TransportType,
				// Vessel:            ce.VesselID,
				Voyage: ce.Voyage,
			}

			if ce.FacilityID != nil {
				facility, err := r.FindFacilityByID(ctx, ce.FacilityID)
				if err != nil {
					return nil, err
				}

				facilityResponse := dto.ShipmentFacilityResponse{
					Name:        facility.Name,
					CountryCode: facility.CountryCode,
					Locode:      facility.Locode,
					BicCode:     facility.BicCode,
					SmdgCode:    facility.SmdgCode,
					Latitude:    facility.Latitude,
					Longitude:   facility.Longitude,
				}
				containerEventsResponse[i].Facility = &facilityResponse
			}

			if ce.VesselID != nil {
				vessel, err := r.FindVesselByID(ctx, ce.VesselID)
				if err != nil {
					return nil, err
				}

				vesselResponse := dto.ShipmentVesselResponse{
					Name:     vessel.Name,
					Imo:      vessel.Imo,
					Mmsi:     vessel.Mmsi,
					CallSign: vessel.CallSign,
					Flag:     vessel.Flag,
				}
				containerEventsResponse[i].Vessel = &vesselResponse
			}

		}

		containersResponse[i].Events = containerEventsResponse

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
		Locations:      locationResponses,
		Route:          routeResponse,
		Vessels:        vesselResponses,
		Facilities:     facilityResponses,
		Containers:     containersResponse,
	}, nil
}
