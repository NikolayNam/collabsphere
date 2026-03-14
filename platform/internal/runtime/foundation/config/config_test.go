package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestAppMetricsRoutePath(t *testing.T) {
	tests := []struct {
		name string
		app  App
		want string
	}{
		{
			name: "default path when empty",
			app:  App{},
			want: "/metrics",
		},
		{
			name: "keeps absolute path",
			app:  App{MetricsPath: "/internal/metrics"},
			want: "/internal/metrics",
		},
		{
			name: "normalizes missing slash",
			app:  App{MetricsPath: "metrics"},
			want: "/metrics",
		},
	}

	for _, tt := range tests {
		if got := tt.app.MetricsRoutePath(); got != tt.want {
			t.Fatalf("%s: MetricsRoutePath() = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestAppNormalizedEnvironment(t *testing.T) {
	tests := []struct {
		name string
		app  App
		want string
	}{
		{
			name: "normalizes uppercase",
			app:  App{Environment: "DEV"},
			want: "dev",
		},
		{
			name: "trims spaces",
			app:  App{Environment: "  Prod  "},
			want: "prod",
		},
		{
			name: "defaults when empty",
			app:  App{},
			want: "dev",
		},
	}

	for _, tt := range tests {
		if got := tt.app.NormalizedEnvironment(); got != tt.want {
			t.Fatalf("%s: NormalizedEnvironment() = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestConfigValidateSuccess(t *testing.T) {
	tempDir := t.TempDir()
	jwtSecretFile := filepath.Join(tempDir, "jwt_secret")
	dbPasswordFile := filepath.Join(tempDir, "db_password")
	if err := os.WriteFile(jwtSecretFile, []byte("super-secret"), 0o600); err != nil {
		t.Fatalf("WriteFile(jwtSecretFile) error = %v", err)
	}
	if err := os.WriteFile(dbPasswordFile, []byte("postgres"), 0o600); err != nil {
		t.Fatalf("WriteFile(dbPasswordFile) error = %v", err)
	}

	cfg := Config{
		APP: App{
			Title:         "CollabSphere",
			Version:       "dev",
			Host:          "0.0.0.0",
			Port:          "8080",
			PublicBaseURL: "http://localhost:8080",
			TimeoutRead:   15 * time.Second,
			TimeoutWrite:  15 * time.Second,
			TimeoutIdle:   60 * time.Second,
		},
		DB: DB{
			Host:         "localhost",
			Port:         5432,
			DBName:       "postgres",
			DBSchema:     "db",
			Username:     "postgres",
			PasswordFile: dbPasswordFile,
		},
		Auth: Auth{
			JWTSecretFile:        jwtSecretFile,
			AccessTTL:            15 * time.Minute,
			RefreshSessionTTL:    720 * time.Hour,
			GuestAccessTTL:       24 * time.Hour,
			BrowserTicketTTL:     time.Minute,
			BrowserDefaultReturn: "/auth/callback",
			BrowserRedirects:     "http://localhost:3000, http://localhost:3001",
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestConfigValidateRejectsInvalidPublicBaseURL(t *testing.T) {
	cfg := Config{
		APP: App{
			Title:         "CollabSphere",
			Version:       "dev",
			Host:          "0.0.0.0",
			Port:          "8080",
			PublicBaseURL: "/relative",
			TimeoutRead:   time.Second,
			TimeoutWrite:  time.Second,
			TimeoutIdle:   time.Second,
		},
		DB: DB{
			Host:     "localhost",
			Port:     5432,
			DBName:   "postgres",
			DBSchema: "db",
			Username: "postgres",
			Password: "postgres",
		},
		Auth: Auth{
			JWTSecret:         "secret",
			AccessTTL:         time.Minute,
			RefreshSessionTTL: time.Hour,
			GuestAccessTTL:    time.Hour,
			BrowserTicketTTL:  time.Minute,
		},
	}

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "application public base url") {
		t.Fatalf("Validate() error = %v, want public base url validation error", err)
	}
}

func TestConfigValidateRejectsInvalidBrowserRedirectOrigin(t *testing.T) {
	cfg := Config{
		APP: App{
			Title:        "CollabSphere",
			Version:      "dev",
			Host:         "0.0.0.0",
			Port:         "8080",
			TimeoutRead:  time.Second,
			TimeoutWrite: time.Second,
			TimeoutIdle:  time.Second,
		},
		DB: DB{
			Host:     "localhost",
			Port:     5432,
			DBName:   "postgres",
			DBSchema: "db",
			Username: "postgres",
			Password: "postgres",
		},
		Auth: Auth{
			JWTSecret:         "secret",
			AccessTTL:         time.Minute,
			RefreshSessionTTL: time.Hour,
			GuestAccessTTL:    time.Hour,
			BrowserTicketTTL:  time.Minute,
			BrowserRedirects:  "http://localhost:3000/callback",
		},
	}

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "auth browser redirect origin") {
		t.Fatalf("Validate() error = %v, want browser redirect origin validation error", err)
	}
}

func TestConfigValidateRejectsMissingSecretFile(t *testing.T) {
	cfg := Config{
		APP: App{
			Title:        "CollabSphere",
			Version:      "dev",
			Host:         "0.0.0.0",
			Port:         "8080",
			TimeoutRead:  time.Second,
			TimeoutWrite: time.Second,
			TimeoutIdle:  time.Second,
		},
		DB: DB{
			Host:         "localhost",
			Port:         5432,
			DBName:       "postgres",
			DBSchema:     "db",
			Username:     "postgres",
			PasswordFile: filepath.Join(t.TempDir(), "missing_password"),
		},
		Auth: Auth{
			JWTSecret:         "secret",
			AccessTTL:         time.Minute,
			RefreshSessionTTL: time.Hour,
			GuestAccessTTL:    time.Hour,
			BrowserTicketTTL:  time.Minute,
		},
	}

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "postgres password") {
		t.Fatalf("Validate() error = %v, want postgres password validation error", err)
	}
}

func TestConfigValidateRejectsIncompleteZitadelConfig(t *testing.T) {
	cfg := Config{
		APP: App{
			Title:        "CollabSphere",
			Version:      "dev",
			Host:         "0.0.0.0",
			Port:         "8080",
			TimeoutRead:  time.Second,
			TimeoutWrite: time.Second,
			TimeoutIdle:  time.Second,
		},
		DB: DB{
			Host:     "localhost",
			Port:     5432,
			DBName:   "postgres",
			DBSchema: "db",
			Username: "postgres",
			Password: "postgres",
		},
		Auth: Auth{
			JWTSecret:         "secret",
			AccessTTL:         time.Minute,
			RefreshSessionTTL: time.Hour,
			GuestAccessTTL:    time.Hour,
			BrowserTicketTTL:  time.Minute,
			Zitadel: Zitadel{
				Enabled:      true,
				ClientSecret: "client-secret",
			},
		},
	}

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "auth zitadel") {
		t.Fatalf("Validate() error = %v, want zitadel validation error", err)
	}
}

func TestConfigValidateForMigrateAllowsMissingJWT(t *testing.T) {
	cfg := Config{
		DB: DB{
			Host:     "localhost",
			Port:     5432,
			DBName:   "postgres",
			DBSchema: "db",
			Username: "postgres",
			Password: "postgres",
		},
	}

	if err := cfg.ValidateFor(ProfileMigrate); err != nil {
		t.Fatalf("ValidateFor(ProfileMigrate) error = %v", err)
	}
}

func TestConfigValidateForSeedAllowsMissingJWT(t *testing.T) {
	cfg := Config{
		DB: DB{
			Host:     "localhost",
			Port:     5432,
			DBName:   "postgres",
			DBSchema: "db",
			Username: "postgres",
			Password: "postgres",
		},
	}

	if err := cfg.ValidateFor(ProfileSeed); err != nil {
		t.Fatalf("ValidateFor(ProfileSeed) error = %v", err)
	}
}

func TestConfigValidateForContractsAllowsMissingDBAndJWT(t *testing.T) {
	cfg := Config{
		APP: App{
			Title:   "CollabSphere",
			Version: "dev",
		},
	}

	if err := cfg.ValidateFor(ProfileContracts); err != nil {
		t.Fatalf("ValidateFor(ProfileContracts) error = %v", err)
	}
}

func TestConfigValidateForContractsRejectsMissingTitleVersion(t *testing.T) {
	cfg := Config{}

	err := cfg.ValidateFor(ProfileContracts)
	if err == nil || !strings.Contains(err.Error(), "application title is empty") {
		t.Fatalf("ValidateFor(ProfileContracts) error = %v, want missing title validation error", err)
	}
}

func TestConfigValidateForAPIRejectsMissingJWT(t *testing.T) {
	cfg := Config{
		APP: App{
			Title:        "CollabSphere",
			Version:      "dev",
			Host:         "0.0.0.0",
			Port:         "8080",
			TimeoutRead:  time.Second,
			TimeoutWrite: time.Second,
			TimeoutIdle:  time.Second,
		},
		DB: DB{
			Host:     "localhost",
			Port:     5432,
			DBName:   "postgres",
			DBSchema: "db",
			Username: "postgres",
			Password: "postgres",
		},
		Auth: Auth{
			AccessTTL:         time.Minute,
			RefreshSessionTTL: time.Hour,
			GuestAccessTTL:    time.Hour,
			BrowserTicketTTL:  time.Minute,
		},
	}

	err := cfg.ValidateFor(ProfileAPI)
	if err == nil || !strings.Contains(err.Error(), "auth jwt secret") {
		t.Fatalf("ValidateFor(ProfileAPI) error = %v, want missing jwt validation error", err)
	}
}

func TestConfigValidateForWorkerRequiresJWT(t *testing.T) {
	cfg := Config{
		DB: DB{
			Host:     "localhost",
			Port:     5432,
			DBName:   "postgres",
			DBSchema: "db",
			Username: "postgres",
			Password: "postgres",
		},
		Auth: Auth{
			AccessTTL:         time.Minute,
			RefreshSessionTTL: time.Hour,
			GuestAccessTTL:    time.Hour,
		},
	}

	err := cfg.ValidateFor(ProfileWorker)
	if err == nil || !strings.Contains(err.Error(), "auth jwt secret") {
		t.Fatalf("ValidateFor(ProfileWorker) error = %v, want missing jwt validation error", err)
	}
}

func TestDBValidateRejectsInvalidPoolLimits(t *testing.T) {
	db := DB{
		Host:            "localhost",
		Port:            5432,
		DBName:          "postgres",
		DBSchema:        "db",
		Username:        "postgres",
		Password:        "postgres",
		MaxOpenConns:    10,
		MaxIdleConns:    20,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: time.Minute,
	}
	err := db.Validate()
	if err == nil || !strings.Contains(err.Error(), "max idle conns") {
		t.Fatalf("Validate() error = %v, want max idle conns validation error", err)
	}
}
