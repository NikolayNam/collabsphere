package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlatformBootstrapAccountUUIDs(t *testing.T) {
	auth := Auth{PlatformBootstrapIDs: "11111111-1111-1111-1111-111111111111, 11111111-1111-1111-1111-111111111111 22222222-2222-2222-2222-222222222222"}
	ids, err := auth.PlatformBootstrapAccountUUIDs()
	if err != nil {
		t.Fatalf("PlatformBootstrapAccountUUIDs() error = %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("PlatformBootstrapAccountUUIDs() len = %d, want 2", len(ids))
	}
}

func TestPlatformAutoGrantRules(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "platform_auto_grant.yaml")
	content := "platform_admin:\n  emails:\n    - ADMIN@collabsphere.ru\n    - admin@collabsphere.ru\n  subjects:\n    - subject-1\n    - subject-1\nsupport_operator:\n  emails:\n    - support@collabsphere.ru\nreview_operator:\n  subjects:\n    - reviewer-1\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	auth := Auth{PlatformAutoGrantFile: path}
	rules, err := auth.PlatformAutoGrantRules()
	if err != nil {
		t.Fatalf("PlatformAutoGrantRules() error = %v", err)
	}
	if len(rules.PlatformAdmin.Emails) != 1 || rules.PlatformAdmin.Emails[0] != "admin@collabsphere.ru" {
		t.Fatalf("PlatformAutoGrantRules() admin emails = %#v, want normalized unique email", rules.PlatformAdmin.Emails)
	}
	if len(rules.PlatformAdmin.Subjects) != 1 || rules.PlatformAdmin.Subjects[0] != "subject-1" {
		t.Fatalf("PlatformAutoGrantRules() admin subjects = %#v, want normalized unique subject", rules.PlatformAdmin.Subjects)
	}
	if len(rules.SupportOperator.Emails) != 1 || rules.SupportOperator.Emails[0] != "support@collabsphere.ru" {
		t.Fatalf("PlatformAutoGrantRules() support emails = %#v, want support rule", rules.SupportOperator.Emails)
	}
	if len(rules.ReviewOperator.Subjects) != 1 || rules.ReviewOperator.Subjects[0] != "reviewer-1" {
		t.Fatalf("PlatformAutoGrantRules() review subjects = %#v, want review rule", rules.ReviewOperator.Subjects)
	}
}
