// Code generated by mockery v2.32.0. DO NOT EDIT.

package mocks

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// HTTPHandler is an autogenerated mock type for the HTTPHandler type
type HTTPHandler struct {
	mock.Mock
}

// Execute provides a mock function with given fields: _a0, _a1
func (_m *HTTPHandler) Execute(_a0 http.ResponseWriter, _a1 *http.Request) (interface{}, error) {
	ret := _m.Called(_a0, _a1)

	var r0 interface{}
	var r1 error
	if rf, ok := ret.Get(0).(func(http.ResponseWriter, *http.Request) (interface{}, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(http.ResponseWriter, *http.Request) interface{}); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	if rf, ok := ret.Get(1).(func(http.ResponseWriter, *http.Request) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewHTTPHandler creates a new instance of HTTPHandler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewHTTPHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *HTTPHandler {
	mock := &HTTPHandler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
