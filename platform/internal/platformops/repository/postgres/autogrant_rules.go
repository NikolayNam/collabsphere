package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	platformdomain "github.com/NikolayNam/collabsphere/internal/platformops/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm/clause"
)

const pgUniqueViolation = "23505"

type autoGrantRuleRow struct {
	ID                 uuid.UUID  `gorm:"column:id"`
	Role               string     `gorm:"column:role"`
	MatchType          string     `gorm:"column:match_type"`
	MatchValue         string     `gorm:"column:match_value"`
	GrantedByAccountID *uuid.UUID `gorm:"column:granted_by_account_id"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at"`
}

func (r *Repo) ListAutoGrantRules(ctx context.Context) ([]platformdomain.AutoGrantRule, error) {
	out := append([]platformdomain.AutoGrantRule{}, r.bootstrapAutoGrantRules...)
	rows := make([]autoGrantRuleRow, 0)
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.platform_auto_grant_rules").
		Select("id, role, match_type, match_value, granted_by_account_id, created_at, updated_at").
		Order("role ASC, match_type ASC, match_value ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		rule, err := autoGrantRuleFromRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, rule)
	}
	return platformdomain.UniqueSortedAutoGrantRules(out), nil
}

func (r *Repo) CreateAutoGrantRule(ctx context.Context, role platformdomain.Role, matchType platformdomain.AutoGrantMatchType, matchValue string, grantedByAccountID *uuid.UUID, now time.Time) (*platformdomain.AutoGrantRule, error) {
	matchValue = platformdomain.NormalizeAutoGrantMatchValue(matchType, matchValue)
	row := autoGrantRuleRow{
		Role:               string(role),
		MatchType:          string(matchType),
		MatchValue:         matchValue,
		GrantedByAccountID: grantedByAccountID,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.platform_auto_grant_rules").
		Create(&row).Error; err != nil {
		return nil, err
	}
	rule, err := autoGrantRuleFromRow(row)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *Repo) DeleteAutoGrantRule(ctx context.Context, ruleID uuid.UUID) (*platformdomain.AutoGrantRule, error) {
	row := autoGrantRuleRow{}
	tx := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.platform_auto_grant_rules").
		Clauses(clause.Returning{}).
		Where("id = ?", ruleID).
		Delete(&row)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, nil
	}
	rule, err := autoGrantRuleFromRow(row)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *Repo) MatchPlatformRoles(ctx context.Context, subject string, email string, emailVerified bool) ([]string, error) {
	subject = strings.TrimSpace(subject)
	email = platformdomain.NormalizeAutoGrantMatchValue(platformdomain.AutoGrantMatchEmail, email)
	query := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.platform_auto_grant_rules").
		Select("role, match_type, match_value")
	switch {
	case subject != "" && emailVerified && email != "":
		query = query.Where("(match_type = ? AND match_value = ?) OR (match_type = ? AND match_value = ?)", string(platformdomain.AutoGrantMatchSubject), subject, string(platformdomain.AutoGrantMatchEmail), email)
	case subject != "":
		query = query.Where("match_type = ? AND match_value = ?", string(platformdomain.AutoGrantMatchSubject), subject)
	case emailVerified && email != "":
		query = query.Where("match_type = ? AND match_value = ?", string(platformdomain.AutoGrantMatchEmail), email)
	default:
		return nil, nil
	}
	rows := make([]autoGrantRuleRow, 0)
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	roles := make([]platformdomain.Role, 0, len(rows))
	for _, row := range rows {
		role := platformdomain.ParseRole(row.Role)
		if !role.IsValid() {
			return nil, fmt.Errorf("unsupported platform auto-grant role in db: %q", row.Role)
		}
		roles = append(roles, role)
	}
	return platformdomain.RoleStrings(platformdomain.UniqueSortedRoles(roles)), nil
}

func (r *Repo) EnsurePlatformRoles(ctx context.Context, accountID uuid.UUID, roles []string, grantedByAccountID *uuid.UUID, now time.Time) error {
	parsed := make([]platformdomain.Role, 0, len(roles))
	for _, role := range roles {
		parsedRole := platformdomain.ParseRole(role)
		if !parsedRole.IsValid() {
			return fmt.Errorf("unsupported platform role: %q", role)
		}
		parsed = append(parsed, parsedRole)
	}
	parsed = platformdomain.UniqueSortedRoles(parsed)
	for _, role := range parsed {
		if err := r.ensureRole(ctx, accountID, role, grantedByAccountID, now); err != nil {
			return err
		}
	}
	return nil
}

func autoGrantRuleFromRow(row autoGrantRuleRow) (platformdomain.AutoGrantRule, error) {
	role := platformdomain.ParseRole(row.Role)
	if !role.IsValid() {
		return platformdomain.AutoGrantRule{}, fmt.Errorf("unsupported platform auto-grant role in db: %q", row.Role)
	}
	matchType := platformdomain.ParseAutoGrantMatchType(row.MatchType)
	if !matchType.IsValid() {
		return platformdomain.AutoGrantRule{}, fmt.Errorf("unsupported platform auto-grant match type in db: %q", row.MatchType)
	}
	id := row.ID
	createdAt := row.CreatedAt
	updatedAt := row.UpdatedAt
	return platformdomain.AutoGrantRule{
		ID:                 &id,
		Role:               role,
		MatchType:          matchType,
		MatchValue:         platformdomain.NormalizeAutoGrantMatchValue(matchType, row.MatchValue),
		Source:             platformdomain.AutoGrantSourceDatabase,
		CreatedByAccountID: row.GrantedByAccountID,
		CreatedAt:          &createdAt,
		UpdatedAt:          &updatedAt,
	}, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return pgErr.Code == pgUniqueViolation
	}
	return false
}
