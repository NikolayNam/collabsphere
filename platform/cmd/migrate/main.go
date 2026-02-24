package main

import (
	"database/sql"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/NikolayNam/collabsphere-go/internal/platform/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	conf := config.New()

	dsn, err := conf.DB.DSN()
	if err != nil {
		log.Fatal(err)
	}

	dir := envOrDefault("MIGRATIONS_DIR", "/bootstrapapp/migrations")
	cmd := strings.ToLower(envOrDefault("MIGRATE_CMD", "up"))

	log.Printf("migrate: connecting dsn=%s schema=%s dir=%s cmd=%s",
		sanitizeDSN(dsn), conf.DB.DBSchema, dir, cmd,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("migrate: close db: %v", err)
		}
	}()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	switch cmd {
	case "up":
		err = goose.Up(db, dir)
	case "down":
		err = goose.Down(db, dir)
	case "status":
		err = goose.Status(db, dir)
	case "version":
		var v int64
		v, err = goose.GetDBVersion(db)
		if err == nil {
			log.Printf("migrate: current version=%d", v)
		}
	default:
		log.Fatalf("migrate: unsupported MIGRATE_CMD=%q", cmd)
	}

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("migrate: done")
}

func envOrDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func sanitizeDSN(dsn string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return "<invalid dsn>"
	}
	if u.User != nil {
		u.User = url.User(u.User.Username())
	}
	return u.String()
}
