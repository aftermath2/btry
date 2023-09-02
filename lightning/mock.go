package lightning

import (
	"context"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/mock"
)

// BlockedStreamMock is a stream of events that blocks to let the test execution end without failures.
type BlockedStreamMock[T any] struct{}

// Recv blocks forever.
func (s BlockedStreamMock[T]) Recv() (T, error) {
	// Block execution to let tests run
	block := make(chan struct{})
	<-block
	var v T
	return v, nil
}

// ClientMock is a mocked implementation of a lightning client.
type ClientMock struct {
	mock.Mock
}

// NewClientMock returns a mocked lightning network node client.
func NewClientMock() *ClientMock {
	return &ClientMock{}
}

// AddHODLInvoice mock.
func (c *ClientMock) AddHODLInvoice(ctx context.Context, invoice *lnrpc.PayReq) (string, error) {
	args := c.Called(ctx, invoice)
	paymentRequest := args.Get(0).(string)
	return paymentRequest, args.Error(1)
}

// AddInvoice mock.
func (c *ClientMock) AddInvoice(ctx context.Context, amount uint64) (*lnrpc.AddInvoiceResponse, error) {
	args := c.Called(ctx, amount)
	var r0 *lnrpc.AddInvoiceResponse
	v0 := args.Get(0)
	if v0 != nil {
		r0 = v0.(*lnrpc.AddInvoiceResponse)
	}
	return r0, args.Error(1)
}

// CancelInvoice mock.
func (c *ClientMock) CancelInvoice(ctx context.Context, rHash string) error {
	args := c.Called(ctx, rHash)
	return args.Error(0)
}

// DecodeInvoice mock.
func (c *ClientMock) DecodeInvoice(ctx context.Context, invoice string) (*lnrpc.PayReq, error) {
	args := c.Called(ctx, invoice)
	var r0 *lnrpc.PayReq
	v0 := args.Get(0)
	if v0 != nil {
		r0 = v0.(*lnrpc.PayReq)
	}
	return r0, args.Error(1)
}

// LightningAddresses mock.
func (c *ClientMock) LightningAddresses() []string {
	args := c.Called()
	addresses := args.Get(0).([]string)
	return addresses
}

// PayInvoice mock.
func (c *ClientMock) PayInvoice(ctx context.Context, invoice *lnrpc.PayReq, feeSat int64, cltvLimit int32) (Stream[*lnrpc.Payment], error) {
	args := c.Called(ctx, invoice, feeSat, cltvLimit)
	var r0 Stream[*lnrpc.Payment]
	v0 := args.Get(0)
	if v0 != nil {
		r0 = v0.(Stream[*lnrpc.Payment])
	}
	return r0, args.Error(1)
}

// RemoteBalance mock.
func (c *ClientMock) RemoteBalance(ctx context.Context) (int64, error) {
	args := c.Called(ctx)
	remoteBalance := args.Get(0).(int64)
	return remoteBalance, args.Error(1)
}

// SettleInvoice mock.
func (c *ClientMock) SettleInvoice(ctx context.Context, originalInvoice string) error {
	args := c.Called(ctx, originalInvoice)
	return args.Error(0)
}

// SubscribeChannelEvents mock.
func (c *ClientMock) SubscribeChannelEvents(ctx context.Context) (Stream[*lnrpc.ChannelEventUpdate], error) {
	args := c.Called(ctx)
	var r0 Stream[*lnrpc.ChannelEventUpdate]
	v0 := args.Get(0)
	if v0 != nil {
		r0 = v0.(Stream[*lnrpc.ChannelEventUpdate])
	}
	return r0, args.Error(1)
}

// SubscribeInvoices mock.
func (c *ClientMock) SubscribeInvoices(ctx context.Context) (Stream[*lnrpc.Invoice], error) {
	args := c.Called(ctx)
	var r0 Stream[*lnrpc.Invoice]
	v0 := args.Get(0)
	if v0 != nil {
		r0 = v0.(Stream[*lnrpc.Invoice])
	}
	return r0, args.Error(1)
}

// SubscribePayments mock.
func (c *ClientMock) SubscribePayments(ctx context.Context) (Stream[*lnrpc.Payment], error) {
	args := c.Called(ctx)
	var r0 Stream[*lnrpc.Payment]
	v0 := args.Get(0)
	if v0 != nil {
		r0 = v0.(Stream[*lnrpc.Payment])
	}
	return r0, args.Error(1)
}
