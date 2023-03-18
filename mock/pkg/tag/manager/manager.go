// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/tag/manager/manager.go

// Package mock_manager is a generated GoMock package.
package mock_manager

import (
	context "context"
	reflect "reflect"

	models "github.com/horizoncd/horizon/pkg/tag/models"
	gomock "github.com/golang/mock/gomock"
)

// MockManager is a mock of Manager interface.
type MockManager struct {
	ctrl     *gomock.Controller
	recorder *MockManagerMockRecorder
}

// MockManagerMockRecorder is the mock recorder for MockManager.
type MockManagerMockRecorder struct {
	mock *MockManager
}

// NewMockManager creates a new mock instance.
func NewMockManager(ctrl *gomock.Controller) *MockManager {
	mock := &MockManager{ctrl: ctrl}
	mock.recorder = &MockManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockManager) EXPECT() *MockManagerMockRecorder {
	return m.recorder
}

// ListByResourceTypeID mocks base method.
func (m *MockManager) ListByResourceTypeID(ctx context.Context, resourceType string, resourceID uint) ([]*models.Tag, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListByResourceTypeID", ctx, resourceType, resourceID)
	ret0, _ := ret[0].([]*models.Tag)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListByResourceTypeID indicates an expected call of ListByResourceTypeID.
func (mr *MockManagerMockRecorder) ListByResourceTypeID(ctx, resourceType, resourceID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByResourceTypeID", reflect.TypeOf((*MockManager)(nil).ListByResourceTypeID), ctx, resourceType, resourceID)
}

// ListByResourceTypeIDs mocks base method.
func (m *MockManager) ListByResourceTypeIDs(ctx context.Context, resourceType string, resourceIDs []uint, deduplicate bool) ([]*models.Tag, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListByResourceTypeIDs", ctx, resourceType, resourceIDs, deduplicate)
	ret0, _ := ret[0].([]*models.Tag)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListByResourceTypeIDs indicates an expected call of ListByResourceTypeIDs.
func (mr *MockManagerMockRecorder) ListByResourceTypeIDs(ctx, resourceType, resourceIDs, deduplicate interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByResourceTypeIDs", reflect.TypeOf((*MockManager)(nil).ListByResourceTypeIDs), ctx, resourceType, resourceIDs, deduplicate)
}

// UpsertByResourceTypeID mocks base method.
func (m *MockManager) UpsertByResourceTypeID(ctx context.Context, resourceType string, resourceID uint, tags []*models.Tag) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertByResourceTypeID", ctx, resourceType, resourceID, tags)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertByResourceTypeID indicates an expected call of UpsertByResourceTypeID.
func (mr *MockManagerMockRecorder) UpsertByResourceTypeID(ctx, resourceType, resourceID, tags interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertByResourceTypeID", reflect.TypeOf((*MockManager)(nil).UpsertByResourceTypeID), ctx, resourceType, resourceID, tags)
}
