package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type visitor struct {
	tokens   float64
	lastSeen time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     float64
	burst    float64
	cleanup  time.Duration
}

func NewRateLimiter(rate int, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     float64(rate),
		burst:    float64(burst),
		cleanup:  10 * time.Minute,
	}

	go rl.cleanupLoop()

	return rl
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		rl.mu.Lock()

		v, exists := rl.visitors[ip]
		now := time.Now()

		if !exists {
			v = &visitor{tokens: rl.burst, lastSeen: now}
			rl.visitors[ip] = v
		}

		elapsed := now.Sub(v.lastSeen).Seconds()
		v.tokens = v.tokens + elapsed*rl.rate
		if v.tokens > rl.burst {
			v.tokens = rl.burst
		}
		v.lastSeen = now

		if v.tokens < 1 {
			rl.mu.Unlock()
			http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
			return
		}

		v.tokens--
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) cleanupLoop() {
	for {
		time.Sleep(rl.cleanup)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.cleanup*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
		log.Debug().Int("visitors", len(rl.visitors)).Msg("rate limiter cleaned up")
	}
}
