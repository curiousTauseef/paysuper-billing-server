// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import context "context"
import grpc "github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
import mock "github.com/stretchr/testify/mock"

// UserProfileRepositoryInterface is an autogenerated mock type for the UserProfileRepositoryInterface type
type UserProfileRepositoryInterface struct {
	mock.Mock
}

// GetById provides a mock function with given fields: _a0, _a1
func (_m *UserProfileRepositoryInterface) GetById(_a0 context.Context, _a1 string) (*grpc.UserProfile, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *grpc.UserProfile
	if rf, ok := ret.Get(0).(func(context.Context, string) *grpc.UserProfile); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*grpc.UserProfile)
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

// GetByUserId provides a mock function with given fields: _a0, _a1
func (_m *UserProfileRepositoryInterface) GetByUserId(_a0 context.Context, _a1 string) (*grpc.UserProfile, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *grpc.UserProfile
	if rf, ok := ret.Get(0).(func(context.Context, string) *grpc.UserProfile); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*grpc.UserProfile)
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
