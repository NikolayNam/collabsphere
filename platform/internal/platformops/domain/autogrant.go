package domain

import (
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AutoGrantSource string

const (
	AutoGrantSourceBootstrap AutoGrantSource = "bootstrap_config"
	AutoGrantSourceDatabase  AutoGrantSource = "database"
)

type AutoGrantMatchType string

const (
	AutoGrantMatchEmail   AutoGrantMatchType = "email"
	AutoGrantMatchSubject AutoGrantMatchType = "subject"
)

type AutoGrantRule struct {
	ID                 *uuid.UUID
	Role               Role
	MatchType          AutoGrantMatchType
	MatchValue         string
	Source             AutoGrantSource
	CreatedByAccountID *uuid.UUID
	CreatedAt          *time.Time
	UpdatedAt          *time.Time
}

func ParseAutoGrantMatchType(raw string) AutoGrantMatchType {
	return AutoGrantMatchType(strings.ToLower(strings.TrimSpace(raw)))
}

func (m AutoGrantMatchType) IsValid() bool {
	switch m {
	case AutoGrantMatchEmail, AutoGrantMatchSubject:
		return true
	default:
		return false
	}
}

func NormalizeAutoGrantMatchValue(matchType AutoGrantMatchType, value string) string {
	value = strings.TrimSpace(value)
	if matchType == AutoGrantMatchEmail {
		return strings.ToLower(value)
	}
	return value
}

func (s AutoGrantSource) IsValid() bool {
	switch s {
	case AutoGrantSourceBootstrap, AutoGrantSourceDatabase:
		return true
	default:
		return false
	}
}

func UniqueSortedAutoGrantRules(values []AutoGrantRule) []AutoGrantRule {
	if len(values) == 0 {
		return nil
	}
	type key struct {
		role      Role
		matchType AutoGrantMatchType
		match     string
		source    AutoGrantSource
	}
	seen := make(map[key]struct{}, len(values))
	out := make([]AutoGrantRule, 0, len(values))
	for _, value := range values {
		if !value.Role.IsValid() || !value.MatchType.IsValid() || !value.Source.IsValid() {
			continue
		}
		value.MatchValue = NormalizeAutoGrantMatchValue(value.MatchType, value.MatchValue)
		if value.MatchValue == "" {
			continue
		}
		k := key{role: value.Role, matchType: value.MatchType, match: value.MatchValue, source: value.Source}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, value)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Source != out[j].Source {
			return autoGrantSourceOrder(out[i].Source) < autoGrantSourceOrder(out[j].Source)
		}
		if out[i].Role != out[j].Role {
			return roleOrder(out[i].Role) < roleOrder(out[j].Role)
		}
		if out[i].MatchType != out[j].MatchType {
			return autoGrantMatchTypeOrder(out[i].MatchType) < autoGrantMatchTypeOrder(out[j].MatchType)
		}
		return out[i].MatchValue < out[j].MatchValue
	})
	return out
}

func autoGrantSourceOrder(source AutoGrantSource) int {
	switch source {
	case AutoGrantSourceBootstrap:
		return 0
	case AutoGrantSourceDatabase:
		return 1
	default:
		return 99
	}
}

func autoGrantMatchTypeOrder(matchType AutoGrantMatchType) int {
	switch matchType {
	case AutoGrantMatchEmail:
		return 0
	case AutoGrantMatchSubject:
		return 1
	default:
		return 99
	}
}
