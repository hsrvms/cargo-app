package dto

import "github.com/google/uuid"

type BulkDeleteUserShipmentsRequest struct {
	ShipmentIDs []uuid.UUID `json:"shipmentIDs"`
}
