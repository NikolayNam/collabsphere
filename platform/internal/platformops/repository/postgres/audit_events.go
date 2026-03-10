package postgres

import (
	"context"

	"github.com/NikolayNam/collabsphere/internal/platformops/domain"
)

func (r *Repo) Append(ctx context.Context, event domain.AuditEvent) error {
	payload := map[string]any{
		"actor_account_id": event.ActorAccountID,
		"actor_roles":      jsonbExpr(domain.RoleStrings(event.ActorRoles)),
		"actor_bootstrap":  event.ActorBootstrap,
		"action":           event.Action,
		"target_type":      event.TargetType,
		"target_id":        event.TargetID,
		"status":           string(event.Status),
		"summary":          event.Summary,
		"created_at":       event.CreatedAt,
	}
	return r.dbFrom(ctx).WithContext(ctx).Table("iam.platform_audit_events").Create(payload).Error
}
