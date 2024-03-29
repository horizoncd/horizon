// Code generated by MockGen. DO NOT EDIT.
// Source: cd.go

// Package mock_cd is a generated GoMock package.
package mock_cd

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	cd "github.com/horizoncd/horizon/pkg/cd"
)

// MockCD is a mock of CD interface.
type MockCD struct {
	ctrl     *gomock.Controller
	recorder *MockCDMockRecorder
}

// MockCDMockRecorder is the mock recorder for MockCD.
type MockCDMockRecorder struct {
	mock *MockCD
}

// NewMockCD creates a new mock instance.
func NewMockCD(ctrl *gomock.Controller) *MockCD {
	mock := &MockCD{ctrl: ctrl}
	mock.recorder = &MockCDMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCD) EXPECT() *MockCDMockRecorder {
	return m.recorder
}

// CreateCluster mocks base method.
func (m *MockCD) CreateCluster(ctx context.Context, params *cd.CreateClusterParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateCluster", ctx, params)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateCluster indicates an expected call of CreateCluster.
func (mr *MockCDMockRecorder) CreateCluster(ctx, params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCluster", reflect.TypeOf((*MockCD)(nil).CreateCluster), ctx, params)
}

// DeleteCluster mocks base method.
func (m *MockCD) DeleteCluster(ctx context.Context, params *cd.DeleteClusterParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteCluster", ctx, params)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteCluster indicates an expected call of DeleteCluster.
func (mr *MockCDMockRecorder) DeleteCluster(ctx, params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCluster", reflect.TypeOf((*MockCD)(nil).DeleteCluster), ctx, params)
}

// DeployCluster mocks base method.
func (m *MockCD) DeployCluster(ctx context.Context, params *cd.DeployClusterParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeployCluster", ctx, params)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeployCluster indicates an expected call of DeployCluster.
func (mr *MockCDMockRecorder) DeployCluster(ctx, params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeployCluster", reflect.TypeOf((*MockCD)(nil).DeployCluster), ctx, params)
}

// GetClusterState mocks base method.
func (m *MockCD) GetClusterState(ctx context.Context, params *cd.GetClusterStateV2Params) (*cd.ClusterStateV2, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClusterState", ctx, params)
	ret0, _ := ret[0].(*cd.ClusterStateV2)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetClusterState indicates an expected call of GetClusterState.
func (mr *MockCDMockRecorder) GetClusterState(ctx, params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClusterState", reflect.TypeOf((*MockCD)(nil).GetClusterState), ctx, params)
}

// GetPodEvents mocks base method.
func (m *MockCD) GetPodEvents(ctx context.Context, params *cd.GetPodEventsParams) ([]cd.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPodEvents", ctx, params)
	ret0, _ := ret[0].([]cd.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPodEvents indicates an expected call of GetPodEvents.
func (mr *MockCDMockRecorder) GetPodEvents(ctx, params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPodEvents", reflect.TypeOf((*MockCD)(nil).GetPodEvents), ctx, params)
}

// GetResourceTree mocks base method.
func (m *MockCD) GetResourceTree(ctx context.Context, params *cd.GetResourceTreeParams) ([]cd.ResourceNode, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetResourceTree", ctx, params)
	ret0, _ := ret[0].([]cd.ResourceNode)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetResourceTree indicates an expected call of GetResourceTree.
func (mr *MockCDMockRecorder) GetResourceTree(ctx, params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetResourceTree", reflect.TypeOf((*MockCD)(nil).GetResourceTree), ctx, params)
}

// GetStep mocks base method.
func (m *MockCD) GetStep(ctx context.Context, params *cd.GetStepParams) (*cd.Step, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStep", ctx, params)
	ret0, _ := ret[0].(*cd.Step)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStep indicates an expected call of GetStep.
func (mr *MockCDMockRecorder) GetStep(ctx, params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStep", reflect.TypeOf((*MockCD)(nil).GetStep), ctx, params)
}
