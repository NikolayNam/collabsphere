package ports

import (
	"context"
	"io"
	"time"
)

type ObjectStorage interface {
	PresignPutObject(ctx context.Context, bucket, objectKey string) (string, time.Time, error)
	PutObject(ctx context.Context, bucket, objectKey string, body io.Reader, size int64, contentType string) error
	ReadObject(ctx context.Context, bucket, objectKey string) (io.ReadCloser, error)
}
