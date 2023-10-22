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
func (w *WinnersStoreMock) Add(winners []Winner) error {
	args := w.Called(winners)
	return args.Error(0)
}

// ClaimPrizes mock.
func (w *WinnersStoreMock) ClaimPrizes(publicKey string, amount uint64) error {
	args := w.Called(publicKey, amount)
	return args.Error(0)
}

// ExpirePrizes mock.
func (w *WinnersStoreMock) ExpirePrizes() (uint64, error) {
	args := w.Called()
	prizes := args.Get(0).(uint64)
	return prizes, args.Error(1)
}

// GetPrizes mock.
func (w *WinnersStoreMock) GetPrizes(publicKey string) (uint64, error) {
	args := w.Called(publicKey)
	prizes := args.Get(0).(uint64)
	return prizes, args.Error(1)
}

// List mock.
func (w *WinnersStoreMock) List() ([]Winner, error) {
	args := w.Called()
	var r0 []Winner
	v0 := args.Get(0)
	if v0 != nil {
		r0 = v0.([]Winner)
	}
	return r0, args.Error(1)
}

// ListHistory mock.
func (w *WinnersStoreMock) ListHistory(from, to uint64) ([]Winner, error) {
	args := w.Called(from, to)
	var r0 []Winner
	v0 := args.Get(0)
	if v0 != nil {
		r0 = v0.([]Winner)
	}
	return r0, args.Error(1)
}

// WriteHistory mock.
func (w *WinnersStoreMock) WriteHistory(winners []Winner) error {
	args := w.Called(winners)
	return args.Error(0)
}
