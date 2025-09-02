package dto

import "github.com/google/uuid"

type GridDataRequest struct {
	StartRow    int                    `json:"startRow"`
	EndRow      int                    `json:"endRow"`
	SortModel   []SortModel            `json:"sortModel"`
	FilterModel map[string]FilterModel `json:"filterModel"`
}

type SortModel struct {
	ColId string `json:"colId"`
	Sort  string `json:"sort"` // "asc" or "desc"
}

type FilterModel struct {
	FilterType string   `json:"filterType"`
	Type       string   `json:"type"`
	Filter     string   `json:"filter"`
	Values     []string `json:"values"`
}

type BulkDeleteRequest struct {
	ShipmentIDs []uuid.UUID `json:"shipment_ids"`
}
