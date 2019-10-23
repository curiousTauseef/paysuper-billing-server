// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import billing "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
import mock "github.com/stretchr/testify/mock"
import service "github.com/paysuper/paysuper-billing-server/internal/service"

// CountryServiceInterface is an autogenerated mock type for the CountryServiceInterface type
type CountryServiceInterface struct {
	mock.Mock
}

// GetAll provides a mock function with given fields:
func (_m *CountryServiceInterface) GetAll() (*billing.CountriesList, error) {
	ret := _m.Called()

	var r0 *billing.CountriesList
	if rf, ok := ret.Get(0).(func() *billing.CountriesList); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing.CountriesList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByIsoCodeA2 provides a mock function with given fields: _a0
func (_m *CountryServiceInterface) GetByIsoCodeA2(_a0 string) (*billing.Country, error) {
	ret := _m.Called(_a0)

	var r0 *billing.Country
	if rf, ok := ret.Get(0).(func(string) *billing.Country); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing.Country)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCountriesAndRegionsByTariffRegion provides a mock function with given fields: tariffRegion
func (_m *CountryServiceInterface) GetCountriesAndRegionsByTariffRegion(tariffRegion string) ([]*service.CountryAndRegionItem, error) {
	ret := _m.Called(tariffRegion)

	var r0 []*service.CountryAndRegionItem
	if rf, ok := ret.Get(0).(func(string) []*service.CountryAndRegionItem); ok {
		r0 = rf(tariffRegion)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*service.CountryAndRegionItem)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(tariffRegion)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCountriesWithVatEnabled provides a mock function with given fields:
func (_m *CountryServiceInterface) GetCountriesWithVatEnabled() (*billing.CountriesList, error) {
	ret := _m.Called()

	var r0 *billing.CountriesList
	if rf, ok := ret.Get(0).(func() *billing.CountriesList); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing.CountriesList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Insert provides a mock function with given fields: _a0
func (_m *CountryServiceInterface) Insert(_a0 *billing.Country) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*billing.Country) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IsRegionExists provides a mock function with given fields: _a0
func (_m *CountryServiceInterface) IsRegionExists(_a0 string) (bool, error) {
	ret := _m.Called(_a0)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MultipleInsert provides a mock function with given fields: _a0
func (_m *CountryServiceInterface) MultipleInsert(_a0 []*billing.Country) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func([]*billing.Country) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: _a0
func (_m *CountryServiceInterface) Update(_a0 *billing.Country) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*billing.Country) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
