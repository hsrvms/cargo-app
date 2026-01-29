package dto

import (
	"time"

	"github.com/google/uuid"
)

type PublicViewCreateResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	ShareLink string    `json:"shareLink"`
	ExpiresAt time.Time `json:"expiresAt"`
}
