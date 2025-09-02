package dto

import "github.com/google/uuid"

type GridDataResponse struct {
	Rows    []GridShipment `json:"rows"`
	LastRow int            `json:"lastRow"`
}

type GridShipment struct {
	ID             uuid.UUID `json:"id"`
	ShipmentNumber string    `json:"shipment_number"`
	ShipmentType   string    `json:"shipment_type"`
	SealineCode    string    `json:"sealine_code"`
	SealineName    string    `json:"sealine_name"`
	ShippingStatus string    `json:"shipping_status"`
	CreatedAt      string    `json:"created_at"`
	UpdatedAt      string    `json:"updated_at"`
}
