package db

import "github.com/stretchr/testify/mock"

// NotificationsStoreMock is a mocked implementation of the Notifications store.
type NotificationsStoreMock struct {
	mock.Mock
}

// NewNotificationsStoreMock returns a mocked Notifications store.
func NewNotificationsStoreMock() *NotificationsStoreMock {
	return &NotificationsStoreMock{}
}

// Add mock.
func (n *NotificationsStoreMock) Add(publicKey string, chatID int64) error {
	args := n.Called(publicKey, chatID)
	return args.Error(0)
}

// GetChatID mock.
func (n *NotificationsStoreMock) GetChatID(publicKey string) (int64, error) {
	args := n.Called(publicKey)
	return args.Get(0).(int64), args.Error(1)
}
