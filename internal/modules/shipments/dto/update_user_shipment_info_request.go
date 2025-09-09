package dto

import (
	"fmt"
	"strings"
)

type UpdateUserShipmentInfoRequest struct {
	Recipient string `json:"recipient" form:"recipient"`
	Address   string `json:"address" form:"address"`
	Notes     string `json:"notes" form:"notes"`
}

func (r *UpdateUserShipmentInfoRequest) Validate() error {
	// Trim whitespace
	r.Recipient = strings.TrimSpace(r.Recipient)
	r.Address = strings.TrimSpace(r.Address)
	r.Notes = strings.TrimSpace(r.Notes)

	// Validate field lengths
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
