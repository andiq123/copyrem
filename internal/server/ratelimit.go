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

	times := rl.counts[key]
	n := 0
	for _, t := range times {
		if t.After(cutoff) {
			times[n] = t
			n++
		}
	}
	times = times[:n]

	if n >= rateLimitBurst {
		rl.counts[key] = times
		return false
	}
	rl.counts[key] = append(times, now)
	return true
}

func (rl *rateLimiter) cleanup() {
	for {
		time.Sleep(5 * time.Minute)
		rl.mu.Lock()
		cutoff := time.Now().Add(-rateLimitWindow)
		for key, times := range rl.counts {
			n := 0
			for _, t := range times {
				if t.After(cutoff) {
					times[n] = t
					n++
				}
			}
			if n == 0 {
				delete(rl.counts, key)
			} else {
				rl.counts[key] = times[:n]
			}
		}
		rl.mu.Unlock()
	}
}

var convertLimiter = newRateLimiter()

func newRateLimiter() *rateLimiter {
	rl := &rateLimiter{counts: make(map[string][]time.Time)}
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
