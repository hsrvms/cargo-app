package dto

import (
	"fmt"
	"slices"
	"strings"
)

type AddShipmentRequest struct {
	ShipmentNumber string `json:"shipmentNumber" form:"shipmentNumber" validate:"required,min=3,max=50"`
	ShipmentType   string `json:"shipmentType" form:"shipmentType"`
	SealineCode    string `json:"sealineCode" form:"sealineCode"`
	Recipient      string `json:"recipient" form:"recipient"`
	Address        string `json:"address" form:"address"`
	Notes          string `json:"notes" form:"notes"`
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

	// Trim whitespace and validate new fields
	r.Recipient = strings.TrimSpace(r.Recipient)
	r.Address = strings.TrimSpace(r.Address)
	r.Notes = strings.TrimSpace(r.Notes)

	if len(r.Recipient) > 255 {
		return fmt.Errorf("recipient must be less than 255 characters")
	}

	if len(r.Address) > 1000 {
		return fmt.Errorf("address must be less than 1000 characters")
	}

	if len(r.Notes) > 2000 {
		return fmt.Errorf("notes must be less than 2000 characters")
	}

	return nil
}

func isValidShipmentType(shipmentType string) bool {
	validTypes := []string{"", "CT", "BK", "BL"}
	return slices.Contains(validTypes, shipmentType)
}
