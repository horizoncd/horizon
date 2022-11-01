// Code generated by MockGen. DO NOT EDIT.
// Source: registry.go

// Package mock_registry is a generated GoMock package.
package mock_registry

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockRegistry is a mock of Registry interface.
type MockRegistry struct {
	ctrl     *gomock.Controller
	recorder *MockRegistryMockRecorder
}

// MockRegistryMockRecorder is the mock recorder for MockRegistry.
type MockRegistryMockRecorder struct {
	mock *MockRegistry
}

// NewMockRegistry creates a new mock instance.
func NewMockRegistry(ctrl *gomock.Controller) *MockRegistry {
	mock := &MockRegistry{ctrl: ctrl}
	mock.recorder = &MockRegistryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRegistry) EXPECT() *MockRegistryMockRecorder {
	return m.recorder
}

// DeleteRepository mocks base method.
func (m *MockRegistry) DeleteRepository(ctx context.Context, names ...string) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range names {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteRepository", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteRepository indicates an expected call of DeleteRepository.
func (mr *MockRegistryMockRecorder) DeleteRepository(ctx interface{}, names ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, names...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteRepository", reflect.TypeOf((*MockRegistry)(nil).DeleteRepository), varargs...)
}
