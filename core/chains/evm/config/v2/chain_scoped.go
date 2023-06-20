package v2

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/multierr"

	ocr "github.com/smartcontractkit/libocr/offchainreporting"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting/types"

	"github.com/smartcontractkit/chainlink/v2/core/assets"
	"github.com/smartcontractkit/chainlink/v2/core/chains/evm/config"
	gencfg "github.com/smartcontractkit/chainlink/v2/core/config"
	"github.com/smartcontractkit/chainlink/v2/core/logger"
)

func NewTOMLChainScopedConfig(genCfg gencfg.AppConfig, chain *EVMConfig, lggr logger.Logger) *ChainScoped {
	return &ChainScoped{AppConfig: genCfg, cfg: chain, lggr: lggr}
}

// ChainScoped implements config.ChainScopedConfig with a gencfg.BasicConfig and EVMConfig.
type ChainScoped struct {
	gencfg.AppConfig
	lggr logger.Logger

	cfg *EVMConfig
}

func (c *ChainScoped) ChainID() *big.Int {
	return c.cfg.ChainID.ToInt()
}

func (c *ChainScoped) ChainType() gencfg.ChainType {
	if c.cfg.ChainType == nil {
		return ""
	}
	return gencfg.ChainType(*c.cfg.ChainType)
}

func (c *ChainScoped) Validate() (err error) {
	// Most per-chain validation is done on startup, but this combines globals as well.
	lc := ocrtypes.LocalConfig{
		BlockchainTimeout:                      c.OCR().BlockchainTimeout(),
		ContractConfigConfirmations:            c.EVM().OCR().ContractConfirmations(),
		ContractConfigTrackerPollInterval:      c.OCR().ContractPollInterval(),
		ContractConfigTrackerSubscribeInterval: c.OCR().ContractSubscribeInterval(),
		ContractTransmitterTransmitTimeout:     c.EVM().OCR().ContractTransmitterTransmitTimeout(),
		DatabaseTimeout:                        c.EVM().OCR().DatabaseTimeout(),
		DataSourceTimeout:                      c.OCR().ObservationTimeout(),
		DataSourceGracePeriod:                  c.EVM().OCR().ObservationGracePeriod(),
	}
	if ocrerr := ocr.SanityCheckLocalConfig(lc); ocrerr != nil {
		err = multierr.Append(err, ocrerr)
	}
	return
}

type evmConfig struct {
	c *EVMConfig
}

func (e *evmConfig) BalanceMonitor() config.BalanceMonitor {
	return &balanceMonitorConfig{c: e.c.BalanceMonitor}
}

func (e *evmConfig) Transactions() config.Transactions {
	return &transactionsConfig{c: e.c.Transactions}
}

func (e *evmConfig) HeadTracker() config.HeadTracker {
	return &headTrackerConfig{c: e.c.HeadTracker}
}

func (e *evmConfig) OCR() config.OCR {
	return &ocrConfig{c: e.c.OCR}
}

func (e *evmConfig) OCR2() config.OCR2 {
	return &ocr2Config{c: e.c.OCR2}
}

func (e *evmConfig) GasEstimator() config.GasEstimator {
	return &gasEstimatorConfig{c: e.c.GasEstimator, blockDelay: e.c.RPCBlockQueryDelay, transactionsMaxInFlight: e.c.Transactions.MaxInFlight}
}

func (e *evmConfig) AutoCreateKey() bool {
	return *e.c.AutoCreateKey
}

func (e *evmConfig) BlockBackfillDepth() uint64 {
	return uint64(*e.c.BlockBackfillDepth)
}

func (e *evmConfig) BlockBackfillSkip() bool {
	return *e.c.BlockBackfillSkip
}

func (e *evmConfig) LogBackfillBatchSize() uint32 {
	return *e.c.LogBackfillBatchSize
}

func (e *evmConfig) LogPollInterval() time.Duration {
	return e.c.LogPollInterval.Duration()
}

func (e *evmConfig) FinalityDepth() uint32 {
	return *e.c.FinalityDepth
}

func (e *evmConfig) LogKeepBlocksDepth() uint32 {
	return *e.c.LogKeepBlocksDepth
}

func (e *evmConfig) NonceAutoSync() bool {
	return *e.c.NonceAutoSync
}

func (e *evmConfig) RPCDefaultBatchSize() uint32 {
	return *e.c.RPCDefaultBatchSize
}

func (e *evmConfig) ChainType() gencfg.ChainType {
	if e.c.ChainType == nil {
		return ""
	}
	return gencfg.ChainType(*e.c.ChainType)
}

func (e *evmConfig) KeySpecificMaxGasPriceWei(addr common.Address) *assets.Wei {
	var keySpecific *assets.Wei
	for i := range e.c.KeySpecific {
		ks := e.c.KeySpecific[i]
		if ks.Key.Address() == addr {
			keySpecific = ks.GasEstimator.PriceMax
			break
		}
	}

	chainSpecific := e.GasEstimator().PriceMax()
	if keySpecific != nil && !keySpecific.IsZero() && keySpecific.Cmp(chainSpecific) < 0 {
		return keySpecific
	}

	return e.GasEstimator().PriceMax()
}

func (c *ChainScoped) EVM() config.EVM {
	return &evmConfig{c: c.cfg}
}

func (c *ChainScoped) BlockEmissionIdleWarningThreshold() time.Duration {
	return c.NodeNoNewHeadsThreshold()
}

func (c *ChainScoped) EvmFinalityDepth() uint32 {
	return *c.cfg.FinalityDepth
}

func (c *ChainScoped) FlagsContractAddress() string {
	if c.cfg.FlagsContractAddress == nil {
		return ""
	}
	return c.cfg.FlagsContractAddress.String()
}

func (c *ChainScoped) KeySpecificMaxGasPriceWei(addr common.Address) *assets.Wei {
	var keySpecific *assets.Wei
	for i := range c.cfg.KeySpecific {
		ks := c.cfg.KeySpecific[i]
		if ks.Key.Address() == addr {
			keySpecific = ks.GasEstimator.PriceMax
			break
		}
	}

	chainSpecific := c.EVM().GasEstimator().PriceMax()
	if keySpecific != nil && !keySpecific.IsZero() && keySpecific.Cmp(chainSpecific) < 0 {
		return keySpecific
	}

	return c.EVM().GasEstimator().PriceMax()
}

func (c *ChainScoped) LinkContractAddress() string {
	if c.cfg.LinkContractAddress == nil {
		return ""
	}
	return c.cfg.LinkContractAddress.String()
}

func (c *ChainScoped) OperatorFactoryAddress() string {
	if c.cfg.OperatorFactoryAddress == nil {
		return ""
	}
	return c.cfg.OperatorFactoryAddress.String()
}

func (c *ChainScoped) MinIncomingConfirmations() uint32 {
	return *c.cfg.MinIncomingConfirmations
}

func (c *ChainScoped) MinimumContractPayment() *assets.Link {
	return c.cfg.MinContractPayment
}

func (c *ChainScoped) NodeNoNewHeadsThreshold() time.Duration {
	return c.cfg.NoNewHeadsThreshold.Duration()
}

func (c *ChainScoped) NodePollFailureThreshold() uint32 {
	return *c.cfg.NodePool.PollFailureThreshold
}

func (c *ChainScoped) NodePollInterval() time.Duration {
	return c.cfg.NodePool.PollInterval.Duration()
}

func (c *ChainScoped) NodeSelectionMode() string {
	return *c.cfg.NodePool.SelectionMode
}

func (c *ChainScoped) NodeSyncThreshold() uint32 {
	return *c.cfg.NodePool.SyncThreshold
}
