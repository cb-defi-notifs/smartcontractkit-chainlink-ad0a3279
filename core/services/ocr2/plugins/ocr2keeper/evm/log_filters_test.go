package evm

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/chainlink/v2/core/chains/evm/logpoller/mocks"
	"github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/keeper_registry_logic_b_wrapper_2_1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLogFiltersProvider_Register(t *testing.T) {
	tests := []struct {
		name      string
		errored   bool
		upkeepID  *big.Int
		upkeepCfg keeper_registry_logic_b_wrapper_2_1.KeeperRegistryBase21LogTriggerConfig
	}{
		{
			"happy flow",
			false,
			big.NewInt(111),
			keeper_registry_logic_b_wrapper_2_1.KeeperRegistryBase21LogTriggerConfig{
				ContractAddress: common.BytesToAddress(common.LeftPadBytes([]byte{1, 2, 3, 4}, 20)),
				Topic0:          common.BytesToHash(common.LeftPadBytes([]byte{1, 2, 3, 4}, 32)),
			},
		},
		{
			"empty config",
			true,
			big.NewInt(111),
			keeper_registry_logic_b_wrapper_2_1.KeeperRegistryBase21LogTriggerConfig{},
		},
		{
			"invalid config",
			true,
			big.NewInt(111),
			keeper_registry_logic_b_wrapper_2_1.KeeperRegistryBase21LogTriggerConfig{
				ContractAddress: common.BytesToAddress(common.LeftPadBytes([]byte{}, 20)),
				Topic0:          common.BytesToHash(common.LeftPadBytes([]byte{}, 32)),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mp := new(mocks.LogPoller)
			mp.On("RegisterFilter", mock.Anything).Return(nil)
			lfp := newLogFiltersManager(mp)
			err := lfp.Register(tc.upkeepID.String(), tc.upkeepCfg)
			if tc.errored {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_AddFiltersBySelector(t *testing.T) {
	var zeroBytes [32]byte
	tests := []struct {
		name           string
		filterSelector uint8
		sigs           []common.Hash
		filters        [][]byte
		expectedSigs   []common.Hash
	}{
		{
			"invalid filters",
			1,
			[]common.Hash{},
			[][]byte{
				zeroBytes[:],
			},
			[]common.Hash{},
		},
		{
			"selector 000",
			0,
			[]common.Hash{},
			[][]byte{
				{1},
			},
			[]common.Hash{},
		},
		{
			"selector 001",
			1,
			[]common.Hash{},
			[][]byte{
				{1},
				{2},
				{3},
			},
			[]common.Hash{
				common.BytesToHash(common.LeftPadBytes([]byte{1}, 32)),
			},
		},
		{
			"selector 010",
			2,
			[]common.Hash{},
			[][]byte{
				{1},
				{2},
				{3},
			},
			[]common.Hash{
				common.BytesToHash(common.LeftPadBytes([]byte{2}, 32)),
			},
		},
		{
			"selector 011",
			3,
			[]common.Hash{},
			[][]byte{
				{1},
				{2},
				{3},
			},
			[]common.Hash{
				common.BytesToHash(common.LeftPadBytes([]byte{1}, 32)),
				common.BytesToHash(common.LeftPadBytes([]byte{2}, 32)),
			},
		},
		{
			"selector 100",
			4,
			[]common.Hash{},
			[][]byte{
				{1},
				{2},
				{3},
			},
			[]common.Hash{
				common.BytesToHash(common.LeftPadBytes([]byte{3}, 32)),
			},
		},
		{
			"selector 101",
			5,
			[]common.Hash{},
			[][]byte{
				{1},
				{2},
				{3},
			},
			[]common.Hash{
				common.BytesToHash(common.LeftPadBytes([]byte{1}, 32)),
				common.BytesToHash(common.LeftPadBytes([]byte{3}, 32)),
			},
		},
		{
			"selector 110",
			6,
			[]common.Hash{},
			[][]byte{
				{1},
				{2},
				{3},
			},
			[]common.Hash{
				common.BytesToHash(common.LeftPadBytes([]byte{2}, 32)),
				common.BytesToHash(common.LeftPadBytes([]byte{3}, 32)),
			},
		},
		{
			"selector 111",
			7,
			[]common.Hash{},
			[][]byte{
				{1},
				{2},
				{3},
			},
			[]common.Hash{
				common.BytesToHash(common.LeftPadBytes([]byte{1}, 32)),
				common.BytesToHash(common.LeftPadBytes([]byte{2}, 32)),
				common.BytesToHash(common.LeftPadBytes([]byte{3}, 32)),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sigs := addFiltersBySelector(tc.filterSelector, tc.sigs, tc.filters...)
			if len(sigs) != len(tc.expectedSigs) {
				t.Fatalf("expected %v, got %v", len(tc.expectedSigs), len(sigs))
			}
			for i := range sigs {
				if sigs[i] != tc.expectedSigs[i] {
					t.Fatalf("expected %v, got %v", tc.expectedSigs[i], sigs[i])
				}
			}
		})
	}
}
