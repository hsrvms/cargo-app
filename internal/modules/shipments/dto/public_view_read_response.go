package dto

import (
	"time"

	"github.com/google/uuid"
)

type PublicViewReadResponse struct {
	ID        uuid.UUID                 `json:"id"`
	Name      string                    `json:"name"`
	ExpiresAt time.Time                 `json:"expiresAt"`
	Shipments []ShipmentDetailsResponse `json:"shipments"`
}

type PublicShipmentResponse struct {
	ID             uuid.UUID  `json:"id"`
	ShipmentType   string     `json:"shipmentType"`
	ShipmentNumber string     `json:"shipmentNumber"`
	ShipmentDate   *time.Time `json:"shipmentDate"`
	SealineCode    string     `json:"sealineCode"`
	SealineName    string     `json:"sealineName"`
	ShippingStatus string     `json:"shippingStatus"`

	Locations []ShipmentLocationResponse `json:"locations"`
	// Route      ShipmentRouteResponse       `json:"route"`
	// Vessels    []ShipmentVesselResponse    `json:"vessels"`
	// Facilities []ShipmentFacilityResponse  `json:"facilities"`
	// Containers []ShipmentContainerResponse `json:"containers"`
	// RouteData  ShipmentRouteDataResponse   `json:"routeData"`
}

// func ToPublicViewResponse(view *models.PublicView) PublicViewReadResponse {
// 	shipmentDTOs := make([]*PublicShipmentResponse, 0, len(view.Shipments))

// 	for _, s := range view.Shipments{

// 	}

// }
