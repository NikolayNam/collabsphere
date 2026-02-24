package middleware

import (
	"net/http"
	"time"

	"log/slog"

	appLogger "github.com/NikolayNam/collabsphere-go/internal/platform/logger"
	chimw "github.com/go-chi/chi/v5/middleware"
)

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

func AccessLog() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw := &statusWriter{ResponseWriter: w}
			start := time.Now()

			next.ServeHTTP(sw, r)

			dur := time.Since(start)

			l := appLogger.From(r.Context())

			level := slog.LevelInfo
			if sw.status >= 500 {
				level = slog.LevelError
			} else if sw.status >= 400 {
				level = slog.LevelWarn
			}

			reqID := chimw.GetReqID(r.Context())

			l.Log(r.Context(), level, "http request",
				"request_id", reqID,
				"status", sw.status,
				"bytes", sw.bytes,
				"duration_ms", dur.Milliseconds(),
				"remote_ip", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}
