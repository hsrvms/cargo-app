package handlers

import (
	"fmt"
	"go-starter/internal/modules/shipments/dto"
	"go-starter/internal/modules/shipments/models"
	"time"

	"github.com/google/uuid"
)

func PublicViewToReadResponse(view *models.PublicView) dto.PublicViewReadResponse {
	return dto.PublicViewReadResponse{
		ID:        view.ID,
		Name:      view.Name,
		ExpiresAt: view.ExpiresAt,
		Shipments: ShipmentsToDetailsResponse(view.Shipments),
	}
}

func ShipmentsToDetailsResponse(shipments []models.Shipment) []dto.ShipmentDetailsResponse {
	if len(shipments) == 0 {
		return []dto.ShipmentDetailsResponse{}
	}

	out := make([]dto.ShipmentDetailsResponse, 0, len(shipments))
	for _, s := range shipments {
		out = append(out, ShipmentToDetailsResponse(s))
	}
	return out
}

func ShipmentToDetailsResponse(s models.Shipment) dto.ShipmentDetailsResponse {

	containersDTO := shipmentContainersToDTO(s.Containers)

	response := dto.ShipmentDetailsResponse{
		ID:             s.ID,
		ShipmentType:   s.ShipmentType,
		ShipmentNumber: s.ShipmentNumber,
		SealineCode:    s.SealineCode,
		SealineName:    s.SealineName,
		ShippingStatus: s.ShippingStatus,

		Locations:  shipmentLocationsToDTO(s.Locations),
		Route:      shipmentRouteToDTO(s.Routes),
		Vessels:    shipmentVesselsToDTO(s.Vessels),
		Facilities: shipmentFacilitiesToDTO(s.Facilities),
		Containers: containersDTO,
		RouteData:  shipmentRouteDataToDTO(s.RouteSegments, s.Coordinates, *s.Ais),
	}

	if date, err := GetShipmentDate(containersDTO); err == nil {
		response.ShipmentDate = date
	}

	return response
}

func GetShipmentDate(containers []dto.ShipmentContainerResponse) (*time.Time, error) {
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

	return nil, fmt.Errorf("no event found in shipment data")
}

func shipmentLocationsToDTO(locations []models.Location) []dto.ShipmentLocationResponse {
	if len(locations) == 0 {
		return []dto.ShipmentLocationResponse{}
	}

	out := make([]dto.ShipmentLocationResponse, 0, len(locations))
	for _, l := range locations {
		out = append(out, shipmentLocationToDTO(l))
	}

	return out
}

func shipmentLocationToDTO(l models.Location) dto.ShipmentLocationResponse {
	return dto.ShipmentLocationResponse{
		Name:        l.Name,
		State:       l.State,
		Country:     l.Country,
		CountryCode: l.CountryCode,
		Locode:      l.Locode,
		Latitude:    l.Latitude,
		Longitude:   l.Longitude,
		Timezone:    l.Timezone,
	}
}

func shipmentRouteToDTO(routes []models.ShipmentRoute) dto.ShipmentRouteResponse {
	if len(routes) == 0 {
		return dto.ShipmentRouteResponse{}
	}

	out := dto.ShipmentRouteResponse{}
	for _, route := range routes {
		item := &dto.ShipmentRoutePoint{
			Location:      shipmentLocationToDTO(route.Location),
			Date:          route.Date,
			Actual:        route.Actual,
			PredictiveETA: route.PredictiveETA,
		}
		switch route.RouteType {
		case "PREPOL":
			out.Prepol = item
		case "POL":
			out.Pol = item
		case "POD":
			out.Pod = item
		case "POSTPOD":
			out.Postpod = item
		}
	}
	return out
}

func shipmentVesselsToDTO(vessels []models.Vessel) []dto.ShipmentVesselResponse {
	if len(vessels) == 0 {
		return []dto.ShipmentVesselResponse{}
	}

	out := make([]dto.ShipmentVesselResponse, 0, len(vessels))
	for _, v := range vessels {
		out = append(out, shipmentVesselToDTO(v))
	}

	return out
}

func shipmentVesselToDTO(v models.Vessel) dto.ShipmentVesselResponse {
	return dto.ShipmentVesselResponse{
		Name:     v.Name,
		Imo:      v.Imo,
		Mmsi:     v.Mmsi,
		CallSign: v.CallSign,
		Flag:     v.Flag,
	}
}

func shipmentFacilitiesToDTO(facilities []models.Facility) []dto.ShipmentFacilityResponse {
	if len(facilities) == 0 {
		return []dto.ShipmentFacilityResponse{}
	}

	out := make([]dto.ShipmentFacilityResponse, 0, len(facilities))
	for _, f := range facilities {
		out = append(out, shipmentFacilityToDTO(f))
	}

	return out
}

func shipmentFacilityToDTO(f models.Facility) dto.ShipmentFacilityResponse {
	return dto.ShipmentFacilityResponse{
		Name:        f.Name,
		CountryCode: f.CountryCode,
		Locode:      f.Locode,
		BicCode:     f.BicCode,
		SmdgCode:    f.SmdgCode,
		Latitude:    *f.Latitude,
		Longitude:   *f.Longitude,
	}
}

func shipmentContainersToDTO(containers []models.Container) []dto.ShipmentContainerResponse {
	if len(containers) == 0 {
		return []dto.ShipmentContainerResponse{}
	}

	out := make([]dto.ShipmentContainerResponse, 0, len(containers))
	for _, c := range containers {
		out = append(out, shipmentContainerToDTO(c))
	}

	return out
}

func shipmentContainerToDTO(c models.Container) dto.ShipmentContainerResponse {
	return dto.ShipmentContainerResponse{
		Number:   c.Number,
		IsoCode:  c.IsoCode,
		SizeType: c.SizeType,
		Status:   c.Status,
		Events:   shipmentContainerEventsToDTO(c.Events),
	}
}

func shipmentContainerEventsToDTO(events []models.ContainerEvent) []dto.ShipmentContainerEventResponse {
	if len(events) == 0 {
		return []dto.ShipmentContainerEventResponse{}
	}

	out := make([]dto.ShipmentContainerEventResponse, 0, len(events))
	for _, e := range events {
		out = append(out, shipmentContainerEventToDTO(e))
	}

	return out
}

func shipmentContainerEventToDTO(e models.ContainerEvent) dto.ShipmentContainerEventResponse {
	out := dto.ShipmentContainerEventResponse{
		Location:          shipmentLocationToDTO(e.Location),
		Description:       e.Description,
		EventType:         e.EventType,
		EventCode:         e.EventCode,
		Status:            e.Status,
		Date:              e.Date,
		IsActual:          e.IsActual,
		IsAdditionalEvent: e.IsAdditionalEvent,
		RouteType:         e.RouteType,
		TransportType:     e.TransportType,
		Voyage:            e.Voyage,
	}

	if e.FacilityID != nil && e.Facility != nil && e.Facility.ID != uuid.Nil {
		facilityDTO := shipmentFacilityToDTO(*e.Facility)
		out.Facility = &facilityDTO
	}

	if e.VesselID != nil && e.Vessel != nil && e.Vessel.ID != uuid.Nil {
		vesselDTO := shipmentVesselToDTO(*e.Vessel)
		out.Vessel = &vesselDTO
	}

	return out
}

func shipmentRouteDataToDTO(routeSegments []models.RouteSegment, coordinates []models.Coordinate, ais models.Ais) dto.ShipmentRouteDataResponse {
	routeSegDTOs := shipmentRouteSegmentsToDTO(routeSegments)
	coordDTO := shipmentCoordinatesToDTO(coordinates)
	aisDTO := shipmentAisToDTO(&ais)

	return dto.ShipmentRouteDataResponse{
		RouteSegments: routeSegDTOs,
		Coordinates:   coordDTO,
		Ais:           aisDTO,
	}
}

func shipmentAisToDTO(ais *models.Ais) dto.ShipmentAisResponse {
	if ais == nil {
		return dto.ShipmentAisResponse{}
	} else {
		aisDTO := dto.ShipmentAisResponse{
			Status:                   ais.Status,
			LastEventDescription:     ais.LastEventDescription,
			LastEventDate:            ais.LastEventDate,
			LastEventVoyage:          ais.LastEventVoyage,
			DischargePortName:        ais.DischargePortName,
			DischargePortCountryCode: ais.DischargePortCountryCode,
			DischargePortCode:        ais.DischargePortCode,
			DischargePortDate:        ais.DischargePortDate,
			DischargePortDateLabel:   ais.DischargePortDateLabel,
			DeparturePortName:        ais.DeparturePortName,
			DeparturePortCountryCode: ais.DeparturePortCountryCode,
			DeparturePortCode:        ais.DeparturePortCode,
			DeparturePortDate:        ais.DeparturePortDate,
			DeparturePortDateLabel:   ais.DeparturePortDateLabel,
			ArrivalPortName:          ais.ArrivalPortName,
			ArrivalPortCountryCode:   ais.ArrivalPortCountryCode,
			ArrivalPortCode:          ais.ArrivalPortCode,
			ArrivalPortDate:          ais.ArrivalPortDate,
			ArrivalPortDateLabel:     ais.ArrivalPortDateLabel,
			LastVesselPositionLat:    ais.LastVesselPositionLat,
			LastVesselPositionLng:    ais.LastVesselPositionLng,
			LastVesselPositionUpdate: ais.LastVesselPositionUpdate,
			UpdatedAt:                ais.UpdatedAt,
		}

		if ais.Vessel != nil {
			vDTO := shipmentVesselToDTO(*ais.Vessel)
			aisDTO.Vessel = &vDTO
		}
		return aisDTO
	}
}

func shipmentCoordinatesToDTO(coordinates []models.Coordinate) dto.ShipmentCoordinatesResponse {
	if len(coordinates) == 0 {
		return dto.ShipmentCoordinatesResponse{}
	} else {
		latestCoord := coordinates[0]
		for _, c := range coordinates {
			if c.UpdatedAt.After(latestCoord.UpdatedAt) {
				latestCoord = c
			}
		}

		return dto.ShipmentCoordinatesResponse{
			Latitude:  latestCoord.Latitude,
			Longitude: latestCoord.Longitude,
			UpdatedAt: latestCoord.UpdatedAt,
		}
	}
}

func shipmentRouteSegmentsToDTO(routeSegments []models.RouteSegment) []dto.ShipmentRouteSegmentResponse {
	if len(routeSegments) == 0 {
		return []dto.ShipmentRouteSegmentResponse{}
	}

	out := make([]dto.ShipmentRouteSegmentResponse, 0, len(routeSegments))
	for _, rs := range routeSegments {
		out = append(out, shipmentRouteSegmentToDTO(rs))
	}

	return out
}

func shipmentRouteSegmentToDTO(rs models.RouteSegment) dto.ShipmentRouteSegmentResponse {
	return dto.ShipmentRouteSegmentResponse{
		RouteType:    rs.RouteType,
		SegmentOrder: rs.SegmentOrder,
		Path:         shipmentRouteSegmentPointsToDTO(rs.Points),
	}
}

func shipmentRouteSegmentPointsToDTO(points []models.RouteSegmentPoint) []dto.ShipmentRouteSegmentPointResponse {
	if len(points) == 0 {
		return []dto.ShipmentRouteSegmentPointResponse{}
	}

	out := make([]dto.ShipmentRouteSegmentPointResponse, 0, len(points))
	for _, p := range points {
		out = append(out, shipmentRouteSegmentPointToDTO(p))
	}

	return out
}

func shipmentRouteSegmentPointToDTO(p models.RouteSegmentPoint) dto.ShipmentRouteSegmentPointResponse {
	return dto.ShipmentRouteSegmentPointResponse{
		Latitude:   p.Latitude,
		Longitude:  p.Longitude,
		PointOrder: p.PointOrder,
	}
}
