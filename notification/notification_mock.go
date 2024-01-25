package notification

import (
	"github.com/aftermath2/BTRY/db"

	"github.com/stretchr/testify/mock"
)

// NotifierMock is a mocked implementation of a notifier.
type NotifierMock struct {
	mock.Mock
}

// NewNotifierMock returns a mocked notifier service.
func NewNotifierMock() *NotifierMock {
	return &NotifierMock{}
}

// GetUpdates mock.
func (n *NotifierMock) GetUpdates() {}

// Notify mock.
func (n *NotifierMock) Notify(chatID int64, message string) {
	_ = n.Called(chatID, message)
}

// PublishWinners mock.
func (n *NotifierMock) PublishWinners(winners []db.Winner) error {
	_ = n.Called(winners)
	return nil
}
