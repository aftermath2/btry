package db

import "github.com/stretchr/testify/mock"

// PrizesStoreMock is a mocked implementation of a prizes store.
type PrizesStoreMock struct {
	mock.Mock
}

// NewPrizesStoreMock returns a mocked prizes store.
func NewPrizesStoreMock() *PrizesStoreMock {
	return &PrizesStoreMock{}
}

// Expire mock.
func (w *PrizesStoreMock) Expire(height uint32) (uint64, error) {
	args := w.Called(height)
	return args.Get(0).(uint64), args.Error(1)
}

// Get mock.
func (w *PrizesStoreMock) Get(publicKey string) (uint64, error) {
	args := w.Called(publicKey)
	return args.Get(0).(uint64), args.Error(1)
}

// Set mock.
func (w *PrizesStoreMock) Set(lotteryHeight uint32, winners []Winner) error {
	args := w.Called(lotteryHeight, winners)
	return args.Error(0)
}

// Withdraw mock.
func (w *PrizesStoreMock) Withdraw(publicKey string, amount uint64) error {
	args := w.Called(publicKey, amount)
	return args.Error(0)
}
