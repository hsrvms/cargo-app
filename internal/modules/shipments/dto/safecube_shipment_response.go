package dto

import (
	"time"
)

type SafeCubeAPIShipmentResponse struct {
	Metadata   SafeCubeMetadata    `json:"metadata"`
	Locations  []SafeCubeLocation  `json:"locations"`
	Route      SafeCubeRoute       `json:"route"`
	Vessels    []SafeCubeVessel    `json:"vessels"`
	Facilities []SafeCubeFacility  `json:"facilities"`
	Containers []SafeCubeContainer `json:"containers"`
	RouteData  SafeCubeRouteData   `json:"routeData"`
}

type SafeCubeMetadata struct {
	ShipmentType   string    `json:"shipmentType"`
	ShipmentNumber string    `json:"shipmentNumber"`
	Sealine        string    `json:"sealine"`
	SealineName    string    `json:"sealineName"`
	ShippingStatus string    `json:"shippingStatus"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Warnings       []string  `json:"warnings"`
}

type SafeCubeLocation struct {
	Name        string              `json:"name"`
	State       *string             `json:"state"`
	Country     string              `json:"country"`
	CountryCode string              `json:"countryCode"`
	Locode      string              `json:"locode"`
	Coordinates SafeCubeCoordinates `json:"coordinates"`
	Timezone    string              `json:"timezone"`
}

type SafeCubeCoordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type SafeCubeRoute struct {
	Prepol  SafeCubeRoutePoint `json:"prepol"`
	Pol     SafeCubeRoutePoint `json:"pol"`
	Pod     SafeCubeRoutePoint `json:"pod"`
	Postpod SafeCubeRoutePoint `json:"postpod"`
}

type SafeCubeRoutePoint struct {
	Location      SafeCubeLocation `json:"location"`
	Date          *time.Time       `json:"date"`
	Actual        *bool            `json:"actual"`
	PredictiveEta *time.Time       `json:"predictiveEta"`
}

type SafeCubeVessel struct {
	Name     string `json:"name"`
	Imo      int    `json:"imo"`
	CallSign string `json:"callSign"`
	Mmsi     int    `json:"mmsi"`
	Flag     string `json:"flag"`
}

type SafeCubeFacility struct {
	Name        string               `json:"name"`
	CountryCode string               `json:"countryCode"`
	Locode      string               `json:"locode"`
	BicCode     *string              `json:"bicCode"`
	SmdgCode    *string              `json:"smdgCode"`
	Coordinates *SafeCubeCoordinates `json:"coordinates"`
}

type SafeCubeContainer struct {
	Number   string          `json:"number"`
	IsoCode  string          `json:"isoCode"`
	SizeType string          `json:"sizeType"`
	Status   string          `json:"status"`
	Events   []SafeCubeEvent `json:"events"`
}

type SafeCubeEvent struct {
	Location          SafeCubeLocation  `json:"location"`
	Facility          *SafeCubeFacility `json:"facility"`
	Description       string            `json:"description"`
	EventType         *string           `json:"eventType"`
	EventCode         *string           `json:"eventCode"`
	Status            string            `json:"status"`
	Date              time.Time         `json:"date"`
	IsActual          bool              `json:"isActual"`
	IsAdditionalEvent bool              `json:"isAdditionalEvent"`
	RouteType         string            `json:"routeType"`
	TransportType     *string           `json:"transportType"`
	Vessel            *SafeCubeVessel   `json:"vessel"`
	Voyage            *string           `json:"voyage"`
}

type SafeCubeRouteData struct {
	RouteSegments []SafeCubeRouteSegment `json:"routeSegments"`
	Coordinates   SafeCubeCoordinates    `json:"coordinates"`
	Ais           SafeCubeAisData        `json:"ais"`
}

type SafeCubeRouteSegment struct {
	Path      []SafeCubeCoordinates `json:"path"`
	RouteType string                `json:"routeType"`
}

type SafeCubeAisData struct {
	Status string             `json:"status"`
	Data   SafeCubeAisDetails `json:"data"`
}

type SafeCubeAisDetails struct {
	LastEvent          SafeCubeLastEvent      `json:"lastEvent"`
	DischargePort      SafeCubePort           `json:"dischargePort"`
	Vessel             SafeCubeVessel         `json:"vessel"`
	LastVesselPosition SafeCubeVesselPosition `json:"lastVesselPosition"`
	DeparturePort      SafeCubePort           `json:"departurePort"`
	ArrivalPort        SafeCubePort           `json:"arrivalPort"`
	UpdatedAt          time.Time              `json:"updatedAt"`
}

type SafeCubeLastEvent struct {
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Voyage      string    `json:"voyage"`
}

type SafeCubePort struct {
	Name        *string   `json:"name,omitempty"`
	CountryCode *string   `json:"countryCode"`
	Code        *string   `json:"code"`
	Date        time.Time `json:"date"`
	DateLabel   string    `json:"dateLabel"`
}

type SafeCubeVesselPosition struct {
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	UpdatedAt time.Time `json:"updatedAt"`
}
