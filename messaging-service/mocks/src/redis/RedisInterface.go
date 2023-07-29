// Code generated by mockery v2.32.0. DO NOT EDIT.

package mocks

import (
	context "context"

	redis "github.com/redis/go-redis/v9"
	mock "github.com/stretchr/testify/mock"

	requests "messaging-service/src/types/requests"

	time "time"
)

// RedisInterface is an autogenerated mock type for the RedisInterface type
type RedisInterface struct {
	mock.Mock
}

// Del provides a mock function with given fields: ctx, key
func (_m *RedisInterface) Del(ctx context.Context, key string) error {
	ret := _m.Called(ctx, key)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAPIKey provides a mock function with given fields: ctx, key
func (_m *RedisInterface) GetAPIKey(ctx context.Context, key string) (*requests.APIKey, error) {
	ret := _m.Called(ctx, key)

	var r0 *requests.APIKey
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*requests.APIKey, error)); ok {
		return rf(ctx, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *requests.APIKey); ok {
		r0 = rf(ctx, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*requests.APIKey)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetEmailByPasswordResetToken provides a mock function with given fields: ctx, key
func (_m *RedisInterface) GetEmailByPasswordResetToken(ctx context.Context, key string) (string, error) {
	ret := _m.Called(ctx, key)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PublishToRedisChannel provides a mock function with given fields: channelName, bytes
func (_m *RedisInterface) PublishToRedisChannel(channelName string, bytes []byte) {
	_m.Called(channelName, bytes)
}

// Set provides a mock function with given fields: ctx, key, value
func (_m *RedisInterface) Set(ctx context.Context, key string, value interface{}) error {
	ret := _m.Called(ctx, key, value)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, interface{}) error); ok {
		r0 = rf(ctx, key, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetWithTTL provides a mock function with given fields: ctx, key, value, ttl
func (_m *RedisInterface) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	ret := _m.Called(ctx, key, value, ttl)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, interface{}, time.Duration) error); ok {
		r0 = rf(ctx, key, value, ttl)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetupChannel provides a mock function with given fields: channelName
func (_m *RedisInterface) SetupChannel(channelName string) *redis.PubSub {
	ret := _m.Called(channelName)

	var r0 *redis.PubSub
	if rf, ok := ret.Get(0).(func(string) *redis.PubSub); ok {
		r0 = rf(channelName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.PubSub)
		}
	}

	return r0
}

// NewRedisInterface creates a new instance of RedisInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRedisInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *RedisInterface {
	mock := &RedisInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
