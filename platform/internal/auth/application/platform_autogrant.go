package application

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	"github.com/google/uuid"
)

type OIDCPlatformAutoGrantPolicy struct {
	PlatformAdminEmails     []string
	PlatformAdminSubjects   []string
	SupportOperatorEmails   []string
	SupportOperatorSubjects []string
	ReviewOperatorEmails    []string
	ReviewOperatorSubjects  []string
}

type oidcPlatformAutoGrantRule struct {
	role     string
	emails   map[string]struct{}
	subjects map[string]struct{}
}

type oidcPlatformAutoGrantPolicy struct {
	rules []oidcPlatformAutoGrantRule
}

func newOIDCPlatformAutoGrantPolicy(cfg OIDCPlatformAutoGrantPolicy) oidcPlatformAutoGrantPolicy {
	rules := make([]oidcPlatformAutoGrantRule, 0, 3)
	appendRule := func(role string, emails []string, subjects []string) {
		emails = normalizeAutoGrantEmails(emails)
		subjects = normalizeStringSlice(subjects)
		if len(emails) == 0 && len(subjects) == 0 {
			return
		}
		rules = append(rules, oidcPlatformAutoGrantRule{
			role:     role,
			emails:   toSet(emails),
			subjects: toSet(subjects),
		})
	}
	appendRule("platform_admin", cfg.PlatformAdminEmails, cfg.PlatformAdminSubjects)
	appendRule("support_operator", cfg.SupportOperatorEmails, cfg.SupportOperatorSubjects)
	appendRule("review_operator", cfg.ReviewOperatorEmails, cfg.ReviewOperatorSubjects)
	return oidcPlatformAutoGrantPolicy{rules: rules}
}

func (p oidcPlatformAutoGrantPolicy) ResolveRoles(identity *ports.OIDCIdentity) []string {
	if identity == nil {
		return nil
	}
	roles := make([]string, 0, len(p.rules))
	for _, rule := range p.rules {
		if len(rule.subjects) > 0 {
			if _, ok := rule.subjects[strings.TrimSpace(identity.Subject)]; ok {
				roles = append(roles, rule.role)
				continue
			}
		}
		if len(rule.emails) == 0 || !identity.EmailVerified {
			continue
		}
		if _, ok := rule.emails[normalizeAutoGrantEmail(identity.Email)]; ok {
			roles = append(roles, rule.role)
		}
	}
	return uniqueSortedStrings(roles)
}

func (f *oidcFlow) autoGrantPlatformRoles(ctx context.Context, accountID uuid.UUID, identity *ports.OIDCIdentity, now time.Time) error {
	if f == nil || f.platformRoles == nil || accountID == uuid.Nil {
		return nil
	}
	roles := f.autoGrantPolicy.ResolveRoles(identity)
	if identity != nil {
		dynamicRoles, err := f.platformRoles.MatchPlatformRoles(ctx, identity.Subject, identity.Email, identity.EmailVerified)
		if err != nil {
			return err
		}
		roles = append(roles, dynamicRoles...)
	}
	roles = uniqueSortedStrings(roles)
	if len(roles) == 0 {
		return nil
	}
	return f.platformRoles.EnsurePlatformRoles(ctx, accountID, roles, nil, now)
}

func normalizeAutoGrantEmails(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = normalizeAutoGrantEmail(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}

func normalizeStringSlice(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}

func normalizeAutoGrantEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func toSet(values []string) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		out[value] = struct{}{}
	}
	return out
}

func uniqueSortedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
