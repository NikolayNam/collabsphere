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

var (
	rlMu          sync.Mutex
	rlClients     = make(map[string]*rlClient, 256)
	rlLastCleanup time.Time
)

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := remoteIP(r.RemoteAddr)
		lim := getLimiter(ip)

		if !lim.Allow() {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getLimiter(ip string) *rate.Limiter {
	now := time.Now()

	rlMu.Lock()
	defer rlMu.Unlock()

	if rlLastCleanup.IsZero() || now.Sub(rlLastCleanup) >= rlCleanupInterval {
		cleanupLocked(now)
		rlLastCleanup = now
	}

	c := rlClients[ip]
	if c == nil {
		c = &rlClient{
			lim:      rate.NewLimiter(rlRate, rlBurst),
			lastSeen: now,
		}
		rlClients[ip] = c
		return c.lim
	}

	c.lastSeen = now
	return c.lim
}

func cleanupLocked(now time.Time) {
	cutoff := now.Add(-rlTTL)

	for ip, c := range rlClients {
		if c.lastSeen.Before(cutoff) {
			delete(rlClients, ip)
		}
	}
}

func remoteIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil || host == "" {
		return remoteAddr
	}
	return host
}
