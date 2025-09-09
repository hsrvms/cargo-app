package dto

import (
	"time"

	"github.com/google/uuid"
)

type ShipmentDetailsResponse struct {
	ID             uuid.UUID                   `json:"id"`
	ShipmentType   string                      `json:"shipmentType"`
	ShipmentNumber string                      `json:"shipmentNumber"`
	SealineCode    string                      `json:"sealineCode"`
	SealineName    string                      `json:"sealineName"`
	ShippingStatus string                      `json:"shippingStatus"`
	CreatedAt      time.Time                   `json:"createdAt"`
	UpdatedAt      time.Time                   `json:"updatedAt"`
	Recipient      string                      `json:"recipient"`
	Address        string                      `json:"address"`
	Notes          string                      `json:"notes"`
	Locations      []ShipmentLocationResponse  `json:"locations"`
	Route          ShipmentRouteResponse       `json:"route"`
	Vessels        []ShipmentVesselResponse    `json:"vessels"`
	Facilities     []ShipmentFacilityResponse  `json:"facilities"`
	Containers     []ShipmentContainerResponse `json:"containers"`
	RouteData      ShipmentRouteDataResponse   `json:"routeData"`
}

type ShipmentLocationResponse struct {
	Name        string  `json:"name"`
	State       *string `json:"state"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Locode      string  `json:"locode"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
}

type ShipmentRouteResponse struct {
	Prepol  *ShipmentRoutePoint `json:"prepol,omitempty"`
	Pol     *ShipmentRoutePoint `json:"pol,omitempty"`
	Pod     *ShipmentRoutePoint `json:"pod,omitempty"`
	Postpod *ShipmentRoutePoint `json:"postpod,omitempty"`
}

type ShipmentRoutePoint struct {
	Location      ShipmentLocationResponse `json:"location"`
	Date          *time.Time               `json:"date"`
	Actual        *bool                    `json:"actual"`
	PredictiveETA *time.Time               `json:"predictiveEta"`
}

type ShipmentVesselResponse struct {
	Name     string `json:"name"`
	Imo      int    `json:"imo"`
	Mmsi     int    `json:"mmsi"`
	CallSign string `json:"callSign"`
	Flag     string `json:"flag"`
}

type ShipmentFacilityResponse struct {
	Name        string  `json:"name"`
	CountryCode string  `json:"countryCode"`
	Locode      string  `json:"locode"`
	BicCode     *string `json:"bicCode"`
	SmdgCode    *string `json:"smdgCode"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

type ShipmentContainerResponse struct {
	Number   string                           `json:"number"`
	IsoCode  string                           `json:"isoCode"`
	SizeType string                           `json:"sizeType"`
	Status   string                           `json:"status"`
	Events   []ShipmentContainerEventResponse `json:"events"`
}

type ShipmentContainerEventResponse struct {
	Location          ShipmentLocationResponse  `json:"location"`
	Facility          *ShipmentFacilityResponse `json:"facility"`
	Description       string                    `json:"description"`
	EventType         *string                   `json:"eventType"`
	EventCode         *string                   `json:"eventCode"`
	Status            string                    `json:"status"`
	Date              time.Time                 `json:"date"`
	IsActual          bool                      `json:"isActual"`
	IsAdditionalEvent bool                      `json:"isAdditionalEvent"`
	RouteType         string                    `json:"routeType"`
	TransportType     *string                   `json:"transportType"`
	Vessel            *ShipmentVesselResponse   `json:"vessel"`
	Voyage            *string                   `json:"voyage"`
}

type ShipmentRouteDataResponse struct {
	RouteSegments []ShipmentRouteSegmentResponse `json:"routeSegments"`
	Coordinates   ShipmentCoordinatesResponse    `json:"coordinates"`
	Ais           ShipmentAisResponse            `json:"ais"`
}

type ShipmentRouteSegmentResponse struct {
	Path         []ShipmentRouteSegmentPointResponse `json:"path"`
	RouteType    string                              `json:"routeType"`
	SegmentOrder int                                 `json:"segmentOrder"`
}

type ShipmentRouteSegmentPointResponse struct {
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	PointOrder int     `json:"pointOrder"`
}

type ShipmentCoordinatesResponse struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ShipmentAisResponse struct {
	Status                   string                  `json:"status"`
	LastEventDescription     *string                 `json:"lastEventDescription"`
	LastEventDate            *time.Time              `json:"lastEventDate"`
	LastEventVoyage          *string                 `json:"lastEventVoyage"`
	DischargePortName        *string                 `json:"dischargePortName"`
	DischargePortCountryCode *string                 `json:"dischargePortCountryCode"`
	DischargePortCode        *string                 `json:"dischargePortCode"`
	DischargePortDate        *time.Time              `json:"dischargePortDate"`
	DischargePortDateLabel   *string                 `json:"dischargePortDateLabel"`
	DeparturePortName        *string                 `json:"departurePortName"`
	DeparturePortCountryCode *string                 `json:"departurePortCountryCode"`
	DeparturePortCode        *string                 `json:"departurePortCode"`
	DeparturePortDate        *time.Time              `json:"departurePortDate"`
	DeparturePortDateLabel   *string                 `json:"departurePortDateLabel"`
	ArrivalPortName          *string                 `json:"arrivalPortName"`
	ArrivalPortCountryCode   *string                 `json:"arrivalPortCountryCode"`
	ArrivalPortCode          *string                 `json:"arrivalPortCode"`
	ArrivalPortDate          *time.Time              `json:"arrivalPortDate"`
	ArrivalPortDateLabel     *string                 `json:"arrivalPortDateLabel"`
	Vessel                   *ShipmentVesselResponse `json:"vessel"`
	LastVesselPositionLat    *float64                `json:"lastVesselPositionLat"`
	LastVesselPositionLng    *float64                `json:"lastVesselPositionLng"`
	LastVesselPositionUpdate *time.Time              `json:"lastVesselPositionUpdate"`
	UpdatedAt                time.Time               `json:"updatedAt"`
}
