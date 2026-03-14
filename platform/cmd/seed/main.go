package main

import (
	"database/sql"
	"log"
	"net/url"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
)

func main() {
	conf := config.NewFor(config.ProfileSeed)

	dsn, err := conf.DB.DSN()
	if err != nil {
		log.Fatal(err)
	}

	dir := envOrDefault("SEEDS_DIR", envOrDefault("MIGRATIONS_DIR", "/app/seeds"))
	cmd := strings.ToLower(envOrDefault("SEED_CMD", envOrDefault("MIGRATE_CMD", "up")))
	tableName := envOrDefault("SEED_TABLE", "goose_db_version_seeds")

	log.Printf("seed: connecting dsn=%s schema=%s dir=%s cmd=%s table=%s",
		sanitizeDSN(dsn), conf.DB.DBSchema, dir, cmd, tableName,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("seed: close db: %v", err)
		}
	}()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}
	goose.SetTableName(tableName)

	switch cmd {
	case "up":
		err = goose.Up(db, dir)
	case "down":
		err = goose.Down(db, dir)
	case "reset":
		err = goose.Reset(db, dir)
	case "reset-demo":
		baseVersion := int64(1)
		err = goose.DownTo(db, dir, baseVersion)
	case "status":
		err = goose.Status(db, dir)
	case "version":
		var v int64
		v, err = goose.GetDBVersion(db)
		if err == nil {
			log.Printf("seed: current version=%d", v)
		}
	default:
		log.Fatalf("seed: unsupported SEED_CMD=%q (supported: up, down, reset, reset-demo, status, version)", cmd)
	}

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("seed: done")
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
