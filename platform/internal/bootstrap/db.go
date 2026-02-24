package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/NikolayNam/collabsphere-go/internal/platform/config"
	"github.com/NikolayNam/collabsphere-go/shared/contracts/persistence/gormblame"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func registerDBHooks(db *gorm.DB) {
	if err := gormblame.Register(db); err != nil {
		panic(err)
	}
}

func mustOpenGormDB(conf *config.Config, log *slog.Logger) *gorm.DB {
	dsn, err := conf.DB.DSN()
	if err != nil {
		panic(err)
	}

	log.Info(
		"db connecting",
		"dsn", sanitizeDSN(dsn),
		"host", conf.DB.Host,
		"port", conf.DB.Port,
		"db", conf.DB.DBName,
		"user", conf.DB.Username,
		"schema", conf.DB.DBSchema,
	)

	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newGormLogger(log, conf.DB.Debug),
	})
	if err != nil {
		log.Error("db gorm open failed", "err", err, "dsn", sanitizeDSN(dsn))
		panic(err)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		panic(err)
	}

	configureSQLPool(sqlDB)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		log.Error("db ping failed", "err", err, "dsn", sanitizeDSN(dsn))
		panic(err)
	}

	log.Info("db connected")
	return gdb
}

func configureSQLPool(sqlDB *sql.DB) {
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
}

func sanitizeDSN(dsn string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return `<invalid dsn>`
	}
	if u.User != nil {
		u.User = url.User(u.User.Username())
	}
	return u.String()
}

func newGormLogger(log *slog.Logger, debug bool) gormlogger.Interface {
	level := gormlogger.Error
	if debug {
		level = gormlogger.Info
	}

	return gormlogger.New(
		slogWriter{log: log},
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  level,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}

type slogWriter struct {
	log *slog.Logger
}

func (w slogWriter) Printf(format string, args ...any) {
	w.log.Info("gorm", "msg", fmt.Sprintf(format, args...))
}
