package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type rlClient struct {
	lim      *rate.Limiter
	lastSeen time.Time
}

const (
	rlRate            = rate.Limit(10) // requests per second per IP
	rlBurst           = 10
	rlTTL             = 10 * time.Minute
	rlCleanupInterval = 1 * time.Minute
)

type RateLimitOptions struct {
	Rate            rate.Limit
	Burst           int
	TTL             time.Duration
	CleanupInterval time.Duration
}

type rateLimiterState struct {
	mu          sync.Mutex
	clients     map[string]*rlClient
	lastCleanup time.Time
	options     RateLimitOptions
}

func NewRateLimit(options RateLimitOptions) func(http.Handler) http.Handler {
	state := &rateLimiterState{
		clients: make(map[string]*rlClient, 256),
		options: normalizeRateLimitOptions(options),
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := remoteIP(r.RemoteAddr)
			lim := state.getLimiter(ip)

			if !lim.Allow() {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RateLimit(next http.Handler) http.Handler {
	return NewRateLimit(RateLimitOptions{})(next)
}

func (s *rateLimiterState) getLimiter(ip string) *rate.Limiter {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.lastCleanup.IsZero() || now.Sub(s.lastCleanup) >= s.options.CleanupInterval {
		s.cleanupLocked(now)
		s.lastCleanup = now
	}

	c := s.clients[ip]
	if c == nil {
		c = &rlClient{
			lim:      rate.NewLimiter(s.options.Rate, s.options.Burst),
			lastSeen: now,
		}
		s.clients[ip] = c
		return c.lim
	}

	c.lastSeen = now
	return c.lim
}

func (s *rateLimiterState) cleanupLocked(now time.Time) {
	cutoff := now.Add(-s.options.TTL)

	for ip, c := range s.clients {
		if c.lastSeen.Before(cutoff) {
			delete(s.clients, ip)
		}
	}
}

func normalizeRateLimitOptions(options RateLimitOptions) RateLimitOptions {
	out := options
	if out.Rate <= 0 {
		out.Rate = rlRate
	}
	if out.Burst <= 0 {
		out.Burst = rlBurst
	}
	if out.TTL <= 0 {
		out.TTL = rlTTL
	}
	if out.CleanupInterval <= 0 {
		out.CleanupInterval = rlCleanupInterval
	}
	return out
}

func remoteIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil || host == "" {
		return remoteAddr
	}
	return host
}
