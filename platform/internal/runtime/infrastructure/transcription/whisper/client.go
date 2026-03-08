package whisper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	collabapp "github.com/NikolayNam/collabsphere/internal/collab/application"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
)

type Client struct {
	endpoint   string
	apiKey     string
	model      string
	httpClient *http.Client
}

type responsePayload struct {
	Text     string          `json:"text"`
	Language *string         `json:"language"`
	Segments json.RawMessage `json:"segments"`
}

func NewClient(cfg config.Transcription) (*Client, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	apiKey, err := cfg.APIKeyValue()
	if err != nil {
		return nil, err
	}
	return &Client{endpoint: strings.TrimSpace(cfg.Endpoint), apiKey: apiKey, model: strings.TrimSpace(cfg.Model), httpClient: &http.Client{Timeout: cfg.RequestTimeout}}, nil
}

func (c *Client) Transcribe(ctx context.Context, fileName string, mimeType *string, content io.Reader) (collabapp.TranscriptionResult, error) {
	if c == nil {
		return collabapp.TranscriptionResult{}, fmt.Errorf("transcription client is disabled")
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return collabapp.TranscriptionResult{}, err
	}
	if _, err := io.Copy(part, content); err != nil {
		return collabapp.TranscriptionResult{}, err
	}
	_ = writer.WriteField("model", c.model)
	if mimeType != nil && strings.TrimSpace(*mimeType) != "" {
		_ = writer.WriteField("content_type", strings.TrimSpace(*mimeType))
	}
	if err := writer.Close(); err != nil {
		return collabapp.TranscriptionResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, &body)
	if err != nil {
		return collabapp.TranscriptionResult{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return collabapp.TranscriptionResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return collabapp.TranscriptionResult{}, fmt.Errorf("transcription request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}

	var decoded responsePayload
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return collabapp.TranscriptionResult{}, err
	}
	if len(decoded.Segments) == 0 {
		decoded.Segments = json.RawMessage(`[]`)
	}
	return collabapp.TranscriptionResult{Text: decoded.Text, SegmentsJSON: decoded.Segments, LanguageCode: decoded.Language}, nil
}

var _ = time.Second
