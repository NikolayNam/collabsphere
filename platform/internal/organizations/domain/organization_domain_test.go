package domain

import (
	"testing"
	"time"
)

func TestNormalizeOrganizationHostname(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "plain host", input: "Tenant1.Collabsphere.Ru", want: "tenant1.collabsphere.ru"},
		{name: "url with port", input: "https://tenant1.collabsphere.ru:8443/", want: "tenant1.collabsphere.ru"},
		{name: "trailing dot", input: "tenant1.collabsphere.ru.", want: "tenant1.collabsphere.ru"},
		{name: "invalid path", input: "tenant1.collabsphere.ru/path", wantErr: true},
		{name: "single label", input: "tenant1", wantErr: true},
		{name: "invalid chars", input: "tenant_1.collabsphere.ru", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeOrganizationHostname(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got hostname %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeOrganizationHostname(%q): %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("hostname mismatch: got %q want %q", got, tt.want)
			}
		})
	}
}

func TestBuildOrganizationDomainsAutoVerifiesSubdomainAndPreservesCustomVerification(t *testing.T) {
	orgID := NewOrganizationID()
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	verifiedAt := now.Add(-time.Hour)

	existingRecord, err := NewOrganizationDomain(NewOrganizationDomainParams{
		ID:             newTestUUID(t, "5c4ec8ff-e81b-4cf4-8edf-e5a67168e2b4"),
		OrganizationID: orgID,
		Hostname:       "brand.collabsphere.ru",
		Kind:           OrganizationDomainKindCustomDomain,
		IsPrimary:      true,
		VerifiedAt:     &verifiedAt,
		Now:            verifiedAt,
	})
	if err != nil {
		t.Fatalf("NewOrganizationDomain(existing): %v", err)
	}

	domains, err := BuildOrganizationDomains(orgID, []OrganizationDomainDraft{
		{Hostname: "tenant1.collabsphere.ru", Kind: "subdomain"},
		{Hostname: "brand.collabsphere.ru", Kind: "custom_domain", IsPrimary: true},
	}, []OrganizationDomain{*existingRecord}, now)
	if err != nil {
		t.Fatalf("BuildOrganizationDomains: %v", err)
	}
	if len(domains) != 2 {
		t.Fatalf("expected 2 domains, got %d", len(domains))
	}

	if !domains[0].IsVerified() {
		t.Fatal("expected subdomain to be auto-verified")
	}
	if domains[1].VerifiedAt() == nil || !domains[1].VerifiedAt().Equal(verifiedAt) {
		t.Fatalf("expected custom domain verification timestamp to be preserved, got %v", domains[1].VerifiedAt())
	}
}

func TestBuildOrganizationDomainsRejectsDuplicateHostname(t *testing.T) {
	_, err := BuildOrganizationDomains(NewOrganizationID(), []OrganizationDomainDraft{
		{Hostname: "tenant1.collabsphere.ru", Kind: "subdomain"},
		{Hostname: "Tenant1.Collabsphere.Ru", Kind: "subdomain"},
	}, nil, time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC))
	if err == nil {
		t.Fatal("expected duplicate hostname error")
	}
	if err != ErrOrganizationDomainDuplicate {
		t.Fatalf("expected ErrOrganizationDomainDuplicate, got %v", err)
	}
}
