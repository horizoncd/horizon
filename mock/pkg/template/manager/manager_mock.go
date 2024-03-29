// Code generated by MockGen. DO NOT EDIT.
// Source: manager.go

// Package mock_manager is a generated GoMock package.
package mock_manager

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	q "github.com/horizoncd/horizon/lib/q"
	models "github.com/horizoncd/horizon/pkg/application/models"
	models0 "github.com/horizoncd/horizon/pkg/cluster/models"
	models1 "github.com/horizoncd/horizon/pkg/template/models"
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

// Create mocks base method.
func (m *MockManager) Create(ctx context.Context, template *models1.Template) (*models1.Template, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, template)
	ret0, _ := ret[0].(*models1.Template)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockManagerMockRecorder) Create(ctx, template interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockManager)(nil).Create), ctx, template)
}

// DeleteByID mocks base method.
func (m *MockManager) DeleteByID(ctx context.Context, id uint) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteByID", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteByID indicates an expected call of DeleteByID.
func (mr *MockManagerMockRecorder) DeleteByID(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteByID", reflect.TypeOf((*MockManager)(nil).DeleteByID), ctx, id)
}

// GetByID mocks base method.
func (m *MockManager) GetByID(ctx context.Context, id uint) (*models1.Template, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*models1.Template)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockManagerMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockManager)(nil).GetByID), ctx, id)
}

// GetByName mocks base method.
func (m *MockManager) GetByName(ctx context.Context, name string) (*models1.Template, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByName", ctx, name)
	ret0, _ := ret[0].(*models1.Template)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByName indicates an expected call of GetByName.
func (mr *MockManagerMockRecorder) GetByName(ctx, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByName", reflect.TypeOf((*MockManager)(nil).GetByName), ctx, name)
}

// GetRefOfApplication mocks base method.
func (m *MockManager) GetRefOfApplication(ctx context.Context, id uint) ([]*models.Application, uint, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRefOfApplication", ctx, id)
	ret0, _ := ret[0].([]*models.Application)
	ret1, _ := ret[1].(uint)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetRefOfApplication indicates an expected call of GetRefOfApplication.
func (mr *MockManagerMockRecorder) GetRefOfApplication(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRefOfApplication", reflect.TypeOf((*MockManager)(nil).GetRefOfApplication), ctx, id)
}

// GetRefOfCluster mocks base method.
func (m *MockManager) GetRefOfCluster(ctx context.Context, id uint) ([]*models0.Cluster, uint, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRefOfCluster", ctx, id)
	ret0, _ := ret[0].([]*models0.Cluster)
	ret1, _ := ret[1].(uint)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetRefOfCluster indicates an expected call of GetRefOfCluster.
func (mr *MockManagerMockRecorder) GetRefOfCluster(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRefOfCluster", reflect.TypeOf((*MockManager)(nil).GetRefOfCluster), ctx, id)
}

// ListByGroupID mocks base method.
func (m *MockManager) ListByGroupID(ctx context.Context, groupID uint) ([]*models1.Template, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListByGroupID", ctx, groupID)
	ret0, _ := ret[0].([]*models1.Template)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListByGroupID indicates an expected call of ListByGroupID.
func (mr *MockManagerMockRecorder) ListByGroupID(ctx, groupID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByGroupID", reflect.TypeOf((*MockManager)(nil).ListByGroupID), ctx, groupID)
}

// ListByGroupIDs mocks base method.
func (m *MockManager) ListByGroupIDs(ctx context.Context, ids []uint) ([]*models1.Template, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListByGroupIDs", ctx, ids)
	ret0, _ := ret[0].([]*models1.Template)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListByGroupIDs indicates an expected call of ListByGroupIDs.
func (mr *MockManagerMockRecorder) ListByGroupIDs(ctx, ids interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByGroupIDs", reflect.TypeOf((*MockManager)(nil).ListByGroupIDs), ctx, ids)
}

// ListByIDs mocks base method.
func (m *MockManager) ListByIDs(ctx context.Context, ids []uint) ([]*models1.Template, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListByIDs", ctx, ids)
	ret0, _ := ret[0].([]*models1.Template)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListByIDs indicates an expected call of ListByIDs.
func (mr *MockManagerMockRecorder) ListByIDs(ctx, ids interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByIDs", reflect.TypeOf((*MockManager)(nil).ListByIDs), ctx, ids)
}

// ListTemplate mocks base method.
func (m *MockManager) ListTemplate(ctx context.Context) ([]*models1.Template, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListTemplate", ctx)
	ret0, _ := ret[0].([]*models1.Template)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListTemplate indicates an expected call of ListTemplate.
func (mr *MockManagerMockRecorder) ListTemplate(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListTemplate", reflect.TypeOf((*MockManager)(nil).ListTemplate), ctx)
}

// ListV2 mocks base method.
func (m *MockManager) ListV2(ctx context.Context, query *q.Query, groupIDs ...uint) ([]*models1.Template, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, query}
	for _, a := range groupIDs {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListV2", varargs...)
	ret0, _ := ret[0].([]*models1.Template)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListV2 indicates an expected call of ListV2.
func (mr *MockManagerMockRecorder) ListV2(ctx, query interface{}, groupIDs ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, query}, groupIDs...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListV2", reflect.TypeOf((*MockManager)(nil).ListV2), varargs...)
}

// UpdateByID mocks base method.
func (m *MockManager) UpdateByID(ctx context.Context, id uint, template *models1.Template) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateByID", ctx, id, template)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateByID indicates an expected call of UpdateByID.
func (mr *MockManagerMockRecorder) UpdateByID(ctx, id, template interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateByID", reflect.TypeOf((*MockManager)(nil).UpdateByID), ctx, id, template)
}
