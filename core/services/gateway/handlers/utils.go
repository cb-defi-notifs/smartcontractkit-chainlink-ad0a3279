package handlers

import (
	"sync"

	"golang.org/x/time/rate"
)

type RateLimiterConfig struct {
	GlobalRPS    float64 `json:"globalRPS"`
	GlobalBurst  int     `json:"globalBurst"`
	PerUserRPS   float64 `json:"perUserRPS"`
	PerUserBurst int     `json:"perUserBurst"`
}

type RateLimiter struct {
	global  *rate.Limiter
	perUser map[string]*rate.Limiter
	config  RateLimiterConfig
	mu      sync.Mutex
}

func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	return &RateLimiter{
		global:  rate.NewLimiter(rate.Limit(config.GlobalRPS), config.GlobalBurst),
		perUser: make(map[string]*rate.Limiter),
		config:  config,
	}
}

func (rl *RateLimiter) Allow(user string) bool {
	if !rl.global.Allow() {
		return false
	}

	rl.mu.Lock()
	userLimiter, ok := rl.perUser[user]
	if !ok {
		userLimiter = rate.NewLimiter(rate.Limit(rl.config.PerUserRPS), rl.config.PerUserBurst)
		rl.perUser[user] = userLimiter
	}
	rl.mu.Unlock()

	return userLimiter.Allow()
}
