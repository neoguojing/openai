package utils

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type UserLimiterMap struct {
	mu       sync.Mutex
	limiter  map[string]*rate.Limiter
	interval time.Duration
}

func NewUserLimiter(interval time.Duration) *UserLimiterMap {

	return &UserLimiterMap{
		limiter:  make(map[string]*rate.Limiter),
		interval: interval,
	}
}

func (ulm *UserLimiterMap) CanAccess(user string) bool {

	limiter := ulm.getLimiter(user)
	err := limiter.Wait(context.Background())

	return err == nil
}

func (ulm *UserLimiterMap) getLimiter(user string) *rate.Limiter {
	ulm.mu.Lock()
	defer ulm.mu.Unlock()

	limiter, ok := ulm.limiter[user]
	if !ok {
		limiter = rate.NewLimiter(rate.Every(ulm.interval), 1)
		ulm.limiter[user] = limiter
	}

	return limiter
}
