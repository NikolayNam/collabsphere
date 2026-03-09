package s3

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
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

func TestPutObjectUsesInternalEndpoint(t *testing.T) {
	const payload = "avatar-bytes"

	var gotMethod string
	var gotPath string
	var gotContentType string
	var gotLength int64
	var gotBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		gotLength = r.ContentLength
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(config.S3{
		Enabled:        true,
		Endpoint:       server.URL,
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

	err = client.PutObject(context.Background(), "collabsphere", "avatars/user.png", strings.NewReader(payload), int64(len(payload)), "image/png")
	if err != nil {
		t.Fatalf("PutObject() error = %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Fatalf("expected method PUT, got %q", gotMethod)
	}
	if gotPath != "/collabsphere/avatars/user.png" {
		t.Fatalf("unexpected request path %q", gotPath)
	}
	if gotContentType != "image/png" {
		t.Fatalf("expected content type image/png, got %q", gotContentType)
	}
	if gotLength != int64(len(payload)) {
		t.Fatalf("expected content length %d, got %d", len(payload), gotLength)
	}
	if gotBody != payload {
		t.Fatalf("unexpected request body %q", gotBody)
	}
}
