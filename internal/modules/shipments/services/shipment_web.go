package services

import (
	"context"
	"go-starter/internal/modules/shipments/dto"

	"github.com/google/uuid"
)

func (s *shipmentService) GetShipmentsForGrid(ctx context.Context, userID uuid.UUID, req *dto.GridDataRequest) (*dto.GridDataResponse, error) {
	// shipments, totalCount, err := s.repo.GetShipmentsForGrid(ctx, userID, req)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to fetch shipments for grid: %w", err)
	// }
	return nil, nil
}
