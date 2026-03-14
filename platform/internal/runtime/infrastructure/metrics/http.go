package metrics

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPOptions struct {
	SkippedPaths []string
}

type HTTP struct {
	registry    *prometheus.Registry
	skippedPath map[string]struct{}
	inFlight    prometheus.Gauge
	requests    *prometheus.CounterVec
	duration    *prometheus.HistogramVec
	size        *prometheus.HistogramVec
}

func NewHTTP(options HTTPOptions) *HTTP {
	registry := prometheus.NewRegistry()
	httpMetrics := &HTTP{
		registry:    registry,
		skippedPath: make(map[string]struct{}, len(options.SkippedPaths)),
		inFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "collabsphere",
			Subsystem: "http",
			Name:      "requests_in_flight",
			Help:      "Current number of HTTP requests in flight.",
		}),
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "collabsphere",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of handled HTTP requests.",
		}, []string{"method", "route", "status"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "collabsphere",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "route", "status"}),
		size: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "collabsphere",
			Subsystem: "http",
			Name:      "response_size_bytes",
			Help:      "HTTP response payload size in bytes.",
			Buckets:   []float64{128, 512, 1024, 4 * 1024, 16 * 1024, 64 * 1024, 256 * 1024, 1024 * 1024},
		}, []string{"method", "route", "status"}),
	}
	for _, path := range options.SkippedPaths {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" {
			continue
		}
		httpMetrics.skippedPath[trimmed] = struct{}{}
	}

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		httpMetrics.inFlight,
		httpMetrics.requests,
		httpMetrics.duration,
		httpMetrics.size,
	)

	return httpMetrics
}

func (m *HTTP) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *HTTP) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, skip := m.skippedPath[r.URL.Path]; skip {
				next.ServeHTTP(w, r)
				return
			}

			sw := &statusWriter{ResponseWriter: w}
			start := time.Now()
			m.inFlight.Inc()
			defer m.inFlight.Dec()

			next.ServeHTTP(sw, r)

			if sw.status == 0 {
				sw.status = http.StatusOK
			}

			route := routePattern(r)
			status := strconv.Itoa(sw.status)
			labels := []string{r.Method, route, status}

			m.requests.WithLabelValues(labels...).Inc()
			m.duration.WithLabelValues(labels...).Observe(time.Since(start).Seconds())
			m.size.WithLabelValues(labels...).Observe(float64(sw.bytes))
		})
	}
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
