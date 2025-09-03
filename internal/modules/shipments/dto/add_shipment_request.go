package dto

import (
	"fmt"
	"slices"
)

type AddShipmentRequest struct {
	ShipmentNumber string `json:"shipmentNumber" form:"shipmentNumber" validate:"required,min=3,max=50"`
	ShipmentType   string `json:"shipmentType" form:"shipmentType"`
	SealineCode    string `json:"sealineCode" form:"sealineCode"`
}

func (r *AddShipmentRequest) Validate() error {
	if r.ShipmentNumber == "" {
		return fmt.Errorf("shipment number is required")
	}

	if len(r.ShipmentNumber) < 3 {
		return fmt.Errorf("shipment number must be at least 3 characters")
	}

	if len(r.ShipmentNumber) > 50 {
		return fmt.Errorf("shipment number must be less than 50 characters")
	}

	if !isValidShipmentType(r.ShipmentType) {
		return fmt.Errorf("shipment type must be one of: CT, BK, BL or empty")
	}

	return nil
}

func isValidShipmentType(shipmentType string) bool {
	validTypes := []string{"", "CT", "BK", "BL"}
	return slices.Contains(validTypes, shipmentType)
}
