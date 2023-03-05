// Code generated by mockery v2.14.0. DO NOT EDIT.

package storage

import (
	pkg "github.com/calvinmclean/automated-garden/garden-app/pkg"
	weather "github.com/calvinmclean/automated-garden/garden-app/pkg/weather"
	mock "github.com/stretchr/testify/mock"

	xid "github.com/rs/xid"
)

// MockClient is an autogenerated mock type for the Client type
type MockClient struct {
	mock.Mock
}

// DeleteGarden provides a mock function with given fields: _a0
func (_m *MockClient) DeleteGarden(_a0 xid.ID) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(xid.ID) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeletePlant provides a mock function with given fields: _a0, _a1
func (_m *MockClient) DeletePlant(_a0 xid.ID, _a1 xid.ID) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(xid.ID, xid.ID) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteWeatherClient provides a mock function with given fields: _a0
func (_m *MockClient) DeleteWeatherClient(_a0 xid.ID) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(xid.ID) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteZone provides a mock function with given fields: _a0, _a1
func (_m *MockClient) DeleteZone(_a0 xid.ID, _a1 xid.ID) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(xid.ID, xid.ID) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetGarden provides a mock function with given fields: _a0
func (_m *MockClient) GetGarden(_a0 xid.ID) (*pkg.Garden, error) {
	ret := _m.Called(_a0)

	var r0 *pkg.Garden
	if rf, ok := ret.Get(0).(func(xid.ID) *pkg.Garden); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pkg.Garden)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(xid.ID) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetGardens provides a mock function with given fields: _a0
func (_m *MockClient) GetGardens(_a0 bool) ([]*pkg.Garden, error) {
	ret := _m.Called(_a0)

	var r0 []*pkg.Garden
	if rf, ok := ret.Get(0).(func(bool) []*pkg.Garden); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*pkg.Garden)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(bool) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPlant provides a mock function with given fields: _a0, _a1
func (_m *MockClient) GetPlant(_a0 xid.ID, _a1 xid.ID) (*pkg.Plant, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *pkg.Plant
	if rf, ok := ret.Get(0).(func(xid.ID, xid.ID) *pkg.Plant); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pkg.Plant)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(xid.ID, xid.ID) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPlants provides a mock function with given fields: _a0, _a1
func (_m *MockClient) GetPlants(_a0 xid.ID, _a1 bool) ([]*pkg.Plant, error) {
	ret := _m.Called(_a0, _a1)

	var r0 []*pkg.Plant
	if rf, ok := ret.Get(0).(func(xid.ID, bool) []*pkg.Plant); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*pkg.Plant)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(xid.ID, bool) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetWeatherClient provides a mock function with given fields: _a0
func (_m *MockClient) GetWeatherClient(_a0 xid.ID) (*weather.Config, error) {
	ret := _m.Called(_a0)

	var r0 *weather.Config
	if rf, ok := ret.Get(0).(func(xid.ID) *weather.Config); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*weather.Config)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(xid.ID) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetWeatherClients provides a mock function with given fields: _a0
func (_m *MockClient) GetWeatherClients(_a0 bool) ([]*weather.Config, error) {
	ret := _m.Called(_a0)

	var r0 []*weather.Config
	if rf, ok := ret.Get(0).(func(bool) []*weather.Config); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*weather.Config)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(bool) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetZone provides a mock function with given fields: _a0, _a1
func (_m *MockClient) GetZone(_a0 xid.ID, _a1 xid.ID) (*pkg.Zone, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *pkg.Zone
	if rf, ok := ret.Get(0).(func(xid.ID, xid.ID) *pkg.Zone); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pkg.Zone)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(xid.ID, xid.ID) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetZones provides a mock function with given fields: _a0, _a1
func (_m *MockClient) GetZones(_a0 xid.ID, _a1 bool) ([]*pkg.Zone, error) {
	ret := _m.Called(_a0, _a1)

	var r0 []*pkg.Zone
	if rf, ok := ret.Get(0).(func(xid.ID, bool) []*pkg.Zone); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*pkg.Zone)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(xid.ID, bool) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SaveGarden provides a mock function with given fields: _a0
func (_m *MockClient) SaveGarden(_a0 *pkg.Garden) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*pkg.Garden) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SavePlant provides a mock function with given fields: _a0, _a1
func (_m *MockClient) SavePlant(_a0 xid.ID, _a1 *pkg.Plant) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(xid.ID, *pkg.Plant) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveWeatherClient provides a mock function with given fields: _a0
func (_m *MockClient) SaveWeatherClient(_a0 *weather.Config) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*weather.Config) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveZone provides a mock function with given fields: _a0, _a1
func (_m *MockClient) SaveZone(_a0 xid.ID, _a1 *pkg.Zone) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(xid.ID, *pkg.Zone) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewMockClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockClient creates a new instance of MockClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockClient(t mockConstructorTestingTNewMockClient) *MockClient {
	mock := &MockClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
