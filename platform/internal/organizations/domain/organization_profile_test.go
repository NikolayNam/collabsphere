package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestOrganizationApplyProfilePatchUpdatesAndClearsOptionalFields(t *testing.T) {
	organization, err := NewOrganization(NewOrganizationParams{
		ID:   NewOrganizationID(),
		Name: "Acme Foods",
		Slug: "acme-foods",
		Now:  time.Date(2026, 3, 8, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewOrganization: %v", err)
	}

	logoID := uuid.New()
	updatedAt := time.Date(2026, 3, 8, 11, 0, 0, 0, time.UTC)
	if err := organization.ApplyProfilePatch(OrganizationProfilePatch{
		Name:         stringPtr(" Acme Logistics "),
		Slug:         stringPtr(" acme-logistics "),
		LogoObjectID: &logoID,
		Description:  stringPtr(" Supply chain operator "),
		Website:      stringPtr(" https://acme.example.com "),
		PrimaryEmail: stringPtr(" contact@example.com "),
		Phone:        stringPtr(" +74951234567 "),
		Address:      stringPtr(" Moscow, Lenina 1 "),
		Industry:     stringPtr(" Logistics "),
		UpdatedAt:    updatedAt,
	}); err != nil {
		t.Fatalf("ApplyProfilePatch(update): %v", err)
	}

	if got := organization.Name(); got != "Acme Logistics" {
		t.Fatalf("Name mismatch: %q", got)
	}
	if got := organization.Slug(); got != "acme-logistics" {
		t.Fatalf("Slug mismatch: %q", got)
	}
	if got := organization.LogoObjectID(); got == nil || *got != logoID {
		t.Fatalf("LogoObjectID mismatch: %v", got)
	}
	if got := organization.Description(); got == nil || *got != "Supply chain operator" {
		t.Fatalf("Description mismatch: %v", got)
	}
	if got := organization.Website(); got == nil || *got != "https://acme.example.com" {
		t.Fatalf("Website mismatch: %v", got)
	}
	if got := organization.PrimaryEmail(); got == nil || *got != "contact@example.com" {
		t.Fatalf("PrimaryEmail mismatch: %v", got)
	}
	if got := organization.Phone(); got == nil || *got != "+74951234567" {
		t.Fatalf("Phone mismatch: %v", got)
	}
	if got := organization.Address(); got == nil || *got != "Moscow, Lenina 1" {
		t.Fatalf("Address mismatch: %v", got)
	}
	if got := organization.Industry(); got == nil || *got != "Logistics" {
		t.Fatalf("Industry mismatch: %v", got)
	}
	if got := organization.UpdatedAt(); got == nil || !got.Equal(updatedAt) {
		t.Fatalf("UpdatedAt mismatch: %v", got)
	}

	clearAt := updatedAt.Add(time.Hour)
	if err := organization.ApplyProfilePatch(OrganizationProfilePatch{
		ClearLogo:   true,
		Description: stringPtr("  "),
		Industry:    stringPtr(" "),
		UpdatedAt:   clearAt,
	}); err != nil {
		t.Fatalf("ApplyProfilePatch(clear): %v", err)
	}

	if organization.LogoObjectID() != nil {
		t.Fatal("LogoObjectID was not cleared")
	}
	if organization.Description() != nil {
		t.Fatalf("Description was not cleared: %v", organization.Description())
	}
	if organization.Industry() != nil {
		t.Fatalf("Industry was not cleared: %v", organization.Industry())
	}
}

func stringPtr(value string) *string {
	return &value
}
