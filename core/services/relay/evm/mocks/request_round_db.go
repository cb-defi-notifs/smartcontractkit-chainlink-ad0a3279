// Code generated by mockery v2.42.2. DO NOT EDIT.

package mocks

import (
	context "context"

	evm "github.com/smartcontractkit/chainlink/v2/core/services/relay/evm"
	mock "github.com/stretchr/testify/mock"

	ocr2aggregator "github.com/smartcontractkit/libocr/gethwrappers2/ocr2aggregator"

	sqlutil "github.com/smartcontractkit/chainlink-common/pkg/sqlutil"
)

// RequestRoundDB is an autogenerated mock type for the RequestRoundDB type
type RequestRoundDB struct {
	mock.Mock
}

// LoadLatestRoundRequested provides a mock function with given fields: _a0
func (_m *RequestRoundDB) LoadLatestRoundRequested(_a0 context.Context) (ocr2aggregator.OCR2AggregatorRoundRequested, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for LoadLatestRoundRequested")
	}

	var r0 ocr2aggregator.OCR2AggregatorRoundRequested
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (ocr2aggregator.OCR2AggregatorRoundRequested, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(context.Context) ocr2aggregator.OCR2AggregatorRoundRequested); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(ocr2aggregator.OCR2AggregatorRoundRequested)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SaveLatestRoundRequested provides a mock function with given fields: ctx, rr
func (_m *RequestRoundDB) SaveLatestRoundRequested(ctx context.Context, rr ocr2aggregator.OCR2AggregatorRoundRequested) error {
	ret := _m.Called(ctx, rr)

	if len(ret) == 0 {
		panic("no return value specified for SaveLatestRoundRequested")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, ocr2aggregator.OCR2AggregatorRoundRequested) error); ok {
		r0 = rf(ctx, rr)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WithDataSource provides a mock function with given fields: _a0
func (_m *RequestRoundDB) WithDataSource(_a0 sqlutil.DataSource) evm.RequestRoundDB {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for WithDataSource")
	}

	var r0 evm.RequestRoundDB
	if rf, ok := ret.Get(0).(func(sqlutil.DataSource) evm.RequestRoundDB); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(evm.RequestRoundDB)
		}
	}

	return r0
}

// NewRequestRoundDB creates a new instance of RequestRoundDB. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRequestRoundDB(t interface {
	mock.TestingT
	Cleanup(func())
}) *RequestRoundDB {
	mock := &RequestRoundDB{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
