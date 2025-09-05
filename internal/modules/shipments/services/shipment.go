package services

import (
	"context"
	"errors"
	"fmt"
	"go-starter/internal/modules/shipments/dto"
	shipmentsDto "go-starter/internal/modules/shipments/dto"
	"go-starter/internal/modules/shipments/models"
	"go-starter/internal/modules/shipments/repositories"
	"go-starter/internal/modules/shipments/types"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

		shipment, err = s.SyncShipment(ctx, userID, shipment.ID)
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
			Latitude:    &f.Coordinates.Lat,
			Longitude:   &f.Coordinates.Lng,
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

		containerEvents := make([]models.ContainerEvent, 0, len(c.Events))
		for _, ce := range c.Events {
			location, err := s.repo.FindLocationByLocode(ctx, ce.Location.Locode)
			if err != nil {
				return nil, err
			}

			var facility *models.Facility
			if ce.Facility != nil {
				facility, err = s.repo.FindFacilityByLocode(ctx, ce.Facility.Locode)
				if err != nil {
					return nil, err
				}
			}

			var vessel *models.Vessel
			if ce.Vessel != nil {
				vessel, err = s.repo.FindVesselByIMOAndMMSI(ctx, ce.Vessel.Imo, ce.Vessel.Mmsi)
				if err != nil {
					if !errors.Is(err, gorm.ErrRecordNotFound) {
						return nil, err
					}
				}
			}

			containerEvent := models.ContainerEvent{
				ContainerID:       createdContainer.ID,
				LocationID:        location.ID,
				Description:       ce.Description,
				EventType:         ce.EventType,
				EventCode:         ce.EventCode,
				Status:            ce.Status,
				Date:              ce.Date,
				IsActual:          ce.IsActual,
				IsAdditionalEvent: ce.IsAdditionalEvent,
				RouteType:         ce.RouteType,
				TransportType:     ce.TransportType,
				Voyage:            ce.Voyage,
			}

			if vessel != nil {
				containerEvent.VesselID = &vessel.ID
			}

			if facility != nil {
				containerEvent.FacilityID = &facility.ID
			}

			createdContainerEvent, err := s.repo.CreateContainerEvent(ctx, &containerEvent)
			if err != nil {
				return nil, err
			}
			containerEvents = append(containerEvents, *createdContainerEvent)
		}
	}

	routeSegments := make([]models.RouteSegment, 0, len(apiResponse.RouteData.RouteSegments))
	for segIdx, rs := range apiResponse.RouteData.RouteSegments {
		segment := models.RouteSegment{
			ShipmentID:   shipment.ID,
			RouteType:    rs.RouteType,
			SegmentOrder: segIdx,
		}
		createdSegment, err := s.repo.CreateRouteSegment(ctx, &segment)
		if err != nil {
			return nil, err
		}
		routeSegments = append(routeSegments, *createdSegment)

		segmentPoints := make([]models.RouteSegmentPoint, 0, len(rs.Path))
		for pointIdx, point := range rs.Path {
			point := models.RouteSegmentPoint{
				SegmentID:  segment.ID,
				Latitude:   point.Lat,
				Longitude:  point.Lng,
				PointOrder: pointIdx,
			}
			createdPoint, err := s.repo.CreateRouteSegmentPoint(ctx, &point)
			if err != nil {
				return nil, err
			}
			segmentPoints = append(segmentPoints, *createdPoint)
		}
	}

	coordinate := &models.Coordinate{
		ShipmentID: shipment.ID,
		Latitude:   apiResponse.RouteData.Coordinates.Lat,
		Longitude:  apiResponse.RouteData.Coordinates.Lng,
	}

	_, err = s.repo.CreateCoordinate(ctx, coordinate)
	if err != nil {
		return nil, err
	}

	if apiResponse.RouteData.Ais.Data != nil {
		aisData := apiResponse.RouteData.Ais.Data
		aisVessel, err := s.repo.FindVesselByIMOAndMMSI(ctx, aisData.Vessel.Imo, aisData.Vessel.Mmsi)
		if err != nil {
			return nil, err
		}
		ais := &models.Ais{
			ShipmentID:               shipment.ID,
			Status:                   apiResponse.RouteData.Ais.Status,
			LastEventDescription:     aisData.LastEvent.Description,
			LastEventDate:            aisData.LastEvent.Date,
			LastEventVoyage:          aisData.LastEvent.Voyage,
			DischargePortName:        aisData.DischargePort.Name,
			DischargePortCountryCode: aisData.DischargePort.CountryCode,
			DischargePortCode:        aisData.DischargePort.Code,
			DischargePortDate:        aisData.DischargePort.Date,
			DischargePortDateLabel:   aisData.DischargePort.DateLabel,
			DeparturePortName:        aisData.DeparturePort.Name,
			DeparturePortCountryCode: aisData.DeparturePort.CountryCode,
			DeparturePortCode:        aisData.DeparturePort.Code,
			DeparturePortDate:        aisData.DeparturePort.Date,
			DeparturePortDateLabel:   aisData.DeparturePort.DateLabel,
			ArrivalPortName:          aisData.ArrivalPort.Name,
			ArrivalPortCountryCode:   aisData.ArrivalPort.CountryCode,
			ArrivalPortCode:          aisData.ArrivalPort.Code,
			ArrivalPortDate:          aisData.ArrivalPort.Date,
			ArrivalPortDateLabel:     aisData.ArrivalPort.DateLabel,
			VesselID:                 &aisVessel.ID,
			LastVesselPositionLat:    aisData.LastVesselPosition.Lat,
			LastVesselPositionLng:    aisData.LastVesselPosition.Lng,
			LastVesselPositionUpdate: aisData.LastVesselPosition.UpdatedAt,
		}
		_, err = s.repo.CreateAis(ctx, ais)
		if err != nil {
			return nil, err
		}
	} else {
		ais := &models.Ais{
			ShipmentID: shipment.ID,
			Status:     apiResponse.RouteData.Ais.Status,
		}
		_, err := s.repo.CreateAis(ctx, ais)
		if err != nil {
			return nil, err
		}
	}

	log.Printf("Created shipment %s in database with ID: %s", req.ShipmentNumber, shipment.ID)

	return shipment, nil
}

func (s *shipmentService) GetShipmentsForGrid(ctx context.Context, userID uuid.UUID) (*dto.GridDataResponse, error) {
	shipments, err := s.repo.GetShipmentsForGrid(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shipments for grid: %w", err)
	}

	gridShipments := make([]dto.GridShipment, len(shipments))
	for i, shipment := range shipments {
		gridShipments[i] = dto.GridShipment{
			ID:             shipment.ID,
			ShipmentNumber: shipment.ShipmentNumber,
			ShipmentType:   shipment.ShipmentType,
			SealineCode:    shipment.SealineCode,
			SealineName:    shipment.SealineName,
			ShippingStatus: shipment.ShippingStatus,
			CreatedAt:      shipment.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      shipment.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return &dto.GridDataResponse{
		Rows: gridShipments,
	}, nil
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

	shipment, err := s.repo.GetShipmentByID(ctx, userID, shipmentID)
	if err != nil {
		return nil, err
	}
	return shipment, nil
}

// validateShipmentForSync validates that a shipment exists and can be synced
func (s *shipmentService) validateShipmentForSync(ctx context.Context, userID, shipmentID uuid.UUID) (*models.Shipment, error) {
	if shipmentID == uuid.Nil {
		return nil, fmt.Errorf("invalid shipment ID: cannot be nil")
	}

	shipment, err := s.repo.GetShipmentByID(ctx, userID, shipmentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("shipment not found: %s", shipmentID)
		}
		return nil, fmt.Errorf("failed to retrieve shipment: %w", err)
	}

	// Validate that shipment has required fields for API sync
	if shipment.ShipmentNumber == "" {
		return nil, fmt.Errorf("shipment %s has no shipment number", shipmentID)
	}
	if shipment.ShipmentType == "" {
		return nil, fmt.Errorf("shipment %s has no shipment type", shipmentID)
	}
	if shipment.SealineCode == "" {
		return nil, fmt.Errorf("shipment %s has no sealine code", shipmentID)
	}

	return shipment, nil
}

// // SyncStats holds statistics about the sync operation
// type SyncStats struct {
// 	LocationsCreated       int
// 	RoutesCreated          int
// 	VesselsCreated         int
// 	FacilitiesCreated      int
// 	ContainersCreated      int
// 	ContainerEventsCreated int
// 	RouteSegmentsCreated   int
// 	CoordinatesCreated     int
// 	AisRecordsCreated      int
// }

// recreateShipmentRelatedData recreates all shipment related data from API response
func (s *shipmentService) recreateShipmentRelatedData(ctx context.Context, shipment *models.Shipment, apiResponse *shipmentsDto.SafeCubeAPIShipmentResponse) (*types.SyncStats, error) {
	// Validate API response
	if apiResponse == nil {
		return nil, fmt.Errorf("API response is nil")
	}
	if shipment == nil {
		return nil, fmt.Errorf("shipment is nil")
	}

	stats := &types.SyncStats{}
	log.Printf("Creating %d locations for shipment %s", len(apiResponse.Locations), shipment.ShipmentNumber)
	// Create locations
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
		_, err := s.repo.CreateLocation(ctx, &shipment.ID, &location)
		if err != nil {
			return nil, fmt.Errorf("failed to create location: %w", err)
		}
		stats.LocationsCreated++
	}

	log.Printf("Creating route points for shipment %s", shipment.ShipmentNumber)
	// Create routes
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
			return nil, fmt.Errorf("failed to create route location: %w", err)
		}

		route := &models.ShipmentRoute{
			ShipmentID:    shipment.ID,
			LocationID:    loc.ID,
			RouteType:     routeType,
			Date:          r.Date,
			Actual:        r.Actual,
			PredictiveETA: r.PredictiveEta,
		}

		_, err = s.repo.CreateRoute(ctx, route)
		if err != nil {
			return nil, fmt.Errorf("failed to create route: %w", err)
		}
		stats.RoutesCreated++
	}

	log.Printf("Creating %d vessels for shipment %s", len(apiResponse.Vessels), shipment.ShipmentNumber)
	// Create vessels
	for _, v := range apiResponse.Vessels {
		vessel := models.Vessel{
			Name:     v.Name,
			Imo:      v.Imo,
			Mmsi:     v.Mmsi,
			CallSign: v.CallSign,
			Flag:     v.Flag,
		}
		_, err := s.repo.CreateVessel(ctx, &shipment.ID, &vessel)
		if err != nil {
			return nil, fmt.Errorf("failed to create vessel: %w", err)
		}
		stats.VesselsCreated++
	}

	log.Printf("Creating %d facilities for shipment %s", len(apiResponse.Facilities), shipment.ShipmentNumber)
	// Create facilities
	for _, f := range apiResponse.Facilities {
		facility := models.Facility{
			Name:        f.Name,
			CountryCode: f.CountryCode,
			Locode:      f.Locode,
			BicCode:     f.BicCode,
			SmdgCode:    f.SmdgCode,
			Latitude:    &f.Coordinates.Lat,
			Longitude:   &f.Coordinates.Lng,
		}
		_, err := s.repo.CreateFacility(ctx, &shipment.ID, &facility)
		if err != nil {
			return nil, fmt.Errorf("failed to create facility: %w", err)
		}
		stats.FacilitiesCreated++
	}

	log.Printf("Creating %d containers for shipment %s", len(apiResponse.Containers), shipment.ShipmentNumber)
	// Create containers and their events
	for _, c := range apiResponse.Containers {
		container := models.Container{
			Number:   c.Number,
			IsoCode:  c.IsoCode,
			SizeType: c.SizeType,
			Status:   c.Status,
		}
		createdContainer, err := s.repo.CreateContainer(ctx, &shipment.ID, &container)
		if err != nil {
			return nil, fmt.Errorf("failed to create container: %w", err)
		}
		stats.ContainersCreated++

		// Create container events
		for _, ce := range c.Events {
			location, err := s.repo.FindLocationByLocode(ctx, ce.Location.Locode)
			if err != nil {
				return nil, fmt.Errorf("failed to find location for container event: %w", err)
			}

			var facility *models.Facility
			if ce.Facility != nil {
				facility, err = s.repo.FindFacilityByLocode(ctx, ce.Facility.Locode)
				if err != nil {
					return nil, fmt.Errorf("failed to find facility for container event: %w", err)
				}
			}

			var vessel *models.Vessel
			if ce.Vessel != nil {
				vessel, err = s.repo.FindVesselByIMOAndMMSI(ctx, ce.Vessel.Imo, ce.Vessel.Mmsi)
				if err != nil {
					if !errors.Is(err, gorm.ErrRecordNotFound) {
						return nil, fmt.Errorf("failed to find vessel for container event: %w", err)
					}
				}
			}

			containerEvent := models.ContainerEvent{
				ContainerID:       createdContainer.ID,
				LocationID:        location.ID,
				Description:       ce.Description,
				EventType:         ce.EventType,
				EventCode:         ce.EventCode,
				Status:            ce.Status,
				Date:              ce.Date,
				IsActual:          ce.IsActual,
				IsAdditionalEvent: ce.IsAdditionalEvent,
				RouteType:         ce.RouteType,
				TransportType:     ce.TransportType,
				Voyage:            ce.Voyage,
			}

			if vessel != nil {
				containerEvent.VesselID = &vessel.ID
			}

			if facility != nil {
				containerEvent.FacilityID = &facility.ID
			}

			_, err = s.repo.CreateContainerEvent(ctx, &containerEvent)
			if err != nil {
				return nil, fmt.Errorf("failed to create container event: %w", err)
			}
			stats.ContainerEventsCreated++
		}
	}

	log.Printf("Creating %d route segments for shipment %s", len(apiResponse.RouteData.RouteSegments), shipment.ShipmentNumber)
	// Create route segments
	for segIdx, rs := range apiResponse.RouteData.RouteSegments {
		segment := models.RouteSegment{
			ShipmentID:   shipment.ID,
			RouteType:    rs.RouteType,
			SegmentOrder: segIdx,
		}
		createdSegment, err := s.repo.CreateRouteSegment(ctx, &segment)
		if err != nil {
			return nil, fmt.Errorf("failed to create route segment: %w", err)
		}
		stats.RouteSegmentsCreated++

		// Create segment points
		for pointIdx, point := range rs.Path {
			segmentPoint := models.RouteSegmentPoint{
				SegmentID:  createdSegment.ID,
				Latitude:   point.Lat,
				Longitude:  point.Lng,
				PointOrder: pointIdx,
			}
			_, err := s.repo.CreateRouteSegmentPoint(ctx, &segmentPoint)
			if err != nil {
				return nil, fmt.Errorf("failed to create route segment point: %w", err)
			}
		}
	}

	log.Printf("Creating coordinates for shipment %s", shipment.ShipmentNumber)
	// Create coordinates
	coordinate := &models.Coordinate{
		ShipmentID: shipment.ID,
		Latitude:   apiResponse.RouteData.Coordinates.Lat,
		Longitude:  apiResponse.RouteData.Coordinates.Lng,
	}

	_, err := s.repo.CreateCoordinate(ctx, coordinate)
	if err != nil {
		return nil, fmt.Errorf("failed to create coordinate: %w", err)
	}
	stats.CoordinatesCreated++

	log.Printf("Creating AIS data for shipment %s", shipment.ShipmentNumber)
	// Create AIS data
	if apiResponse.RouteData.Ais.Data != nil {
		aisData := apiResponse.RouteData.Ais.Data
		aisVessel, err := s.repo.FindVesselByIMOAndMMSI(ctx, aisData.Vessel.Imo, aisData.Vessel.Mmsi)
		if err != nil {
			return nil, fmt.Errorf("failed to find AIS vessel: %w", err)
		}

		ais := &models.Ais{
			ShipmentID:               shipment.ID,
			Status:                   apiResponse.RouteData.Ais.Status,
			LastEventDescription:     aisData.LastEvent.Description,
			LastEventDate:            aisData.LastEvent.Date,
			LastEventVoyage:          aisData.LastEvent.Voyage,
			DischargePortName:        aisData.DischargePort.Name,
			DischargePortCountryCode: aisData.DischargePort.CountryCode,
			DischargePortCode:        aisData.DischargePort.Code,
			DischargePortDate:        aisData.DischargePort.Date,
			DischargePortDateLabel:   aisData.DischargePort.DateLabel,
			DeparturePortName:        aisData.DeparturePort.Name,
			DeparturePortCountryCode: aisData.DeparturePort.CountryCode,
			DeparturePortCode:        aisData.DeparturePort.Code,
			DeparturePortDate:        aisData.DeparturePort.Date,
			DeparturePortDateLabel:   aisData.DeparturePort.DateLabel,
			ArrivalPortName:          aisData.ArrivalPort.Name,
			ArrivalPortCountryCode:   aisData.ArrivalPort.CountryCode,
			ArrivalPortCode:          aisData.ArrivalPort.Code,
			ArrivalPortDate:          aisData.ArrivalPort.Date,
			ArrivalPortDateLabel:     aisData.ArrivalPort.DateLabel,
			VesselID:                 &aisVessel.ID,
			LastVesselPositionLat:    aisData.LastVesselPosition.Lat,
			LastVesselPositionLng:    aisData.LastVesselPosition.Lng,
			LastVesselPositionUpdate: aisData.LastVesselPosition.UpdatedAt,
		}
		_, err = s.repo.CreateAis(ctx, ais)
		if err != nil {
			return nil, fmt.Errorf("failed to create AIS data: %w", err)
		}
		stats.AisRecordsCreated++
	} else {
		ais := &models.Ais{
			ShipmentID: shipment.ID,
			Status:     apiResponse.RouteData.Ais.Status,
		}
		_, err := s.repo.CreateAis(ctx, ais)
		if err != nil {
			return nil, err
		}
	}
	log.Printf("Sync statistics for shipment %s: %+v", shipment.ShipmentNumber, stats)
	return stats, nil
}

func (s *shipmentService) SyncShipment(ctx context.Context, userID, shipmentID uuid.UUID) (*models.Shipment, error) {
	// Validate shipment before starting sync
	existingShipment, err := s.validateShipmentForSync(ctx, userID, shipmentID)
	if err != nil {
		return nil, err
	}

	log.Printf("Starting sync for shipment %s (ID: %s)", existingShipment.ShipmentNumber, shipmentID)

	// Get data summary before sync
	beforeSummary, err := s.repo.GetShipmentDataSummary(ctx, shipmentID)
	if err != nil {
		log.Printf("Warning: Failed to get before-sync summary for shipment %s: %v", shipmentID, err)
	} else {
		log.Printf("Before sync - Shipment %s data counts: locations=%d, routes=%d, vessels=%d, facilities=%d, containers=%d, events=%d, segments=%d, coordinates=%d, ais=%d",
			existingShipment.ShipmentNumber, beforeSummary.LocationsCount, beforeSummary.RoutesCount, beforeSummary.VesselsCount,
			beforeSummary.FacilitiesCount, beforeSummary.ContainersCount, beforeSummary.ContainerEventsCount,
			beforeSummary.RouteSegmentsCount, beforeSummary.CoordinatesCount, beforeSummary.AisCount)
	}
	// Get fresh data from SafeCube API
	apiResponse, err := s.safeCubeAPIService.GetShipmentDetails(
		ctx,
		existingShipment.ShipmentNumber,
		existingShipment.ShipmentType,
		existingShipment.SealineCode,
	)
	if err != nil {
		return nil, err
	}

	log.Printf("Starting sync for shipment %s (ID: %s)", existingShipment.ShipmentNumber, shipmentID)

	// Start a transaction to ensure data consistency
	tx := s.repo.GetDB().DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic occurred during shipment sync, rolling back transaction: %v", r)
			tx.Rollback()
			panic(r)
		}
	}()

	// Create a context with the transaction
	txCtx := context.WithValue(ctx, "tx", tx)

	log.Printf("Cleaning up existing data for shipment %s", existingShipment.ShipmentNumber)
	// Delete all existing related data
	err = s.repo.DeleteAllShipmentRelatedData(txCtx, shipmentID)
	if err != nil {
		log.Printf("Failed to delete existing data for shipment %s: %v", existingShipment.ShipmentNumber, err)
		tx.Rollback()
		return nil, fmt.Errorf("failed to delete existing shipment data: %w", err)
	}

	// Update shipment metadata
	shipmentModel := &models.Shipment{
		ShipmentNumber: apiResponse.Metadata.ShipmentNumber,
		ShipmentType:   apiResponse.Metadata.ShipmentType,
		SealineCode:    apiResponse.Metadata.Sealine,
		SealineName:    apiResponse.Metadata.SealineName,
		ShippingStatus: apiResponse.Metadata.ShippingStatus,
		Warnings:       apiResponse.Metadata.Warnings,
	}

	log.Printf("Updating shipment metadata for %s", existingShipment.ShipmentNumber)
	// Use transaction context for update
	err = tx.Model(&models.Shipment{}).Where("id = ?", shipmentID).Updates(shipmentModel).Error
	if err != nil {
		log.Printf("Failed to update shipment metadata for %s: %v", existingShipment.ShipmentNumber, err)
		tx.Rollback()
		return nil, fmt.Errorf("failed to update shipment: %w", err)
	}

	// Get the updated shipment
	var shipment models.Shipment
	err = tx.First(&shipment, "id = ?", shipmentID).Error
	if err != nil {
		log.Printf("Failed to retrieve updated shipment %s: %v", existingShipment.ShipmentNumber, err)
		tx.Rollback()
		return nil, fmt.Errorf("failed to get updated shipment: %w", err)
	}

	log.Printf("Recreating related data for shipment %s", shipment.ShipmentNumber)
	// Recreate all related data from fresh API response
	stats, err := s.recreateShipmentRelatedData(txCtx, &shipment, apiResponse)
	if err != nil {
		log.Printf("Failed to recreate related data for shipment %s: %v", shipment.ShipmentNumber, err)
		tx.Rollback()
		return nil, fmt.Errorf("failed to recreate shipment data: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Failed to commit transaction for shipment %s: %v", shipment.ShipmentNumber, err)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Get data summary after sync
	afterSummary, err := s.repo.GetShipmentDataSummary(ctx, shipmentID)
	if err != nil {
		log.Printf("Warning: Failed to get after-sync summary for shipment %s: %v", shipmentID, err)
	} else {
		log.Printf("After sync - Shipment %s data counts: locations=%d, routes=%d, vessels=%d, facilities=%d, containers=%d, events=%d, segments=%d, coordinates=%d, ais=%d",
			shipment.ShipmentNumber, afterSummary.LocationsCount, afterSummary.RoutesCount, afterSummary.VesselsCount,
			afterSummary.FacilitiesCount, afterSummary.ContainersCount, afterSummary.ContainerEventsCount,
			afterSummary.RouteSegmentsCount, afterSummary.CoordinatesCount, afterSummary.AisCount)
	}

	log.Printf("Successfully synced shipment %s (ID: %s) with fresh data from SafeCube API. Created: %d locations, %d routes, %d vessels, %d facilities, %d containers, %d events, %d segments, %d coordinates, %d AIS records",
		shipment.ShipmentNumber, shipment.ID,
		stats.LocationsCreated, stats.RoutesCreated, stats.VesselsCreated, stats.FacilitiesCreated,
		stats.ContainersCreated, stats.ContainerEventsCreated, stats.RouteSegmentsCreated,
		stats.CoordinatesCreated, stats.AisRecordsCreated)

	return &shipment, nil
}

func (s *shipmentService) RefreshShipment(ctx context.Context, userID, shipmentID uuid.UUID) (*models.Shipment, error) {
	log.Printf("User %s requesting refresh for shipment %s", userID, shipmentID)

	owns, err := s.repo.CheckUserOwnsShipment(ctx, userID, shipmentID)
	if err != nil {
		log.Printf("Error checking ownership for shipment %s by user %s: %v", shipmentID, userID, err)
		return nil, err
	}
	if !owns {
		log.Printf("User %s denied access to shipment %s (not owned)", userID, shipmentID)
		return nil, fmt.Errorf("shipment not found or access denied")
	}

	log.Printf("User %s authorized to refresh shipment %s", userID, shipmentID)
	shipment, err := s.SyncShipment(ctx, userID, shipmentID)
	if err != nil {
		log.Printf("Failed to sync shipment %s for user %s: %v", shipmentID, userID, err)
		return nil, err
	}

	log.Printf("Successfully refreshed shipment %s for user %s", shipmentID, userID)
	return shipment, nil
}

// SystemRefreshShipment refreshes a shipment without user authentication (for background jobs)
func (s *shipmentService) SystemRefreshShipment(ctx context.Context, shipmentID uuid.UUID) (*models.Shipment, error) {
	log.Printf("System refresh requested for shipment %s", shipmentID)

	shipment, err := s.SystemSyncShipment(ctx, shipmentID)
	if err != nil {
		log.Printf("Failed to system refresh shipment %s: %v", shipmentID, err)
		return nil, err
	}

	log.Printf("Successfully system refreshed shipment %s", shipmentID)
	return shipment, nil
}

// SystemSyncShipment syncs a shipment without user ownership validation (for background jobs)
func (s *shipmentService) SystemSyncShipment(ctx context.Context, shipmentID uuid.UUID) (*models.Shipment, error) {
	// Get shipment to verify it exists and check its status
	var existingShipment models.Shipment
	err := s.repo.GetDB().DB.WithContext(ctx).First(&existingShipment, "id = ?", shipmentID).Error
	if err != nil {
		return nil, fmt.Errorf("shipment not found: %w", err)
	}

	// Skip delivered shipments - no need to refresh completed shipments
	if existingShipment.ShippingStatus == "delivered" {
		log.Printf("Skipping delivered shipment %s", existingShipment.ShipmentNumber)
		return &existingShipment, nil
	}

	log.Printf("Starting system sync for shipment %s (ID: %s)", existingShipment.ShipmentNumber, shipmentID)

	// Get data summary before sync
	beforeSummary, err := s.repo.GetShipmentDataSummary(ctx, shipmentID)
	if err != nil {
		log.Printf("Warning: Failed to get before-sync summary for shipment %s: %v", shipmentID, err)
	} else {
		log.Printf("Before sync - Shipment %s data counts: locations=%d, routes=%d, vessels=%d, facilities=%d, containers=%d, events=%d, segments=%d, coordinates=%d, ais=%d",
			existingShipment.ShipmentNumber, beforeSummary.LocationsCount, beforeSummary.RoutesCount, beforeSummary.VesselsCount,
			beforeSummary.FacilitiesCount, beforeSummary.ContainersCount, beforeSummary.ContainerEventsCount,
			beforeSummary.RouteSegmentsCount, beforeSummary.CoordinatesCount, beforeSummary.AisCount)
	}

	// Get fresh data from SafeCube API
	apiResponse, err := s.safeCubeAPIService.GetShipmentDetails(
		ctx,
		existingShipment.ShipmentNumber,
		existingShipment.ShipmentType,
		existingShipment.SealineCode,
	)
	if err != nil {
		return nil, err
	}

	// Start a transaction to ensure data consistency
	tx := s.repo.GetDB().DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic occurred during system shipment sync, rolling back transaction: %v", r)
			tx.Rollback()
			panic(r)
		}
	}()

	// Create a context with the transaction
	txCtx := context.WithValue(ctx, "tx", tx)

	log.Printf("Cleaning up existing data for shipment %s", existingShipment.ShipmentNumber)
	// Delete all existing related data
	err = s.repo.DeleteAllShipmentRelatedData(txCtx, shipmentID)
	if err != nil {
		log.Printf("Failed to delete existing data for shipment %s: %v", existingShipment.ShipmentNumber, err)
		tx.Rollback()
		return nil, fmt.Errorf("failed to delete existing shipment data: %w", err)
	}

	// Update shipment metadata
	shipmentModel := &models.Shipment{
		ShipmentNumber: apiResponse.Metadata.ShipmentNumber,
		ShipmentType:   apiResponse.Metadata.ShipmentType,
		SealineCode:    apiResponse.Metadata.Sealine,
		SealineName:    apiResponse.Metadata.SealineName,
		ShippingStatus: apiResponse.Metadata.ShippingStatus,
		Warnings:       apiResponse.Metadata.Warnings,
	}

	log.Printf("Updating shipment metadata for %s", existingShipment.ShipmentNumber)
	// Use transaction context for update
	err = tx.Model(&models.Shipment{}).Where("id = ?", shipmentID).Updates(shipmentModel).Error
	if err != nil {
		log.Printf("Failed to update shipment metadata for %s: %v", existingShipment.ShipmentNumber, err)
		tx.Rollback()
		return nil, fmt.Errorf("failed to update shipment: %w", err)
	}

	// Get the updated shipment
	var shipment models.Shipment
	err = tx.First(&shipment, "id = ?", shipmentID).Error
	if err != nil {
		log.Printf("Failed to retrieve updated shipment %s: %v", existingShipment.ShipmentNumber, err)
		tx.Rollback()
		return nil, fmt.Errorf("failed to get updated shipment: %w", err)
	}

	log.Printf("Recreating related data for shipment %s", shipment.ShipmentNumber)
	// Recreate all related data from fresh API response
	stats, err := s.recreateShipmentRelatedData(txCtx, &shipment, apiResponse)
	if err != nil {
		log.Printf("Failed to recreate related data for shipment %s: %v", shipment.ShipmentNumber, err)
		tx.Rollback()
		return nil, fmt.Errorf("failed to recreate shipment data: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Failed to commit transaction for shipment %s: %v", shipment.ShipmentNumber, err)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Get data summary after sync
	afterSummary, err := s.repo.GetShipmentDataSummary(ctx, shipmentID)
	if err != nil {
		log.Printf("Warning: Failed to get after-sync summary for shipment %s: %v", shipmentID, err)
	} else {
		log.Printf("After sync - Shipment %s data counts: locations=%d, routes=%d, vessels=%d, facilities=%d, containers=%d, events=%d, segments=%d, coordinates=%d, ais=%d",
			shipment.ShipmentNumber, afterSummary.LocationsCount, afterSummary.RoutesCount, afterSummary.VesselsCount,
			afterSummary.FacilitiesCount, afterSummary.ContainersCount, afterSummary.ContainerEventsCount,
			afterSummary.RouteSegmentsCount, afterSummary.CoordinatesCount, afterSummary.AisCount)
	}

	log.Printf("Successfully system synced shipment %s (ID: %s) with fresh data from SafeCube API. Created: %d locations, %d routes, %d vessels, %d facilities, %d containers, %d events, %d segments, %d coordinates, %d AIS records",
		shipment.ShipmentNumber, shipment.ID,
		stats.LocationsCreated, stats.RoutesCreated, stats.VesselsCreated, stats.FacilitiesCreated,
		stats.ContainersCreated, stats.ContainerEventsCreated, stats.RouteSegmentsCreated,
		stats.CoordinatesCreated, stats.AisRecordsCreated)

	return &shipment, nil
}

func (s *shipmentService) GetShipmentDetails(ctx context.Context, userID, shipmentID uuid.UUID) (*dto.ShipmentDetailsResponse, error) {
	owns, err := s.repo.CheckUserOwnsShipment(ctx, userID, shipmentID)
	if err != nil {
		return nil, err
	}
	if !owns {
		return nil, fmt.Errorf("shipment not found or access denied")
	}

	shipmentDetails, err := s.repo.GetShipmentDetails(ctx, userID, shipmentID)
	if err != nil {
		return nil, err
	}

	return shipmentDetails, nil
}

func (s *shipmentService) DeleteUserShipment(ctx context.Context, userID, shipmentID uuid.UUID) error {
	err := s.repo.DeleteUserShipment(ctx, userID, shipmentID)
	if err != nil {
		return err
	}
	return nil
}

func (s *shipmentService) BulkDeleteUserShipments(ctx context.Context, userID uuid.UUID, shipmentIDs []uuid.UUID) error {
	err := s.repo.BulkDeleteUserShipments(ctx, userID, shipmentIDs)
	if err != nil {
		return err
	}
	return nil
}
