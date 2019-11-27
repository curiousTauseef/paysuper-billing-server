// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import billing "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
import mock "github.com/stretchr/testify/mock"
import pkg "github.com/paysuper/paysuper-billing-server/internal/pkg"

// PaymentChannelCostMerchantInterface is an autogenerated mock type for the PaymentChannelCostMerchantInterface type
type PaymentChannelCostMerchantInterface struct {
	mock.Mock
}

// Delete provides a mock function with given fields: obj
func (_m *PaymentChannelCostMerchantInterface) Delete(obj *billing.PaymentChannelCostMerchant) error {
	ret := _m.Called(obj)

	var r0 error
	if rf, ok := ret.Get(0).(func(*billing.PaymentChannelCostMerchant) error); ok {
		r0 = rf(obj)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: merchantId, name, payoutCurrency, region, country, mccCode
func (_m *PaymentChannelCostMerchantInterface) Get(merchantId string, name string, payoutCurrency string, region string, country string, mccCode string) ([]*pkg.PaymentChannelCostMerchantSet, error) {
	ret := _m.Called(merchantId, name, payoutCurrency, region, country, mccCode)

	var r0 []*pkg.PaymentChannelCostMerchantSet
	if rf, ok := ret.Get(0).(func(string, string, string, string, string, string) []*pkg.PaymentChannelCostMerchantSet); ok {
		r0 = rf(merchantId, name, payoutCurrency, region, country, mccCode)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*pkg.PaymentChannelCostMerchantSet)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string, string, string) error); ok {
		r1 = rf(merchantId, name, payoutCurrency, region, country, mccCode)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllForMerchant provides a mock function with given fields: merchantId
func (_m *PaymentChannelCostMerchantInterface) GetAllForMerchant(merchantId string) (*billing.PaymentChannelCostMerchantList, error) {
	ret := _m.Called(merchantId)

	var r0 *billing.PaymentChannelCostMerchantList
	if rf, ok := ret.Get(0).(func(string) *billing.PaymentChannelCostMerchantList); ok {
		r0 = rf(merchantId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing.PaymentChannelCostMerchantList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(merchantId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetById provides a mock function with given fields: id
func (_m *PaymentChannelCostMerchantInterface) GetById(id string) (*billing.PaymentChannelCostMerchant, error) {
	ret := _m.Called(id)

	var r0 *billing.PaymentChannelCostMerchant
	if rf, ok := ret.Get(0).(func(string) *billing.PaymentChannelCostMerchant); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing.PaymentChannelCostMerchant)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MultipleInsert provides a mock function with given fields: obj
func (_m *PaymentChannelCostMerchantInterface) MultipleInsert(obj []*billing.PaymentChannelCostMerchant) error {
	ret := _m.Called(obj)

	var r0 error
	if rf, ok := ret.Get(0).(func([]*billing.PaymentChannelCostMerchant) error); ok {
		r0 = rf(obj)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: obj
func (_m *PaymentChannelCostMerchantInterface) Update(obj *billing.PaymentChannelCostMerchant) error {
	ret := _m.Called(obj)

	var r0 error
	if rf, ok := ret.Get(0).(func(*billing.PaymentChannelCostMerchant) error); ok {
		r0 = rf(obj)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}