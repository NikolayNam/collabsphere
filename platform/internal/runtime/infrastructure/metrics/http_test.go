package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestHTTPMiddlewareUsesRoutePatternLabels(t *testing.T) {
	httpMetrics := NewHTTP(HTTPOptions{})
	router := chi.NewRouter()
	router.Use(httpMetrics.Middleware())
	router.Get("/v1/organizations/{organizationId}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/organizations/123", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsRec := httptest.NewRecorder()
	httpMetrics.Handler().ServeHTTP(metricsRec, metricsReq)

	body := metricsRec.Body.String()
	if !strings.Contains(body, `collabsphere_http_requests_total{method="GET",route="/v1/organizations/{organizationId}",status="201"} 1`) {
		t.Fatalf("expected route-pattern metric, got %s", body)
	}
	if !strings.Contains(body, `collabsphere_http_response_size_bytes_bucket{method="GET",route="/v1/organizations/{organizationId}",status="201"`) {
		t.Fatalf("expected response-size metric, got %s", body)
	}
}

func TestHTTPMiddlewareSkipsConfiguredPaths(t *testing.T) {
	httpMetrics := NewHTTP(HTTPOptions{SkippedPaths: []string{"/metrics"}})
	handler := httpMetrics.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsRec := httptest.NewRecorder()
	httpMetrics.Handler().ServeHTTP(metricsRec, metricsReq)

	body := metricsRec.Body.String()
	if strings.Contains(body, `collabsphere_http_requests_total{method="GET",route="unmatched",status="200"} 1`) {
		t.Fatalf("expected skipped path not to be recorded, got %s", body)
	}
}
