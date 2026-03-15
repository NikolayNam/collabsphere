package log

import (
	"context"
	"errors"
	"log/slog"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Logger struct {
	log             *slog.Logger
	slowThreshold   time.Duration
	level           logger.LogLevel
	logSQLOnSuccess bool
}

func New(log *slog.Logger, slowThreshold time.Duration, level logger.LogLevel, logSQLOnSuccess bool) *Logger {
	if log == nil {
		panic("gorm logger: nil slog logger")
	}

	return &Logger{
		log:             log,
		slowThreshold:   slowThreshold,
		level:           level,
		logSQLOnSuccess: logSQLOnSuccess,
	}
}

func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	cp := *l
	cp.level = level
	return &cp
}

func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	if l.level < logger.Info {
		return
	}
	l.log.InfoContext(ctx, msg, args...)
}

func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	if l.level < logger.Warn {
		return
	}
	l.log.WarnContext(ctx, msg, args...)
}

func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	if l.level < logger.Error {
		return
	}
	l.log.ErrorContext(ctx, msg, args...)
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level == logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	requestID := chimw.GetReqID(ctx)

	sql, rows := fc()

	attrs := []any{
		"request_id", requestID,
		"duration_ms", elapsed.Milliseconds(),
		"rows", rows,
	}

	if err != nil {
		if l.level >= logger.Error && !errors.Is(err, gorm.ErrRecordNotFound) {
			l.log.ErrorContext(ctx, "db query failed",
				append(attrs,
					"event", "db.query.error",
					"error", err.Error(),
				)...,
			)
		}
		return
	}

	if l.slowThreshold > 0 && elapsed > l.slowThreshold {
		l.log.WarnContext(ctx, "db query slow",
			append(attrs,
				"event", "db.query.slow",
			)...,
		)
		return
	}

	if l.logSQLOnSuccess && l.level >= logger.Info {
		l.log.InfoContext(ctx, "db query",
			append(attrs,
				"event", "db.query",
				"sql", sql,
			)...,
		)
	}
}
