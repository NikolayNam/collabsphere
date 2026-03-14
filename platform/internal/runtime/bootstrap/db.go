package bootstrap

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
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

func MustOpenNoopGormDB(dbLog *slog.Logger) *gorm.DB {
	dbLog.Info("db noop connection enabled",
		"event", "db.connect.noop",
	)

	sqlDB := sql.OpenDB(noopConnector{})
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		DisableAutomaticPing: true,
		Logger: log.New(
			dbLog,
			200*time.Millisecond,
			logger.Warn,
			false,
		),
	})
	if err != nil {
		panic(err)
	}
	return db
}

type noopConnector struct{}

func (noopConnector) Connect(context.Context) (driver.Conn, error) {
	return noopConn{}, nil
}

func (noopConnector) Driver() driver.Driver {
	return noopDriver{}
}

type noopDriver struct{}

func (noopDriver) Open(string) (driver.Conn, error) {
	return noopConn{}, nil
}

type noopConn struct{}

func (noopConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("noop connection does not support statements")
}

func (noopConn) Close() error {
	return nil
}

func (noopConn) Begin() (driver.Tx, error) {
	return noopTx{}, nil
}

func (noopConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return noopTx{}, nil
}

func (noopConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return noopResult(0), nil
}

func (noopConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return noopRows{}, nil
}

func (noopConn) Ping(context.Context) error {
	return nil
}

type noopTx struct{}

func (noopTx) Commit() error {
	return nil
}

func (noopTx) Rollback() error {
	return nil
}

type noopRows struct{}

func (noopRows) Columns() []string {
	return nil
}

func (noopRows) Close() error {
	return nil
}

func (noopRows) Next([]driver.Value) error {
	return io.EOF
}

type noopResult int64

func (r noopResult) LastInsertId() (int64, error) {
	return int64(r), nil
}

func (r noopResult) RowsAffected() (int64, error) {
	return int64(r), nil
}
