package postgresitest

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
)

type TempDatabase struct {
	AdminDB    *sql.DB
	QueryDB    *sql.DB
	DBName     string
	ConnConfig *pgx.ConnConfig
}

func NewTempDatabase(t *testing.T, prefix string) *TempDatabase {
	t.Helper()

	adminDSN := strings.TrimSpace(os.Getenv("COLLABSPHERE_TEST_POSTGRES_DSN"))
	if adminDSN == "" {
		t.Skip("COLLABSPHERE_TEST_POSTGRES_DSN is not set")
	}

	adminConnConfig, err := pgx.ParseConfig(adminDSN)
	if err != nil {
		t.Fatalf("parse COLLABSPHERE_TEST_POSTGRES_DSN: %v", err)
	}
	adminDB, err := sql.Open("pgx", stdlib.RegisterConnConfig(adminConnConfig))
	if err != nil {
		t.Fatalf("open admin db: %v", err)
	}

	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "collabsphere_it"
	}
	dbName := fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	if _, err := adminDB.ExecContext(context.Background(), `CREATE DATABASE `+dbName); err != nil {
		t.Fatalf("create database %s: %v", dbName, err)
	}

	testConnConfig := adminConnConfig.Copy()
	testConnConfig.Database = dbName

	queryDB, err := sql.Open("pgx", stdlib.RegisterConnConfig(testConnConfig))
	if err != nil {
		t.Fatalf("open query db: %v", err)
	}

	env := &TempDatabase{
		AdminDB:    adminDB,
		QueryDB:    queryDB,
		DBName:     dbName,
		ConnConfig: testConnConfig,
	}
	t.Cleanup(func() {
		_ = queryDB.Close()
		if _, err := adminDB.ExecContext(context.Background(), `
			SELECT pg_terminate_backend(pid)
			FROM pg_stat_activity
			WHERE datname = $1 AND pid <> pg_backend_pid()
		`, dbName); err != nil {
			t.Fatalf("terminate connections for %s: %v", dbName, err)
		}
		if _, err := adminDB.ExecContext(context.Background(), `DROP DATABASE `+dbName); err != nil {
			t.Fatalf("drop database %s: %v", dbName, err)
		}
		_ = adminDB.Close()
	})

	return env
}

func ApplyBundledMigrations(t *testing.T, db *sql.DB) {
	t.Helper()

	dir := BundledMigrationsDir(t)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		path := filepath.Join(dir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read migration %s: %v", name, err)
		}
		statements := gooseUpStatements(string(content))
		tx, err := db.BeginTx(context.Background(), nil)
		if err != nil {
			t.Fatalf("begin tx for %s: %v", name, err)
		}
		for _, statement := range statements {
			if strings.TrimSpace(statement) == "" {
				continue
			}
			if _, err := tx.ExecContext(context.Background(), statement); err != nil {
				_ = tx.Rollback()
				t.Fatalf("exec migration %s: %v", name, err)
			}
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit migration %s: %v", name, err)
		}
	}
}

func BundledMigrationsDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "runtime", "infrastructure", "db", "migrations"))
}

func TestConfig(conn *pgx.ConnConfig, schema string) *config.Config {
	if strings.TrimSpace(schema) == "" {
		schema = "db"
	}
	return &config.Config{
		TZ: "UTC",
		APP: config.App{
			Title:        "integration",
			Version:      "test",
			Host:         "127.0.0.1",
			Port:         "0",
			Environment:  "test",
			TimeoutRead:  15 * time.Second,
			TimeoutWrite: 15 * time.Second,
			TimeoutIdle:  60 * time.Second,
		},
		DB: config.DB{
			Host:         conn.Host,
			Port:         int(conn.Port),
			DBName:       conn.Database,
			DBSchema:     schema,
			Username:     conn.User,
			Password:     conn.Password,
			PasswordFile: "",
		},
		Auth: config.Auth{
			JWTSecret:            "integration-secret",
			AccessTTL:            15 * time.Minute,
			RefreshSessionTTL:    24 * time.Hour,
			GuestAccessTTL:       24 * time.Hour,
			BrowserTicketTTL:     time.Minute,
			PasswordLoginEnabled: true,
			LocalSignupEnabled:   true,
		},
	}
}

func gooseUpStatements(content string) []string {
	lines := strings.Split(content, "\n")
	statements := make([]string, 0, 4)

	var (
		inUp      bool
		inBlock   bool
		statement strings.Builder
	)

	flush := func() {
		value := strings.TrimSpace(statement.String())
		if value != "" {
			statements = append(statements, value)
		}
		statement.Reset()
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch trimmed {
		case "-- +goose Up":
			inUp = true
			continue
		case "-- +goose Down":
			flush()
			return statements
		case "-- +goose StatementBegin":
			inBlock = true
			continue
		case "-- +goose StatementEnd":
			flush()
			inBlock = false
			continue
		}
		if !inUp || strings.HasPrefix(trimmed, "-- +goose") {
			continue
		}
		statement.WriteString(line)
		statement.WriteString("\n")
		if !inBlock && strings.HasSuffix(trimmed, ";") {
			flush()
		}
	}

	flush()
	return statements
}
