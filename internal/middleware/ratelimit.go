package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// IP-based limiter with TTL cleanup.
type ipLimiter struct {
	mu       sync.Mutex
	limiters map[string]*clientLimiter
	r        rate.Limit // tokens per second
	b        int        // burst
	ttl      time.Duration
}

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewIPLimiter(r rate.Limit, b int, ttl time.Duration) *ipLimiter {
	l := &ipLimiter{
		limiters: make(map[string]*clientLimiter),
		r:        r,
		b:        b,
		ttl:      ttl,
	}
	go l.cleanupLoop()
	return l
}

func (l *ipLimiter) get(ip string) *rate.Limiter {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if cl, ok := l.limiters[ip]; ok {
		cl.lastSeen = now
		return cl.limiter
	}

	lim := rate.NewLimiter(l.r, l.b)
	l.limiters[ip] = &clientLimiter{
		limiter:  lim,
		lastSeen: now,
	}
	return lim
}

func (l *ipLimiter) cleanupLoop() {
	t := time.NewTicker(time.Minute)
	defer t.Stop()

	for range t.C {
		now := time.Now()

		l.mu.Lock()
		for ip, cl := range l.limiters {
			if now.Sub(cl.lastSeen) > l.ttl {
				delete(l.limiters, ip)
			}
		}
		l.mu.Unlock()
	}
}

func clientIP(r *http.Request) string {
	// chi middleware RealIP sets RemoteAddr to the real client IP when possible.
	// Still normalize it (RemoteAddr can be "IP:port").
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	// If it is already an IP without port
	if net.ParseIP(r.RemoteAddr) != nil {
		return r.RemoteAddr
	}
	return r.RemoteAddr
}

func RateLimit(l *ipLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			lim := l.get(ip)

			if !lim.Allow() {
				w.Header().Set("Retry-After", "1")
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
