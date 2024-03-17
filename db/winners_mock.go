package db

import "github.com/stretchr/testify/mock"

// WinnersStoreMock is a mocked implementation of a winners store.
type WinnersStoreMock struct {
	mock.Mock
}

// NewWinnersStoreMock returns a mocked winners store.
func NewWinnersStoreMock() *WinnersStoreMock {
	return &WinnersStoreMock{}
}

// Add mock.
func (w *WinnersStoreMock) Add(height uint32, winners []Winner) error {
	args := w.Called(height, winners)
	return args.Error(0)
}

// List mock.
func (w *WinnersStoreMock) List(height uint32) ([]Winner, error) {
	args := w.Called(height)
	var r0 []Winner
	v0 := args.Get(0)
	if v0 != nil {
		r0 = v0.([]Winner)
	}
	return r0, args.Error(1)
}
