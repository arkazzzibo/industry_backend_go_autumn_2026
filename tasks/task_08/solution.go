package main

import (
	"sync"
	"time"
)

type Clock interface{ Now() time.Time }
type Limiter struct {
	mu     sync.Mutex
	clock  Clock
	rate   float64
	burst  int
	tokens float64
	last   time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	var last time.Time

	if clock != nil {
		last = clock.Now()
	}

	return &Limiter{
		clock:  clock,
		rate:   ratePerSec,
		burst:  burst,
		tokens: float64(burst),
		last:   last,
	}
}
func (l *Limiter) AllowN(n int) bool {
	if n <= 0 || n > l.burst {
		return false
	}
	if l.clock == nil || l.burst <= 0 {
		return false
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock.Now()

	if now.After(l.last) {
		elapsed := now.Sub(l.last)

		if l.rate > 0 {
			l.tokens += elapsed.Seconds() * l.rate

			if l.tokens > float64(l.burst) {
				l.tokens = float64(l.burst)
			}
		}

		l.last = now
	}

	if l.tokens >= float64(n) {
		l.tokens -= float64(n)
		return true
	}
	return false
}
