// Code generated by MockGen. DO NOT EDIT.
// Source: services/unusual_db_service.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockUnusualDbService is a mock of UnusualDbService interface.
type MockUnusualDbService struct {
	ctrl     *gomock.Controller
	recorder *MockUnusualDbServiceMockRecorder
}

// MockUnusualDbServiceMockRecorder is the mock recorder for MockUnusualDbService.
type MockUnusualDbServiceMockRecorder struct {
	mock *MockUnusualDbService
}

// NewMockUnusualDbService creates a new mock instance.
func NewMockUnusualDbService(ctrl *gomock.Controller) *MockUnusualDbService {
	mock := &MockUnusualDbService{ctrl: ctrl}
	mock.recorder = &MockUnusualDbServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUnusualDbService) EXPECT() *MockUnusualDbServiceMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockUnusualDbService) Get(key string) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", key)
	ret0, _ := ret[0].(string)
	return ret0
}

// Get indicates an expected call of Get.
func (mr *MockUnusualDbServiceMockRecorder) Get(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockUnusualDbService)(nil).Get), key)
}

// Set mocks base method.
func (m *MockUnusualDbService) Set(key, value string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Set", key, value)
}

// Set indicates an expected call of Set.
func (mr *MockUnusualDbServiceMockRecorder) Set(key, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockUnusualDbService)(nil).Set), key, value)
}