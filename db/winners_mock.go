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

// ClaimPrizes mock.
func (w *WinnersStoreMock) ClaimPrizes(publicKey string, amount uint64) error {
	args := w.Called(publicKey, amount)
	return args.Error(0)
}

// ExpirePrizes mock.
func (w *WinnersStoreMock) ExpirePrizes(height uint32) (uint64, error) {
	args := w.Called(height)
	return args.Get(0).(uint64), args.Error(1)
}

// GetPrizes mock.
func (w *WinnersStoreMock) GetPrizes(publicKey string) (uint64, error) {
	args := w.Called(publicKey)
	return args.Get(0).(uint64), args.Error(1)
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
