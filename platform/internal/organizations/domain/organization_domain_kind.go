package domain

import "strings"

type OrganizationDomainKind string

const (
	OrganizationDomainKindSubdomain    OrganizationDomainKind = "subdomain"
	OrganizationDomainKindCustomDomain OrganizationDomainKind = "custom_domain"
)

func ParseOrganizationDomainKind(value string) OrganizationDomainKind {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case string(OrganizationDomainKindSubdomain):
		return OrganizationDomainKindSubdomain
	case string(OrganizationDomainKindCustomDomain):
		return OrganizationDomainKindCustomDomain
	default:
		return OrganizationDomainKind("")
	}
}

func (k OrganizationDomainKind) IsValid() bool {
	return k == OrganizationDomainKindSubdomain || k == OrganizationDomainKindCustomDomain
}
