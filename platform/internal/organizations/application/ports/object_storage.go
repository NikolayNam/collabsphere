package ports

import (
	"context"
	"time"
)

type ObjectStorage interface {
	PresignPutObject(ctx context.Context, bucket, objectKey string) (string, time.Time, error)
}
