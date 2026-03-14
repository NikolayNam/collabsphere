package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestNewRateLimitBlocksAfterBurst(t *testing.T) {
	limited := NewRateLimit(RateLimitOptions{
		Rate:            rate.Limit(1),
		Burst:           1,
		TTL:             time.Minute,
		CleanupInterval: time.Minute,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
	req1.RemoteAddr = "10.0.0.1:1234"
	rec1 := httptest.NewRecorder()
	limited.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request status = %d, want %d", rec1.Code, http.StatusOK)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
	req2.RemoteAddr = "10.0.0.1:5678"
	rec2 := httptest.NewRecorder()
	limited.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want %d", rec2.Code, http.StatusTooManyRequests)
	}
}

func TestNewRateLimitIsolatedByIP(t *testing.T) {
	limited := NewRateLimit(RateLimitOptions{
		Rate:            rate.Limit(1),
		Burst:           1,
		TTL:             time.Minute,
		CleanupInterval: time.Minute,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	reqA := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
	reqA.RemoteAddr = "10.0.0.1:1234"
	recA := httptest.NewRecorder()
	limited.ServeHTTP(recA, reqA)
	if recA.Code != http.StatusOK {
		t.Fatalf("request A status = %d, want %d", recA.Code, http.StatusOK)
	}

	reqB := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
	reqB.RemoteAddr = "10.0.0.2:1234"
	recB := httptest.NewRecorder()
	limited.ServeHTTP(recB, reqB)
	if recB.Code != http.StatusOK {
		t.Fatalf("request B status = %d, want %d", recB.Code, http.StatusOK)
	}
}
