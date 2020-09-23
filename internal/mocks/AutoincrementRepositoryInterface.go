// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import context "context"
import mock "github.com/stretchr/testify/mock"

// AutoincrementRepositoryInterface is an autogenerated mock type for the AutoincrementRepositoryInterface type
type AutoincrementRepositoryInterface struct {
	mock.Mock
}

// GatPayoutAutoincrementId provides a mock function with given fields: ctx
func (_m *AutoincrementRepositoryInterface) GatPayoutAutoincrementId(ctx context.Context) (int64, error) {
	ret := _m.Called(ctx)

	var r0 int64
	if rf, ok := ret.Get(0).(func(context.Context) int64); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
