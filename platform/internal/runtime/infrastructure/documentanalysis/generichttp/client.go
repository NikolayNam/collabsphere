package generichttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	orgports "github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
)

type Client struct {
	endpoint   string
	apiKey     string
	provider   string
	model      string
	httpClient *http.Client
}

type responsePayload struct {
	Text                 string          `json:"text"`
	Summary              *string         `json:"summary"`
	Fields               json.RawMessage `json:"fields"`
	DetectedDocumentType *string         `json:"detectedDocumentType"`
	ConfidenceScore      *float64        `json:"confidenceScore"`
}

func NewClient(cfg config.DocumentAnalysis) (*Client, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Client{
		endpoint:   strings.TrimSpace(cfg.Endpoint),
		apiKey:     strings.TrimSpace(cfg.APIKey),
		provider:   strings.TrimSpace(cfg.Provider),
		model:      strings.TrimSpace(cfg.Model),
		httpClient: &http.Client{Timeout: cfg.RequestTimeout},
	}, nil
}

func (c *Client) Analyze(ctx context.Context, fileName string, mimeType *string, content io.Reader) (orgports.LegalDocumentAnalysisResult, error) {
	if c == nil {
		return orgports.LegalDocumentAnalysisResult{}, fmt.Errorf("document analysis client is disabled")
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return orgports.LegalDocumentAnalysisResult{}, err
	}
	if _, err := io.Copy(part, content); err != nil {
		return orgports.LegalDocumentAnalysisResult{}, err
	}
	_ = writer.WriteField("provider", c.provider)
	_ = writer.WriteField("model", c.model)
	if mimeType != nil && strings.TrimSpace(*mimeType) != "" {
		_ = writer.WriteField("content_type", strings.TrimSpace(*mimeType))
	}
	if err := writer.Close(); err != nil {
		return orgports.LegalDocumentAnalysisResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, &body)
	if err != nil {
		return orgports.LegalDocumentAnalysisResult{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return orgports.LegalDocumentAnalysisResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return orgports.LegalDocumentAnalysisResult{}, fmt.Errorf("document analysis request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}

	var decoded responsePayload
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return orgports.LegalDocumentAnalysisResult{}, err
	}
	if len(decoded.Fields) == 0 {
		decoded.Fields = json.RawMessage(`{}`)
	}
	return orgports.LegalDocumentAnalysisResult{
		ExtractedText:        decoded.Text,
		Summary:              decoded.Summary,
		ExtractedFieldsJSON:  decoded.Fields,
		DetectedDocumentType: decoded.DetectedDocumentType,
		ConfidenceScore:      decoded.ConfidenceScore,
	}, nil
}
