// Code generated by MockGen. DO NOT EDIT.
// Source: services/chat_service.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockChatService is a mock of ChatService interface.
type MockChatService struct {
	ctrl     *gomock.Controller
	recorder *MockChatServiceMockRecorder
}

// MockChatServiceMockRecorder is the mock recorder for MockChatService.
type MockChatServiceMockRecorder struct {
	mock *MockChatService
}

// NewMockChatService creates a new mock instance.
func NewMockChatService(ctrl *gomock.Controller) *MockChatService {
	mock := &MockChatService{ctrl: ctrl}
	mock.recorder = &MockChatServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockChatService) EXPECT() *MockChatServiceMockRecorder {
	return m.recorder
}

// AddUser mocks base method.
func (m *MockChatService) AddUser(name string) (int, chan string) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddUser", name)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(chan string)
	return ret0, ret1
}

// AddUser indicates an expected call of AddUser.
func (mr *MockChatServiceMockRecorder) AddUser(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddUser", reflect.TypeOf((*MockChatService)(nil).AddUser), name)
}

// Broadcast mocks base method.
func (m *MockChatService) Broadcast(userId int, event string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Broadcast", userId, event)
}

// Broadcast indicates an expected call of Broadcast.
func (mr *MockChatServiceMockRecorder) Broadcast(userId, event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Broadcast", reflect.TypeOf((*MockChatService)(nil).Broadcast), userId, event)
}

// IsValidName mocks base method.
func (m *MockChatService) IsValidName(name string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsValidName", name)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsValidName indicates an expected call of IsValidName.
func (mr *MockChatServiceMockRecorder) IsValidName(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsValidName", reflect.TypeOf((*MockChatService)(nil).IsValidName), name)
}

// ListCurrentUsersNames mocks base method.
func (m *MockChatService) ListCurrentUsersNames() []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListCurrentUsersNames")
	ret0, _ := ret[0].([]string)
	return ret0
}

// ListCurrentUsersNames indicates an expected call of ListCurrentUsersNames.
func (mr *MockChatServiceMockRecorder) ListCurrentUsersNames() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListCurrentUsersNames", reflect.TypeOf((*MockChatService)(nil).ListCurrentUsersNames))
}

// RemoveUser mocks base method.
func (m *MockChatService) RemoveUser(id int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RemoveUser", id)
}

// RemoveUser indicates an expected call of RemoveUser.
func (mr *MockChatServiceMockRecorder) RemoveUser(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveUser", reflect.TypeOf((*MockChatService)(nil).RemoveUser), id)
}
