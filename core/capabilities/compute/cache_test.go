package compute

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host"
	"github.com/smartcontractkit/chainlink/v2/core/internal/testutils/wasmtest"
	"github.com/smartcontractkit/chainlink/v2/core/logger"
)

const (
	binaryLocation = "test/simple/cmd/testmodule.wasm"
	binaryCmd      = "core/capabilities/compute/test/simple/cmd"
)

// Verify that cache evicts an expired module.
func TestCache(t *testing.T) {
	t.Parallel()
	clock := clockwork.NewFakeClock()
	tick := 1 * time.Second
	timeout := 1 * time.Second
	reapTicker := make(chan time.Time)

	cache := newModuleCache(clock, tick, timeout, 0)
	cache.onReaper = make(chan struct{}, 1)
	cache.reapTicker = reapTicker
	cache.start()
	defer cache.close()

	binary := wasmtest.CreateTestBinary(binaryCmd, binaryLocation, false, t)
	hmod, err := host.NewModule(&host.ModuleConfig{
		Logger:         logger.TestLogger(t),
		IsUncompressed: true,
	}, binary)
	require.NoError(t, err)

	id := uuid.New().String()
	mod := &module{
		module: hmod,
	}
	cache.add(id, mod)

	got, ok := cache.get(id)
	assert.True(t, ok)

	assert.Equal(t, got, mod)

	clock.Advance(15 * time.Second)
	reapTicker <- time.Now()
	<-cache.onReaper
	_, ok = cache.get(id)
	assert.False(t, ok)
}

// Verify that an expired module is not evicted because evictAfterSize is 1
func TestCache_EvictAfterSize(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)
	clock := clockwork.NewFakeClock()
	tick := 1 * time.Second
	timeout := 1 * time.Second
	reapTicker := make(chan time.Time)

	cache := newModuleCache(clock, tick, timeout, 1)
	cache.onReaper = make(chan struct{}, 1)
	cache.reapTicker = reapTicker
	cache.start()
	defer cache.close()

	binary := wasmtest.CreateTestBinary(binaryCmd, binaryLocation, false, t)
	hmod, err := host.NewModule(&host.ModuleConfig{
		Logger:         logger.TestLogger(t),
		IsUncompressed: true,
	}, binary)
	require.NoError(t, err)

	id := uuid.New().String()
	mod := &module{
		module: hmod,
	}
	cache.add(id, mod)
	assert.Len(t, cache.m, 1)

	got, ok := cache.get(id)
	assert.True(t, ok)

	assert.Equal(t, got, mod)

	clock.Advance(15 * time.Second)
	reapTicker <- time.Now()
	select {
	case <-ctx.Done():
		return
	case <-cache.onReaper:
	}
	_, ok = cache.get(id)
	assert.True(t, ok)
}
