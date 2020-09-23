// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import billingpb "github.com/paysuper/paysuper-proto/go/billingpb"
import context "context"
import mock "github.com/stretchr/testify/mock"

// ProductRepositoryInterface is an autogenerated mock type for the ProductRepositoryInterface type
type ProductRepositoryInterface struct {
	mock.Mock
}

// CountByProject provides a mock function with given fields: _a0, _a1
func (_m *ProductRepositoryInterface) CountByProject(_a0 context.Context, _a1 string) (int64, error) {
	ret := _m.Called(_a0, _a1)

	var r0 int64
	if rf, ok := ret.Get(0).(func(context.Context, string) int64); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CountByProjectSku provides a mock function with given fields: _a0, _a1, _a2
func (_m *ProductRepositoryInterface) CountByProjectSku(_a0 context.Context, _a1 string, _a2 string) (int64, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 int64
	if rf, ok := ret.Get(0).(func(context.Context, string, string) int64); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Find provides a mock function with given fields: _a0, _a1, _a2, _a3, _a4, _a5, _a6, _a7
func (_m *ProductRepositoryInterface) Find(_a0 context.Context, _a1 string, _a2 string, _a3 string, _a4 string, _a5 int32, _a6 int64, _a7 int64) ([]*billingpb.Product, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3, _a4, _a5, _a6, _a7)

	var r0 []*billingpb.Product
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string, int32, int64, int64) []*billingpb.Product); ok {
		r0 = rf(_a0, _a1, _a2, _a3, _a4, _a5, _a6, _a7)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*billingpb.Product)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, string, int32, int64, int64) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3, _a4, _a5, _a6, _a7)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindCount provides a mock function with given fields: _a0, _a1, _a2, _a3, _a4, _a5
func (_m *ProductRepositoryInterface) FindCount(_a0 context.Context, _a1 string, _a2 string, _a3 string, _a4 string, _a5 int32) (int64, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3, _a4, _a5)

	var r0 int64
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string, int32) int64); ok {
		r0 = rf(_a0, _a1, _a2, _a3, _a4, _a5)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, string, int32) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3, _a4, _a5)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetById provides a mock function with given fields: _a0, _a1
func (_m *ProductRepositoryInterface) GetById(_a0 context.Context, _a1 string) (*billingpb.Product, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *billingpb.Product
	if rf, ok := ret.Get(0).(func(context.Context, string) *billingpb.Product); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billingpb.Product)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MultipleInsert provides a mock function with given fields: _a0, _a1
func (_m *ProductRepositoryInterface) MultipleInsert(_a0 context.Context, _a1 []*billingpb.Product) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []*billingpb.Product) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Upsert provides a mock function with given fields: ctx, product
func (_m *ProductRepositoryInterface) Upsert(ctx context.Context, product *billingpb.Product) error {
	ret := _m.Called(ctx, product)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *billingpb.Product) error); ok {
		r0 = rf(ctx, product)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
