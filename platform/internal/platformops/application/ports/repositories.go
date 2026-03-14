package ports

import (
	"context"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	authports "github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	"github.com/NikolayNam/collabsphere/internal/platformops/domain"
	"github.com/google/uuid"
)

type RoleBindingRepository interface {
	ListRoles(ctx context.Context, accountID uuid.UUID) ([]domain.Role, error)
	ListAccountIDsByRole(ctx context.Context, role domain.Role) ([]uuid.UUID, error)
	ReplaceRoles(ctx context.Context, accountID uuid.UUID, roles []domain.Role, grantedByAccountID *uuid.UUID, now time.Time) error
}

type AutoGrantRuleRepository interface {
	ListAutoGrantRules(ctx context.Context) ([]domain.AutoGrantRule, error)
	CreateAutoGrantRule(ctx context.Context, role domain.Role, matchType domain.AutoGrantMatchType, matchValue string, grantedByAccountID *uuid.UUID, now time.Time) (*domain.AutoGrantRule, error)
	DeleteAutoGrantRule(ctx context.Context, ruleID uuid.UUID) (*domain.AutoGrantRule, error)
}

type AuditRepository interface {
	Append(ctx context.Context, event domain.AuditEvent) error
}

type AccountReader interface {
	GetByID(ctx context.Context, id accdomain.AccountID) (*accdomain.Account, error)
}

type DashboardReader interface {
	GetDashboardSummary(ctx context.Context) (*domain.DashboardSummary, error)
}

type UploadQueueReader interface {
	ListUploadQueue(ctx context.Context, query domain.UploadQueueQuery) ([]domain.UploadQueueItem, int, error)
}

type OrganizationReviewRepository interface {
	ListOrganizationReviewQueue(ctx context.Context, query domain.OrganizationReviewQueueQuery) ([]domain.OrganizationReviewQueueItem, int, error)
	GetOrganizationReview(ctx context.Context, organizationID uuid.UUID) (*domain.OrganizationReviewDetail, error)
	UpdateCooperationApplicationReview(ctx context.Context, organizationID uuid.UUID, patch domain.CooperationApplicationReviewPatch) (*domain.OrganizationReviewCooperationApplication, error)
	UpdateLegalDocumentReview(ctx context.Context, organizationID, documentID uuid.UUID, patch domain.LegalDocumentReviewPatch) (*domain.OrganizationReviewLegalDocument, error)
}

type ZitadelAdminClient = authports.ZitadelAdminClient
