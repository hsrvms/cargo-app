package dto

import (
	"fmt"
	"strings"
)

type UpdateShipmentInfoRequest struct {
	// Shipment Information Fields
	Consignee        string `json:"consignee" form:"consignee"`
	Recipient        string `json:"recipient" form:"recipient"`
	AssignedTo       string `json:"assignedTo" form:"assignedTo"`
	PlaceOfLoading   string `json:"placeOfLoading" form:"placeOfLoading"`
	PlaceOfDelivery  string `json:"placeOfDelivery" form:"placeOfDelivery"`
	FinalDestination string `json:"finalDestination" form:"finalDestination"`
	ContainerType    string `json:"containerType" form:"containerType"`
	Shipper          string `json:"shipper" form:"shipper"`
	InvoiceAmount    string `json:"invoiceAmount" form:"invoiceAmount"`
	Cost             string `json:"cost" form:"cost"`
	Customs          string `json:"customs" form:"customs"`
	MBL              string `json:"mbl" form:"mbl"`
	Notes            string `json:"notes" form:"notes"`

	// Boolean Fields
	CustomsProcessed bool `json:"customsProcessed" form:"customsProcessed"`
	Invoiced         bool `json:"invoiced" form:"invoiced"`
	PaymentReceived  bool `json:"paymentReceived" form:"paymentReceived"`
}

// Backward compatibility alias
type UpdateUserShipmentInfoRequest = UpdateShipmentInfoRequest

func (r *UpdateShipmentInfoRequest) Validate() error {
	// Trim whitespace for all string fields
	r.Consignee = strings.TrimSpace(r.Consignee)
	r.Recipient = strings.TrimSpace(r.Recipient)
	r.AssignedTo = strings.TrimSpace(r.AssignedTo)
	r.PlaceOfLoading = strings.TrimSpace(r.PlaceOfLoading)
	r.PlaceOfDelivery = strings.TrimSpace(r.PlaceOfDelivery)
	r.FinalDestination = strings.TrimSpace(r.FinalDestination)
	r.ContainerType = strings.TrimSpace(r.ContainerType)
	r.Shipper = strings.TrimSpace(r.Shipper)
	r.InvoiceAmount = strings.TrimSpace(r.InvoiceAmount)
	r.Cost = strings.TrimSpace(r.Cost)
	r.Customs = strings.TrimSpace(r.Customs)
	r.MBL = strings.TrimSpace(r.MBL)
	r.Notes = strings.TrimSpace(r.Notes)

	// Validate field lengths
	if len(r.Consignee) > 255 {
		return fmt.Errorf("consignee must be less than 255 characters")
	}
	if len(r.Recipient) > 255 {
		return fmt.Errorf("recipient must be less than 255 characters")
	}
	if len(r.AssignedTo) > 255 {
		return fmt.Errorf("assigned to must be less than 255 characters")
	}
	if len(r.PlaceOfLoading) > 255 {
		return fmt.Errorf("place of loading must be less than 255 characters")
	}
	if len(r.PlaceOfDelivery) > 255 {
		return fmt.Errorf("place of delivery must be less than 255 characters")
	}
	if len(r.FinalDestination) > 1000 {
		return fmt.Errorf("final destination must be less than 1000 characters")
	}
	if len(r.ContainerType) > 100 {
		return fmt.Errorf("container type must be less than 100 characters")
	}
	if len(r.Shipper) > 255 {
		return fmt.Errorf("shipper must be less than 255 characters")
	}
	if len(r.InvoiceAmount) > 100 {
		return fmt.Errorf("invoice amount must be less than 100 characters")
	}
	if len(r.Cost) > 100 {
		return fmt.Errorf("cost must be less than 100 characters")
	}
	if len(r.Customs) > 255 {
		return fmt.Errorf("customs must be less than 255 characters")
	}
	if len(r.MBL) > 100 {
		return fmt.Errorf("MBL must be less than 100 characters")
	}
	if len(r.Notes) > 2000 {
		return fmt.Errorf("notes must be less than 2000 characters")
	}

	return nil
}
