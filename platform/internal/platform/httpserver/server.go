package httpserver

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Options struct {
	ReadHeaderTimeout int // seconds
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   int // seconds

	// optional:
	Logger *slog.Logger
}

func Run(handler http.Handler, addr string, opt Options) error {
	logger := opt.Logger
	if logger == nil {
		// прод: JSON; локально можно TextHandler
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: time.Duration(opt.ReadHeaderTimeout) * time.Second,
		ReadTimeout:       opt.ReadTimeout,
		WriteTimeout:      opt.WriteTimeout,
		IdleTimeout:       opt.IdleTimeout,

		// важно: ошибки самого сервера в тот же логгер
		ErrorLog: slogToStdLogger(logger, slog.LevelError),
	}

	serverErr := make(chan error, 1)

	go func() {
		logger.Info("http httpserver starting",
			"addr", addr,
			"read_header_timeout_s", opt.ReadHeaderTimeout,
			"read_timeout", opt.ReadTimeout.String(),
			"write_timeout", opt.WriteTimeout.String(),
			"idle_timeout", opt.IdleTimeout.String(),
		)

		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case sig := <-quit:
		logger.Warn("shutdown signal received", "signal", sig.String())
	case err := <-serverErr:
		if err != nil {
			logger.Error("http httpserver failed", "error", err.Error())
		}
		return err
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(opt.ShutdownTimeout)*time.Second,
	)
	defer cancel()

	logger.Info("http httpserver shutting down", "timeout_s", opt.ShutdownTimeout)

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("http httpserver shutdown failed",
			"error", err.Error(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return err
	}

	logger.Info("http httpserver stopped",
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return nil
}

// slogToStdLogger адаптирует slog -> *log.Logger для http.Server.ErrorLog
func slogToStdLogger(l *slog.Logger, level slog.Level) *log.Logger {
	return log.New(&slogWriter{l: l, level: level}, "", 0)
}

type slogWriter struct {
	l     *slog.Logger
	level slog.Level
}

func (w *slogWriter) Write(p []byte) (n int, err error) {
	// p обычно уже с \n
	msg := string(p)
	w.l.Log(context.Background(), w.level, "http httpserver error", "message", msg)
	return len(p), nil
}
