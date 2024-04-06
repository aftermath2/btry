package db

import "github.com/stretchr/testify/mock"

// LightningStoreMock is a mocked implementation of a lightning store.
type LightningStoreMock struct {
	mock.Mock
}

// NewLightningStoreMock returns a mocked lightning store.
func NewLightningStoreMock() *LightningStoreMock {
	return &LightningStoreMock{}
}

// Get mock.
func (l *LightningStoreMock) GetAddress(publicKey string) (string, error) {
	args := l.Called(publicKey)
	var r0 string
	v0 := args.Get(0)
	if v0 != nil {
		r0 = v0.(string)
	}
	return r0, args.Error(1)
}

// Set mock.
func (l *LightningStoreMock) SetAddress(publicKey, address string) error {
	args := l.Called(publicKey, address)
	return args.Error(0)
}
