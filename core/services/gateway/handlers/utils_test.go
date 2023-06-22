package handlers_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink/v2/core/services/gateway/handlers"
)

func TestRateLimiter_Simple(t *testing.T) {
	t.Parallel()

	config := handlers.RateLimiterConfig{
		GlobalRPS:    3.0,
		GlobalBurst:  3,
		PerUserRPS:   1.0,
		PerUserBurst: 2,
	}
	rl := handlers.NewRateLimiter(config)
	require.True(t, rl.Allow("user1"))
	require.True(t, rl.Allow("user2"))
	require.True(t, rl.Allow("user1"))
	require.False(t, rl.Allow("user1"))
	require.False(t, rl.Allow("user3"))
}
