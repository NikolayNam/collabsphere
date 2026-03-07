package create_product_import_upload

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	catalogaccess "github.com/NikolayNam/collabsphere/internal/catalog/application/access"
	catalogerrors "github.com/NikolayNam/collabsphere/internal/catalog/application/errors"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/ports"
	"github.com/google/uuid"
)

type Handler struct {
	repo          ports.CatalogRepository
	organizations ports.OrganizationReader
	memberships   ports.MembershipReader
	clock         ports.Clock
	storage       ports.ObjectStorage
	bucket        string
}

func NewHandler(repo ports.CatalogRepository, organizations ports.OrganizationReader, memberships ports.MembershipReader, clock ports.Clock, storage ports.ObjectStorage, bucket string) *Handler {
	return &Handler{
		repo:          repo,
		organizations: organizations,
		memberships:   memberships,
		clock:         clock,
		storage:       storage,
		bucket:        strings.TrimSpace(bucket),
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	if err := catalogaccess.RequireOrganizationAccess(ctx, h.organizations, h.memberships, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	if h.storage == nil || h.bucket == "" {
		return nil, catalogerrors.ProductImportUnavailable()
	}

	fileName := strings.TrimSpace(cmd.FileName)
	if fileName == "" {
		return nil, catalogerrors.ProductImportFileInvalid("fileName is required")
	}
	if !strings.EqualFold(filepath.Ext(fileName), ".csv") {
		return nil, catalogerrors.ProductImportFileInvalid("Only .csv import files are supported")
	}

	sizeBytes := int64(0)
	if cmd.SizeBytes != nil {
		if *cmd.SizeBytes < 0 {
			return nil, catalogerrors.ProductImportFileInvalid("sizeBytes must be non-negative")
		}
		sizeBytes = *cmd.SizeBytes
	}

	now := h.clock.Now()
	objectID := uuid.New()
	objectKey := buildObjectKey(cmd.OrganizationID.UUID(), objectID, fileName, now)
	contentType := normalizeOptional(cmd.ContentType)
	checksum := normalizeOptional(cmd.ChecksumSHA256)

	object := &ports.StorageObject{
		ID:             objectID,
		OrganizationID: cmd.OrganizationID,
		Bucket:         h.bucket,
		ObjectKey:      objectKey,
		FileName:       fileName,
		ContentType:    contentType,
		SizeBytes:      sizeBytes,
		ChecksumSHA256: checksum,
		CreatedAt:      now,
	}
	if err := h.repo.CreateStorageObject(ctx, object); err != nil {
		return nil, err
	}

	uploadURL, expiresAt, err := h.storage.PresignPutObject(ctx, object.Bucket, object.ObjectKey)
	if err != nil {
		return nil, catalogerrors.Internal("presign product import upload", err)
	}

	return &Result{
		ObjectID:  object.ID,
		Bucket:    object.Bucket,
		ObjectKey: object.ObjectKey,
		UploadURL: uploadURL,
		ExpiresAt: expiresAt,
		FileName:  object.FileName,
		SizeBytes: object.SizeBytes,
	}, nil
}

func buildObjectKey(organizationID, objectID uuid.UUID, fileName string, now time.Time) string {
	safeName := sanitizeFileName(fileName)
	return strings.Join([]string{
		"catalog",
		"imports",
		organizationID.String(),
		now.UTC().Format("2006"),
		now.UTC().Format("01"),
		now.UTC().Format("02"),
		objectID.String(),
		safeName,
	}, "/")
}

func sanitizeFileName(fileName string) string {
	base := filepath.Base(strings.TrimSpace(fileName))
	if base == "." || base == string(filepath.Separator) || base == "" {
		return "import.csv"
	}

	var b strings.Builder
	b.Grow(len(base))
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '.', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}

	out := strings.Trim(strings.TrimSpace(b.String()), "-")
	if out == "" {
		return "import.csv"
	}
	return out
}

func normalizeOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
