package bootstrap

import (
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"

	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/persistence/gorm/hooks/blame"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/persistence/gorm/log"
)

func RegisterDBHooks(db *gorm.DB) {
	if err := blame.Register(db); err != nil {
		panic(err)
	}
}

func MustOpenGormDB(conf *config.Config, dbLog *slog.Logger) *gorm.DB {

	dsn, err := conf.DB.DSN()
	if err != nil {
		panic(err)
	}

	dbLog.Info("db connecting",
		"event", "db.connect.start",
		"host", conf.DB.Host,
		"port", conf.DB.Port,
		"db", conf.DB.DBName,
		"schema", conf.DB.DBSchema,
		"user", conf.DB.Username,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: log.New(
			dbLog,
			200*time.Millisecond, // slow query threshold
			logger.Warn,          // success queries не логируем
			false,                // SQL на успешных запросах не логируем
		),
	})
	if err != nil {
		dbLog.Error("db connect failed",
			"event", "db.connect.error",
			"host", conf.DB.Host,
			"port", conf.DB.Port,
			"db", conf.DB.DBName,
			"error", err.Error(),
		)
		panic(err)
	}

	dbLog.Info("db connected",
		"event", "db.connect.success",
		"host", conf.DB.Host,
		"port", conf.DB.Port,
		"db", conf.DB.DBName,
		"schema", conf.DB.DBSchema,
	)

	return db
}
