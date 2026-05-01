// Package tmdb — limiter.go: process-wide token-bucket rate limiter for TMDb HTTP calls.
//
// Why: parallel scan workers can fan out enough requests to trip TMDb's
// ~50 req/sec ceiling. We cap ourselves at 40 req/sec with a burst of 40
// so a freshly-started scan can dispatch a full burst, then steady-state.
//
// Implementation: classic token bucket. A single goroutine refills tokens
// every (1s / rate) up to burst capacity. Wait() blocks until a token is
// available. The limiter is created once via DefaultLimiter() and shared
// across all *Client instances in the process — TMDb rate limits are
// per-IP, not per-client.
package tmdb

import (
	"sync"
	"time"
)

// LimiterRate is the steady-state requests-per-second cap.
const LimiterRate = 40

// LimiterBurst is the max tokens the bucket can hold (initial burst capacity).
const LimiterBurst = 40

// Limiter is a token-bucket rate limiter.
type Limiter struct {
	tokens chan struct{}
	stop   chan struct{}
	once   sync.Once
}

// NewLimiter returns a started limiter with the given rate and burst.
func NewLimiter(rate, burst int) *Limiter {
	l := &Limiter{
		tokens: make(chan struct{}, burst),
		stop:   make(chan struct{}),
	}
	for i := 0; i < burst; i++ {
		l.tokens <- struct{}{}
	}
	go l.refill(rate)
	return l
}

// Wait blocks until a token is available. Safe for concurrent use.
func (l *Limiter) Wait() {
	if l == nil {
		return
	}
	<-l.tokens
}

// Stop halts the refill goroutine. Safe to call multiple times.
func (l *Limiter) Stop() {
	if l == nil {
		return
	}
	l.once.Do(func() { close(l.stop) })
}

func (l *Limiter) refill(rate int) {
	if rate <= 0 {
		return
	}
	interval := time.Second / time.Duration(rate)
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-l.stop:
			return
		case <-t.C:
			l.tryAddToken()
		}
	}
}

func (l *Limiter) tryAddToken() {
	select {
	case l.tokens <- struct{}{}:
	default:
	}
}

var (
	defaultLimiter     *Limiter
	defaultLimiterOnce sync.Once
)

// DefaultLimiter returns the process-wide TMDb rate limiter (lazy init).
func DefaultLimiter() *Limiter {
	defaultLimiterOnce.Do(func() {
		defaultLimiter = NewLimiter(LimiterRate, LimiterBurst)
	})
	return defaultLimiter
}
