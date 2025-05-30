// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// SecretsGetter is an autogenerated mock type for the SecretsGetter type
type SecretsGetter struct {
	mock.Mock
}

type SecretsGetter_Expecter struct {
	mock *mock.Mock
}

func (_m *SecretsGetter) EXPECT() *SecretsGetter_Expecter {
	return &SecretsGetter_Expecter{mock: &_m.Mock}
}

// Secrets provides a mock function with given fields: namespace
func (_m *SecretsGetter) Secrets(namespace string) v1.SecretInterface {
	ret := _m.Called(namespace)

	if len(ret) == 0 {
		panic("no return value specified for Secrets")
	}

	var r0 v1.SecretInterface
	if rf, ok := ret.Get(0).(func(string) v1.SecretInterface); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1.SecretInterface)
		}
	}

	return r0
}

// SecretsGetter_Secrets_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Secrets'
type SecretsGetter_Secrets_Call struct {
	*mock.Call
}

// Secrets is a helper method to define mock.On call
//   - namespace string
func (_e *SecretsGetter_Expecter) Secrets(namespace interface{}) *SecretsGetter_Secrets_Call {
	return &SecretsGetter_Secrets_Call{Call: _e.mock.On("Secrets", namespace)}
}

func (_c *SecretsGetter_Secrets_Call) Run(run func(namespace string)) *SecretsGetter_Secrets_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *SecretsGetter_Secrets_Call) Return(_a0 v1.SecretInterface) *SecretsGetter_Secrets_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SecretsGetter_Secrets_Call) RunAndReturn(run func(string) v1.SecretInterface) *SecretsGetter_Secrets_Call {
	_c.Call.Return(run)
	return _c
}

// NewSecretsGetter creates a new instance of SecretsGetter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSecretsGetter(t interface {
	mock.TestingT
	Cleanup(func())
}) *SecretsGetter {
	mock := &SecretsGetter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
