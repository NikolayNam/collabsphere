package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type PlatformAutoGrantRoleConfig struct {
	Emails   []string `yaml:"emails"`
	Subjects []string `yaml:"subjects"`
}

type PlatformAutoGrantConfig struct {
	PlatformAdmin   PlatformAutoGrantRoleConfig `yaml:"platform_admin"`
	SupportOperator PlatformAutoGrantRoleConfig `yaml:"support_operator"`
	ReviewOperator  PlatformAutoGrantRoleConfig `yaml:"review_operator"`
}

func (a Auth) PlatformAutoGrantRules() (*PlatformAutoGrantConfig, error) {
	path := strings.TrimSpace(a.PlatformAutoGrantFile)
	if path == "" {
		return &PlatformAutoGrantConfig{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read platform auto-grant file: %w", err)
	}
	if strings.TrimSpace(string(data)) == "" {
		return &PlatformAutoGrantConfig{}, nil
	}

	var cfg PlatformAutoGrantConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal platform auto-grant yaml: %w", err)
	}

	cfg.PlatformAdmin.Emails = normalizeAutoGrantEmails(cfg.PlatformAdmin.Emails)
	cfg.PlatformAdmin.Subjects = normalizeStringSlice(cfg.PlatformAdmin.Subjects)
	cfg.SupportOperator.Emails = normalizeAutoGrantEmails(cfg.SupportOperator.Emails)
	cfg.SupportOperator.Subjects = normalizeStringSlice(cfg.SupportOperator.Subjects)
	cfg.ReviewOperator.Emails = normalizeAutoGrantEmails(cfg.ReviewOperator.Emails)
	cfg.ReviewOperator.Subjects = normalizeStringSlice(cfg.ReviewOperator.Subjects)
	return &cfg, nil
}

func normalizeAutoGrantEmails(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
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
