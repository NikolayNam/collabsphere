package ports

import (
	"context"
	"io"
	"time"
)

type ObjectStorage interface {
	PresignPutObject(ctx context.Context, bucket, objectKey string) (string, time.Time, error)
	ReadObject(ctx context.Context, bucket, objectKey string) (io.ReadCloser, error)
}
