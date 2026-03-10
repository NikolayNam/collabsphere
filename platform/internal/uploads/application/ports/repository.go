package ports

import (
	"context"
	"time"

	uploaddomain "github.com/NikolayNam/collabsphere/internal/uploads/domain"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, upload *uploaddomain.Upload) error
	GetByID(ctx context.Context, uploadID uuid.UUID) (*uploaddomain.Upload, error)
	GetByObjectID(ctx context.Context, objectID uuid.UUID) (*uploaddomain.Upload, error)
	MarkReady(ctx context.Context, uploadID uuid.UUID, actualSizeBytes *int64, resultKind uploaddomain.ResultKind, resultID uuid.UUID, completedAt time.Time, updatedAt time.Time) (*uploaddomain.Upload, error)
	MarkFailed(ctx context.Context, uploadID uuid.UUID, errorCode, errorMessage string, updatedAt time.Time) (*uploaddomain.Upload, error)
}
