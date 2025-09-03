package dto

import "github.com/google/uuid"

type GridDataResponse struct {
	Rows []GridShipment `json:"rows"`
}

type GridShipment struct {
	ID             uuid.UUID `json:"id"`
	ShipmentNumber string    `json:"shipmentNumber"`
	ShipmentType   string    `json:"shipmentType"`
	SealineCode    string    `json:"sealineCode"`
	SealineName    string    `json:"sealineName"`
	ShippingStatus string    `json:"shippingStatus"`
	CreatedAt      string    `json:"createdAt"`
	UpdatedAt      string    `json:"updatedAt"`
}
