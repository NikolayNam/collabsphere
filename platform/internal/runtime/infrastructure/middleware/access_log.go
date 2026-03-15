package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

type AccessLogOptions struct {
	QuietPaths []string
	Disabled   bool
}

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func AccessLog(base *slog.Logger, options AccessLogOptions) func(http.Handler) http.Handler {
	if base == nil {
		panic("base logger is nil")
	}

	log := base.With("event", "request.completed")
	quietPaths := make(map[string]struct{}, len(options.QuietPaths))
	for _, path := range options.QuietPaths {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" {
			continue
		}
		quietPaths[trimmed] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw := &statusWriter{ResponseWriter: w}
			start := time.Now()

			next.ServeHTTP(sw, r)

			if options.Disabled {
				return
			}
			if _, quiet := quietPaths[r.URL.Path]; quiet {
				return
			}

			if sw.status == 0 {
				sw.status = http.StatusOK
			}

			level := slog.LevelInfo
			if sw.status >= 500 {
				level = slog.LevelError
			} else if sw.status >= 400 {
				level = slog.LevelWarn
			}

			log.Log(r.Context(), level, "request completed",
				"request_id", chimw.GetReqID(r.Context()),
				"method", r.Method,
				"route", routePattern(r),
				"path", r.URL.Path,
				"status", sw.status,
				"bytes", sw.bytes,
				"duration_ms", time.Since(start).Milliseconds(),
				"remote_ip", clientIP(r.RemoteAddr),
				"user_agent", r.UserAgent(),
			)
		})
	}
}

func clientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

func routePattern(r *http.Request) string {
	routeCtx := chi.RouteContext(r.Context())
	if routeCtx != nil {
		pattern := strings.TrimSpace(routeCtx.RoutePattern())
		if pattern != "" {
			return pattern
		}
	}
	return "unmatched"
}
