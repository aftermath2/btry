package db

import "github.com/stretchr/testify/mock"

// LotteriesStoreMock is a mocked implementation of the Lottery store.
type LotteriesStoreMock struct {
	mock.Mock
}

// NewLotteriesStoreMock returns a mocked Lottery store.
func NewLotteriesStoreMock() *LotteriesStoreMock {
	return &LotteriesStoreMock{}
}

// AddHeight mock.
func (l *LotteriesStoreMock) AddHeight(height uint32) error {
	args := l.Called(height)
	return args.Error(0)
}

// DeleteHeight mock.
func (l *LotteriesStoreMock) DeleteHeight(height uint32) error {
	args := l.Called(height)
	return args.Error(0)
}

// GetNextHeight mock.
func (l *LotteriesStoreMock) GetNextHeight() (uint32, error) {
	args := l.Called()
	return args.Get(0).(uint32), args.Error(1)
}

// ListHeights mock.
func (l *LotteriesStoreMock) ListHeights(offset, limit uint64, reverse bool) ([]uint32, error) {
	args := l.Called(offset, limit, reverse)
	return args.Get(0).([]uint32), args.Error(1)
}
