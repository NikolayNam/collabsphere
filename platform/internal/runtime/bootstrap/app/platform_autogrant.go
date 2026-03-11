package app

import (
	platformdomain "github.com/NikolayNam/collabsphere/internal/platformops/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
)

func buildBootstrapAutoGrantRules(cfg *config.PlatformAutoGrantConfig) []platformdomain.AutoGrantRule {
	if cfg == nil {
		return nil
	}
	rules := make([]platformdomain.AutoGrantRule, 0)
	appendRules := func(role platformdomain.Role, roleCfg config.PlatformAutoGrantRoleConfig) {
		for _, email := range roleCfg.Emails {
			rules = append(rules, platformdomain.AutoGrantRule{
				Role:       role,
				MatchType:  platformdomain.AutoGrantMatchEmail,
				MatchValue: platformdomain.NormalizeAutoGrantMatchValue(platformdomain.AutoGrantMatchEmail, email),
				Source:     platformdomain.AutoGrantSourceBootstrap,
			})
		}
		for _, subject := range roleCfg.Subjects {
			rules = append(rules, platformdomain.AutoGrantRule{
				Role:       role,
				MatchType:  platformdomain.AutoGrantMatchSubject,
				MatchValue: platformdomain.NormalizeAutoGrantMatchValue(platformdomain.AutoGrantMatchSubject, subject),
				Source:     platformdomain.AutoGrantSourceBootstrap,
			})
		}
	}
	appendRules(platformdomain.RolePlatformAdmin, cfg.PlatformAdmin)
	appendRules(platformdomain.RoleSupportOperator, cfg.SupportOperator)
	appendRules(platformdomain.RoleReviewOperator, cfg.ReviewOperator)
	return platformdomain.UniqueSortedAutoGrantRules(rules)
}
