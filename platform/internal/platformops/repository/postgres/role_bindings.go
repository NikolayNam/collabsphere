package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/NikolayNam/collabsphere/internal/platformops/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type roleRow struct {
	AccountID uuid.UUID `gorm:"column:account_id"`
	Role      string    `gorm:"column:role"`
}

func (r *Repo) ListRoles(ctx context.Context, accountID uuid.UUID) ([]domain.Role, error) {
	rows := make([]roleRow, 0)
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.platform_role_bindings").
		Select("account_id, role").
		Where("account_id = ?", accountID).
		Order("role ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Role, 0, len(rows))
	for _, row := range rows {
		role := domain.ParseRole(row.Role)
		if !role.IsValid() {
			return nil, fmt.Errorf("unsupported platform role in db: %q", row.Role)
		}
		out = append(out, role)
	}
	return domain.UniqueSortedRoles(out), nil
}

func (r *Repo) ListAccountIDsByRole(ctx context.Context, role domain.Role) ([]uuid.UUID, error) {
	rows := make([]roleRow, 0)
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.platform_role_bindings").
		Select("account_id, role").
		Where("role = ?", string(role)).
		Order("account_id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]uuid.UUID, 0, len(rows))
	seen := make(map[uuid.UUID]struct{}, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.AccountID]; ok {
			continue
		}
		seen[row.AccountID] = struct{}{}
		out = append(out, row.AccountID)
	}
	return out, nil
}

func (r *Repo) ReplaceRoles(ctx context.Context, accountID uuid.UUID, roles []domain.Role, grantedByAccountID *uuid.UUID, now time.Time) error {
	db := r.dbFrom(ctx).WithContext(ctx)
	normalized := domain.UniqueSortedRoles(roles)
	if len(normalized) == 0 {
		return db.Table("iam.platform_role_bindings").Where("account_id = ?", accountID).Delete(nil).Error
	}

	roleNames := make([]string, 0, len(normalized))
	for _, role := range normalized {
		roleNames = append(roleNames, string(role))
	}

	if err := db.Table("iam.platform_role_bindings").
		Where("account_id = ? AND role NOT IN ?", accountID, roleNames).
		Delete(nil).Error; err != nil {
		return err
	}

	for _, role := range normalized {
		if err := r.ensureRole(ctx, accountID, role, grantedByAccountID, now); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repo) ensureRole(ctx context.Context, accountID uuid.UUID, role domain.Role, grantedByAccountID *uuid.UUID, now time.Time) error {
	if !role.IsValid() {
		return fmt.Errorf("unsupported platform role: %q", role)
	}
	return r.dbFrom(ctx).WithContext(ctx).
		Table("iam.platform_role_bindings").
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "account_id"}, {Name: "role"}},
			DoUpdates: clause.Assignments(map[string]any{
				"granted_by_account_id": grantedByAccountID,
				"updated_at":            now,
			}),
		}).
		Create(map[string]any{
			"account_id":            accountID,
			"role":                  string(role),
			"granted_by_account_id": grantedByAccountID,
			"created_at":            now,
			"updated_at":            now,
		}).Error
}

func jsonbExpr(value any) any {
	data, err := json.Marshal(value)
	if err != nil {
		data = []byte("{}")
	}
	return gorm.Expr("?::jsonb", string(data))
}
