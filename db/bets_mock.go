package db

import "github.com/stretchr/testify/mock"

// BetsStoreMock is a mocked implementation of the bets store.
type BetsStoreMock struct {
	mock.Mock
}

// NewBetsStoreMock returns a mocked bets store.
func NewBetsStoreMock() *BetsStoreMock {
	return &BetsStoreMock{}
}

// Add mock.
func (b *BetsStoreMock) Add(bet Bet) error {
	args := b.Called(bet)
	return args.Error(0)
}

// GetPrizePool mock.
func (b *BetsStoreMock) GetPrizePool() (uint64, error) {
	args := b.Called()
	return args.Get(0).(uint64), args.Error(1)
}

// List mock.
func (b *BetsStoreMock) List(offset, limit uint64, reverse bool) ([]Bet, error) {
	args := b.Called(offset, limit, reverse)
	return args.Get(0).([]Bet), args.Error(1)
}

// Reset mock.
func (b *BetsStoreMock) Reset() error {
	args := b.Called()
	return args.Error(0)
}
