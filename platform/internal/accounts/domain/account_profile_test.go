package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAccountApplyProfilePatchUpdatesAndClearsOptionalFields(t *testing.T) {
	email, err := NewEmail("user@example.com")
	if err != nil {
		t.Fatalf("NewEmail: %v", err)
	}
	hash, err := NewPasswordHash("hashed-password")
	if err != nil {
		t.Fatalf("NewPasswordHash: %v", err)
	}
	account, err := NewAccount(NewAccountParams{
		ID:           NewAccountID(),
		Email:        email,
		PasswordHash: hash,
		DisplayName:  stringPtr("User"),
		Now:          time.Date(2026, 3, 8, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}

	avatarID := uuid.New()
	updatedAt := time.Date(2026, 3, 8, 11, 0, 0, 0, time.UTC)
	if err := account.ApplyProfilePatch(AccountProfilePatch{
		DisplayName:    stringPtr(" Updated User "),
		AvatarObjectID: &avatarID,
		Bio:            stringPtr(" About me "),
		Phone:          stringPtr(" +123456789 "),
		Locale:         stringPtr(" ru-RU "),
		Timezone:       stringPtr(" Europe/Moscow "),
		Website:        stringPtr(" https://example.com "),
		UpdatedAt:      updatedAt,
	}); err != nil {
		t.Fatalf("ApplyProfilePatch(update): %v", err)
	}

	if got := account.DisplayName(); got == nil || *got != "Updated User" {
		t.Fatalf("DisplayName mismatch: %v", got)
	}
	if got := account.AvatarObjectID(); got == nil || *got != avatarID {
		t.Fatalf("AvatarObjectID mismatch: %v", got)
	}
	if got := account.Bio(); got == nil || *got != "About me" {
		t.Fatalf("Bio mismatch: %v", got)
	}
	if got := account.Phone(); got == nil || *got != "+123456789" {
		t.Fatalf("Phone mismatch: %v", got)
	}
	if got := account.Locale(); got == nil || *got != "ru-RU" {
		t.Fatalf("Locale mismatch: %v", got)
	}
	if got := account.Timezone(); got == nil || *got != "Europe/Moscow" {
		t.Fatalf("Timezone mismatch: %v", got)
	}
	if got := account.Website(); got == nil || *got != "https://example.com" {
		t.Fatalf("Website mismatch: %v", got)
	}
	if got := account.UpdatedAt(); got == nil || !got.Equal(updatedAt) {
		t.Fatalf("UpdatedAt mismatch: %v", got)
	}

	clearAt := updatedAt.Add(time.Hour)
	if err := account.ApplyProfilePatch(AccountProfilePatch{
		ClearAvatar: true,
		Bio:         stringPtr("   "),
		Website:     stringPtr(" "),
		UpdatedAt:   clearAt,
	}); err != nil {
		t.Fatalf("ApplyProfilePatch(clear): %v", err)
	}

	if account.AvatarObjectID() != nil {
		t.Fatal("AvatarObjectID was not cleared")
	}
	if account.Bio() != nil {
		t.Fatalf("Bio was not cleared: %v", account.Bio())
	}
	if account.Website() != nil {
		t.Fatalf("Website was not cleared: %v", account.Website())
	}
}

func stringPtr(value string) *string {
	return &value
}
