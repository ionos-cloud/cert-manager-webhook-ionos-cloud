// Code generated by mockery v2.53.2. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"

	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// K8Client is an autogenerated mock type for the K8Client type
type K8Client struct {
	mock.Mock
}

type K8Client_Expecter struct {
	mock *mock.Mock
}

func (_m *K8Client) EXPECT() *K8Client_Expecter {
	return &K8Client_Expecter{mock: &_m.Mock}
}

// CoreV1 provides a mock function with no fields
func (_m *K8Client) CoreV1() v1.CoreV1Interface {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for CoreV1")
	}

	var r0 v1.CoreV1Interface
	if rf, ok := ret.Get(0).(func() v1.CoreV1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1.CoreV1Interface)
		}
	}

	return r0
}

// K8Client_CoreV1_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CoreV1'
type K8Client_CoreV1_Call struct {
	*mock.Call
}

// CoreV1 is a helper method to define mock.On call
func (_e *K8Client_Expecter) CoreV1() *K8Client_CoreV1_Call {
	return &K8Client_CoreV1_Call{Call: _e.mock.On("CoreV1")}
}

func (_c *K8Client_CoreV1_Call) Run(run func()) *K8Client_CoreV1_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *K8Client_CoreV1_Call) Return(_a0 v1.CoreV1Interface) *K8Client_CoreV1_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *K8Client_CoreV1_Call) RunAndReturn(run func() v1.CoreV1Interface) *K8Client_CoreV1_Call {
	_c.Call.Return(run)
	return _c
}

// NewK8Client creates a new instance of K8Client. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewK8Client(t interface {
	mock.TestingT
	Cleanup(func())
}) *K8Client {
	mock := &K8Client{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
