package server

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	rateLimitBurst   = 10
	rateLimitWindow  = time.Minute
)

type rateLimiter struct {
	mu      sync.Mutex
	counts  map[string][]time.Time
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-rateLimitWindow)
	if rl.counts == nil {
		rl.counts = make(map[string][]time.Time)
	}
	times := rl.counts[key]
	var valid []time.Time
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	if len(valid) >= rateLimitBurst {
		return false
	}
	valid = append(valid, now)
	rl.counts[key] = valid
	return true
}

var convertLimiter = &rateLimiter{}

func rateLimitIP(r *http.Request) string {
	ip := r.RemoteAddr
	if os.Getenv("TRUST_PROXY") != "1" {
		return ip
	}
	if f := r.Header.Get("X-Forwarded-For"); f != "" {
		if first := strings.TrimSpace(strings.Split(f, ",")[0]); first != "" {
			return first
		}
	}
	return ip
}

func RateLimitConvert(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := rateLimitIP(r)
		if !convertLimiter.allow(ip) {
			writeError(w, http.StatusTooManyRequests, "too many requests; try again later")
			return
		}
		next(w, r)
	}
}
