package s3

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
)

const (
	serviceName       = "s3"
	signatureAlg      = "AWS4-HMAC-SHA256"
	unsignedPayload   = "UNSIGNED-PAYLOAD"
	iso8601BasicDate  = "20060102"
	iso8601BasicStamp = "20060102T150405Z"
)

type Client struct {
	endpoint       *url.URL
	publicEndpoint *url.URL
	region         string
	accessKey      string
	secretKey      string
	pathStyle      bool
	presignTTL     time.Duration
	downloadTTL    time.Duration
	httpClient     *http.Client
}

func NewClient(cfg config.S3) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	endpoint, err := parseEndpoint("storage s3 endpoint", cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	publicEndpoint := endpoint
	if strings.TrimSpace(cfg.PublicEndpoint) != "" {
		publicEndpoint, err = parseEndpoint("storage s3 public endpoint", cfg.PublicEndpoint)
		if err != nil {
			return nil, err
		}
	}

	return &Client{
		endpoint:       endpoint,
		publicEndpoint: publicEndpoint,
		region:         strings.TrimSpace(cfg.Region),
		accessKey:      strings.TrimSpace(cfg.AccessKey),
		secretKey:      strings.TrimSpace(cfg.SecretKey),
		pathStyle:      cfg.PathStyle,
		presignTTL:     cfg.PresignTTL,
		downloadTTL:    cfg.DownloadTTL,
		httpClient:     &http.Client{Timeout: 2 * time.Minute},
	}, nil
}

func parseEndpoint(label, raw string) (*url.URL, error) {
	endpoint, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", label, err)
	}
	if endpoint.Scheme == "" || endpoint.Host == "" {
		return nil, fmt.Errorf("%s must include scheme and host", label)
	}

	endpoint.Path = strings.TrimRight(endpoint.Path, "/")
	endpoint.RawPath = ""
	endpoint.RawQuery = ""
	endpoint.Fragment = ""
	return endpoint, nil
}

func (c *Client) PresignPutObject(ctx context.Context, bucket, objectKey string) (string, time.Time, error) {
	return c.presign(ctx, c.publicEndpoint, http.MethodPut, bucket, objectKey, c.presignTTL)
}

func (c *Client) ReadObject(ctx context.Context, bucket, objectKey string) (io.ReadCloser, error) {
	signedURL, _, err := c.presign(ctx, c.endpoint, http.MethodGet, bucket, objectKey, c.downloadTTL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, signedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build s3 get request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download s3 object: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, fmt.Errorf("download s3 object: unexpected status %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func (c *Client) presign(_ context.Context, endpoint *url.URL, method, bucket, objectKey string, expiresIn time.Duration) (string, time.Time, error) {
	bucket = strings.TrimSpace(bucket)
	objectKey = strings.Trim(strings.TrimSpace(objectKey), "/")
	if bucket == "" {
		return "", time.Time{}, fmt.Errorf("s3 bucket is required")
	}
	if objectKey == "" {
		return "", time.Time{}, fmt.Errorf("s3 object key is required")
	}
	if expiresIn <= 0 {
		return "", time.Time{}, fmt.Errorf("s3 presign ttl must be positive")
	}

	now := time.Now().UTC()
	expiresAt := now.Add(expiresIn)
	requestURL, rawPath, err := c.objectURL(endpoint, bucket, objectKey)
	if err != nil {
		return "", time.Time{}, err
	}

	shortDate := now.Format(iso8601BasicDate)
	amzDate := now.Format(iso8601BasicStamp)
	credentialScope := shortDate + "/" + c.region + "/" + serviceName + "/aws4_request"

	query := map[string]string{
		"X-Amz-Algorithm":     signatureAlg,
		"X-Amz-Credential":    c.accessKey + "/" + credentialScope,
		"X-Amz-Date":          amzDate,
		"X-Amz-Expires":       fmt.Sprintf("%d", int(expiresIn.Seconds())),
		"X-Amz-SignedHeaders": "host",
	}

	canonicalRequest := strings.Join([]string{
		method,
		canonicalURI(rawPath),
		canonicalQueryString(query),
		"host:" + requestURL.Host,
		"",
		"host",
		unsignedPayload,
	}, "\n")

	stringToSign := strings.Join([]string{
		signatureAlg,
		amzDate,
		credentialScope,
		hexSHA256(canonicalRequest),
	}, "\n")

	query["X-Amz-Signature"] = c.signature(shortDate, stringToSign)

	signed := *requestURL
	signed.RawQuery = canonicalQueryString(query)
	return signed.String(), expiresAt, nil
}

func (c *Client) objectURL(endpoint *url.URL, bucket, objectKey string) (*url.URL, string, error) {
	out := *endpoint

	pathParts := make([]string, 0, 3)
	if strings.TrimSpace(endpoint.Path) != "" {
		pathParts = append(pathParts, strings.Trim(endpoint.Path, "/"))
	}
	if c.pathStyle {
		pathParts = append(pathParts, bucket, objectKey)
	} else {
		out.Host = bucket + "." + endpoint.Host
		pathParts = append(pathParts, objectKey)
	}

	rawPath := "/" + strings.Join(pathParts, "/")
	out.Path = rawPath
	out.RawPath = canonicalURI(rawPath)
	out.RawQuery = ""
	out.Fragment = ""

	return &out, rawPath, nil
}

func (c *Client) signature(shortDate, stringToSign string) string {
	dateKey := hmacSHA256([]byte("AWS4"+c.secretKey), shortDate)
	regionKey := hmacSHA256(dateKey, c.region)
	serviceKey := hmacSHA256(regionKey, serviceName)
	signingKey := hmacSHA256(serviceKey, "aws4_request")
	return hex.EncodeToString(hmacSHA256(signingKey, stringToSign))
}

func canonicalURI(path string) string {
	if path == "" {
		return "/"
	}

	parts := strings.Split(path, "/")
	escaped := make([]string, 0, len(parts))
	for _, part := range parts {
		escaped = append(escaped, pathEscape(part))
	}
	uri := strings.Join(escaped, "/")
	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}
	return uri
}

func canonicalQueryString(values map[string]string) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, queryEscape(key)+"="+queryEscape(values[key]))
	}
	return strings.Join(parts, "&")
}

func queryEscape(v string) string {
	escaped := url.QueryEscape(v)
	escaped = strings.ReplaceAll(escaped, "+", "%20")
	escaped = strings.ReplaceAll(escaped, "*", "%2A")
	escaped = strings.ReplaceAll(escaped, "%7E", "~")
	return escaped
}

func pathEscape(v string) string {
	escaped := url.PathEscape(v)
	escaped = strings.ReplaceAll(escaped, "+", "%20")
	return escaped
}

func hexSHA256(v string) string {
	sum := sha256.Sum256([]byte(v))
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key []byte, value string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}
