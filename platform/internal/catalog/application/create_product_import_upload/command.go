package create_product_import_upload

import (
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
)

type Command struct {
	OrganizationID orgdomain.OrganizationID
	ActorAccountID accdomain.AccountID
	FileName       string
	ContentType    *string
	SizeBytes      *int64
	ChecksumSHA256 *string
}

type Result struct {
	ObjectID  uuid.UUID
	Bucket    string
	ObjectKey string
	UploadURL string
	ExpiresAt time.Time
	FileName  string
	SizeBytes int64
}
