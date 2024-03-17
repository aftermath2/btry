package lightning

import (
	"context"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
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

// GetInfo mock.
func (c *ClientMock) GetInfo(ctx context.Context) (*lnrpc.GetInfoResponse, error) {
	args := c.Called(ctx)
	var r0 *lnrpc.GetInfoResponse
	v0 := args.Get(0)
	if v0 != nil {
		r0 = v0.(*lnrpc.GetInfoResponse)
	}
	return r0, args.Error(1)
}

// PayInvoice mock.
func (c *ClientMock) PayInvoice(ctx context.Context, invoice *lnrpc.PayReq, feeSat int64, inflightUpdates bool) (Stream[*lnrpc.Payment], error) {
	args := c.Called(ctx, invoice, feeSat, inflightUpdates)
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

// SubscribeBlocks mock.
func (c *ClientMock) SubscribeBlocks(ctx context.Context) (Stream[*chainrpc.BlockEpoch], error) {
	args := c.Called(ctx)
	var r0 Stream[*chainrpc.BlockEpoch]
	v0 := args.Get(0)
	if v0 != nil {
		r0 = v0.(Stream[*chainrpc.BlockEpoch])
	}
	return r0, args.Error(1)
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
