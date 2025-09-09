package dto

type UpdateUserShipmentInfoRequest struct {
	Recipient string `json:"recipient" form:"recipient"`
	Address   string `json:"address" form:"address"`
	Notes     string `json:"notes" form:"notes"`
}
