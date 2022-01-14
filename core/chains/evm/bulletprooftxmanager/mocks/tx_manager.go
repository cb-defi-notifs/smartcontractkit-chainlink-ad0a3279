// Code generated by mockery v2.8.0. DO NOT EDIT.

package mocks

import (
	big "math/big"

	assets "github.com/smartcontractkit/chainlink/core/assets"

	bulletprooftxmanager "github.com/smartcontractkit/chainlink/core/chains/evm/bulletprooftxmanager"

	common "github.com/ethereum/go-ethereum/common"

	context "context"

	gas "github.com/smartcontractkit/chainlink/core/chains/evm/gas"

	mock "github.com/stretchr/testify/mock"

	pg "github.com/smartcontractkit/chainlink/core/services/pg"

	types "github.com/smartcontractkit/chainlink/core/chains/evm/types"
)

// TxManager is an autogenerated mock type for the TxManager type
type TxManager struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *TxManager) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateEthTransaction provides a mock function with given fields: newTx, qopts
func (_m *TxManager) CreateEthTransaction(newTx bulletprooftxmanager.NewTx, qopts ...pg.QOpt) (bulletprooftxmanager.EthTx, error) {
	_va := make([]interface{}, len(qopts))
	for _i := range qopts {
		_va[_i] = qopts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, newTx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 bulletprooftxmanager.EthTx
	if rf, ok := ret.Get(0).(func(bulletprooftxmanager.NewTx, ...pg.QOpt) bulletprooftxmanager.EthTx); ok {
		r0 = rf(newTx, qopts...)
	} else {
		r0 = ret.Get(0).(bulletprooftxmanager.EthTx)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(bulletprooftxmanager.NewTx, ...pg.QOpt) error); ok {
		r1 = rf(newTx, qopts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetGasEstimator provides a mock function with given fields:
func (_m *TxManager) GetGasEstimator() gas.Estimator {
	ret := _m.Called()

	var r0 gas.Estimator
	if rf, ok := ret.Get(0).(func() gas.Estimator); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(gas.Estimator)
		}
	}

	return r0
}

// Healthy provides a mock function with given fields:
func (_m *TxManager) Healthy() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OnNewLongestChain provides a mock function with given fields: ctx, head
func (_m *TxManager) OnNewLongestChain(ctx context.Context, head *types.Head) {
	_m.Called(ctx, head)
}

// Ready provides a mock function with given fields:
func (_m *TxManager) Ready() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RegisterResumeCallback provides a mock function with given fields: fn
func (_m *TxManager) RegisterResumeCallback(fn bulletprooftxmanager.ResumeCallback) {
	_m.Called(fn)
}

// SendEther provides a mock function with given fields: chainID, from, to, value, gasLimit
func (_m *TxManager) SendEther(chainID *big.Int, from common.Address, to common.Address, value assets.Eth, gasLimit uint64) (bulletprooftxmanager.EthTx, error) {
	ret := _m.Called(chainID, from, to, value, gasLimit)

	var r0 bulletprooftxmanager.EthTx
	if rf, ok := ret.Get(0).(func(*big.Int, common.Address, common.Address, assets.Eth, uint64) bulletprooftxmanager.EthTx); ok {
		r0 = rf(chainID, from, to, value, gasLimit)
	} else {
		r0 = ret.Get(0).(bulletprooftxmanager.EthTx)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*big.Int, common.Address, common.Address, assets.Eth, uint64) error); ok {
		r1 = rf(chainID, from, to, value, gasLimit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Start provides a mock function with given fields:
func (_m *TxManager) Start() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Trigger provides a mock function with given fields: addr
func (_m *TxManager) Trigger(addr common.Address) {
	_m.Called(addr)
}
