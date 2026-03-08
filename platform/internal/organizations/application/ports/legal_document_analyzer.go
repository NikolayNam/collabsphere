package ports

import (
	"context"
	"encoding/json"
	"io"
)

type LegalDocumentAnalysisResult struct {
	ExtractedText        string
	Summary              *string
	ExtractedFieldsJSON  json.RawMessage
	DetectedDocumentType *string
	ConfidenceScore      *float64
}

type LegalDocumentAnalyzer interface {
	Analyze(ctx context.Context, fileName string, mimeType *string, content io.Reader) (LegalDocumentAnalysisResult, error)
}
