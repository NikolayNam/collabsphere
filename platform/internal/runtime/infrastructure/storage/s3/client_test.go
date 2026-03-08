package s3

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
)

func TestPresignPutObjectUsesPublicEndpointWhenConfigured(t *testing.T) {
	client, err := NewClient(config.S3{
		Enabled:        true,
		Endpoint:       "http://s3:9000",
		PublicEndpoint: "http://localhost:9000",
		Region:         "us-east-1",
		AccessKey:      "access-key",
		SecretKey:      "secret",
		Bucket:         "collabsphere",
		PathStyle:      true,
		PresignTTL:     15 * time.Minute,
		DownloadTTL:    5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	signedURL, _, err := client.PresignPutObject(context.Background(), "collabsphere", "avatars/user.png")
	if err != nil {
		t.Fatalf("PresignPutObject() error = %v", err)
	}

	parsed, err := url.Parse(signedURL)
	if err != nil {
		t.Fatalf("parse presigned URL: %v", err)
	}
	if parsed.Host != "localhost:9000" {
		t.Fatalf("expected public host localhost:9000, got %q", parsed.Host)
	}
	if !strings.Contains(parsed.Path, "/collabsphere/avatars/user.png") {
		t.Fatalf("unexpected presigned path %q", parsed.Path)
	}
}

func TestPresignPutObjectFallsBackToInternalEndpoint(t *testing.T) {
	client, err := NewClient(config.S3{
		Enabled:     true,
		Endpoint:    "http://s3:9000",
		Region:      "us-east-1",
		AccessKey:   "access-key",
		SecretKey:   "secret",
		Bucket:      "collabsphere",
		PathStyle:   true,
		PresignTTL:  15 * time.Minute,
		DownloadTTL: 5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	signedURL, _, err := client.PresignPutObject(context.Background(), "collabsphere", "imports/test.csv")
	if err != nil {
		t.Fatalf("PresignPutObject() error = %v", err)
	}

	parsed, err := url.Parse(signedURL)
	if err != nil {
		t.Fatalf("parse presigned URL: %v", err)
	}
	if parsed.Host != "s3:9000" {
		t.Fatalf("expected internal host s3:9000, got %q", parsed.Host)
	}
}
