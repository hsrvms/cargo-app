package dto

import (
	"fmt"

	"github.com/google/uuid"
)

type PublicViewCreateRequest struct {
	Name        string      `json:"name" validate:"required,min=3,max=50"`
	ShipmentIDs []uuid.UUID `json:"shipmentIDs" validate:"required,min=1"`
}

func (r *PublicViewCreateRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("'name' is required")
	}

	if len(r.Name) < 3 {
		return fmt.Errorf("'name' must be at least 3 characters")
	}

	if len(r.Name) > 50 {
		return fmt.Errorf("'name' must be less then 50 characters")
	}

	if len(r.ShipmentIDs) < 1 {
		return fmt.Errorf("at least one 'shipmentID' is required")
	}

	return nil
}
