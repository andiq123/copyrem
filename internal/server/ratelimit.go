package server

import (
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	rateLimitBurst  = 10
	rateLimitWindow = time.Minute
)

type rateLimiter struct {
	mu     sync.Mutex
	counts map[string][]time.Time
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-rateLimitWindow)
	if rl.counts == nil {
		rl.counts = make(map[string][]time.Time)
	}
	var valid []time.Time
	for _, t := range rl.counts[key] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	if len(valid) >= rateLimitBurst {
		rl.counts[key] = valid
		return false
	}
	valid = append(valid, now)
	rl.counts[key] = valid
	return true
}

func (rl *rateLimiter) cleanup() {
	for {
		time.Sleep(5 * time.Minute)
		rl.mu.Lock()
		cutoff := time.Now().Add(-rateLimitWindow)
		for key, times := range rl.counts {
			var valid []time.Time
			for _, t := range times {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.counts, key)
			} else {
				rl.counts[key] = valid
			}
		}
		rl.mu.Unlock()
	}
}

var convertLimiter = newRateLimiter()

func newRateLimiter() *rateLimiter {
	rl := &rateLimiter{}
	go rl.cleanup()
	return rl
}

func clientIP(r *http.Request) string {
	if os.Getenv("TRUST_PROXY") == "1" {
		if f := r.Header.Get("X-Forwarded-For"); f != "" {
			if first := strings.TrimSpace(strings.Split(f, ",")[0]); first != "" {
				return first
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func RateLimitConvert(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !convertLimiter.allow(clientIP(r)) {
			writeError(w, http.StatusTooManyRequests, "too many requests; try again later")
			return
		}
		next(w, r)
	}
}
