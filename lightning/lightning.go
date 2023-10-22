// Package lightning provides utilities for interacting with a Lightning Netowrk node.
package lightning

import (
	"context"
	"encoding/hex"
	"os"
	"time"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/logger"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/macaroons"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/macaroon.v2"
)

const (
	// DefaultInvoiceExpiry is the default time used for invoices expiration.
	DefaultInvoiceExpiry = time.Hour * 3

	// LNURL invoice description hash
	emptySHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

// TODO:
// - Accept receiving and sending via on-chain
// - Open channels programmatically

// Stream implements a method that receives updates from a stream.
type Stream[T any] interface {
	Recv() (T, error)
}

// Client represents a Lightning Network node client.
type Client interface {
	AddInvoice(ctx context.Context, amountSat uint64) (*lnrpc.AddInvoiceResponse, error)
	DecodeInvoice(ctx context.Context, invoice string) (*lnrpc.PayReq, error)
	PayInvoice(ctx context.Context, invoice *lnrpc.PayReq, feeSat int64, inflightUpdates bool) (Stream[*lnrpc.Payment], error)
	RemoteBalance(ctx context.Context) (int64, error)
	SubscribeChannelEvents(ctx context.Context) (Stream[*lnrpc.ChannelEventUpdate], error)
	SubscribeInvoices(ctx context.Context) (Stream[*lnrpc.Invoice], error)
	SubscribePayments(ctx context.Context) (Stream[*lnrpc.Payment], error)
}

type client struct {
	ln     lnrpc.LightningClient
	router routerrpc.RouterClient
	logger *logger.Logger
}

// NewClient returns a new client that communicates with a Lightning node.
func NewClient(config config.Lightning) (Client, error) {
	opts, err := loadGRPCOpts(config)
	if err != nil {
		return nil, errors.Wrap(err, "loading grpc options")
	}

	conn, err := grpc.Dial(config.RPCAddress, opts...)
	if err != nil {
		return nil, err
	}

	logger, err := logger.New(config.Logger)
	if err != nil {
		return nil, err
	}

	return &client{
		ln:     lnrpc.NewLightningClient(conn),
		router: routerrpc.NewRouterClient(conn),
		logger: logger,
	}, nil
}

func loadGRPCOpts(config config.Lightning) ([]grpc.DialOption, error) {
	tlsCred, err := credentials.NewClientTLSFromFile(config.TLSCertPath, "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to read TLS certificate")
	}

	macBytes, err := os.ReadFile(config.MacaroonPath)
	if err != nil {
		return nil, errors.Wrap(err, "reading macaroon file")
	}

	mac := &macaroon.Macaroon{}
	if err := mac.UnmarshalBinary(macBytes); err != nil {
		return nil, errors.Wrap(err, "unmarshaling macaroon")
	}

	macCred, err := macaroons.NewMacaroonCredential(mac)
	if err != nil {
		return nil, errors.Wrap(err, "creating macaroon credential")
	}

	return []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCred),
		grpc.WithPerRPCCredentials(macCred),
	}, nil
}

// AddInvoice attempts to add a new invoice to the invoice database.
func (c *client) AddInvoice(ctx context.Context, amountSat uint64) (*lnrpc.AddInvoiceResponse, error) {
	invoice := &lnrpc.Invoice{
		Memo:    "BTRY",
		Value:   int64(amountSat),
		Expiry:  int64(DefaultInvoiceExpiry.Seconds()),
		Private: false,
	}
	return c.ln.AddInvoice(ctx, invoice)
}

// DecodeInvoice parses the provided encoded invoice and returns a decoded Invoice if it is valid by
// BOLT-0011 and matches the provided active network.
func (c *client) DecodeInvoice(ctx context.Context, invoice string) (*lnrpc.PayReq, error) {
	return c.ln.DecodePayReq(ctx, &lnrpc.PayReqString{PayReq: invoice})
}

// PayInvoice attempts to route a payment to the final destination.
func (c *client) PayInvoice(
	ctx context.Context,
	invoice *lnrpc.PayReq,
	feeSat int64,
	inflightUpdates bool,
) (Stream[*lnrpc.Payment], error) {
	if feeSat < 0 {
		return nil, errors.New("invalid fee")
	}

	dest, err := hex.DecodeString(invoice.Destination)
	if err != nil {
		return nil, errors.Wrap(err, "decoding destination")
	}

	paymentHash, err := hex.DecodeString(invoice.PaymentHash)
	if err != nil {
		return nil, errors.Wrap(err, "decoding payment hash")
	}

	req := &routerrpc.SendPaymentRequest{
		Amt:               invoice.NumSatoshis,
		FeeLimitSat:       feeSat,
		Dest:              dest,
		PaymentHash:       paymentHash,
		PaymentAddr:       invoice.PaymentAddr[:],
		RouteHints:        invoice.RouteHints,
		MaxParts:          16,
		NoInflightUpdates: !inflightUpdates,
		TimePref:          0.5,
		FinalCltvDelta:    80,
		TimeoutSeconds:    120,
	}
	return c.router.SendPaymentV2(ctx, req)
}

// RemoteBalance returns a report on the total remote funds across all open public channels.
func (c *client) RemoteBalance(ctx context.Context) (int64, error) {
	resp, err := c.ln.ListChannels(ctx, &lnrpc.ListChannelsRequest{
		ActiveOnly: true,
		PublicOnly: true,
	})
	if err != nil {
		return 0, errors.Wrap(err, "listing channels")
	}

	remoteBalance := int64(0)
	for _, ch := range resp.Channels {
		remoteBalance += ch.RemoteBalance
	}

	return remoteBalance, nil
}

// SubscribeChannelEvents creates a uni-directional stream from the server to the client in which
// any updates relevant to the state of the channels are sent over.
//
// Events include new active channels, inactive channels, and closed channels.
func (c *client) SubscribeChannelEvents(ctx context.Context) (Stream[*lnrpc.ChannelEventUpdate], error) {
	return c.ln.SubscribeChannelEvents(ctx, &lnrpc.ChannelEventSubscription{})
}

// SubscribeInvoices returns a uni-directional stream (server -> client) for notifying the client of
// newly added/settled invoices.
func (c *client) SubscribeInvoices(ctx context.Context) (Stream[*lnrpc.Invoice], error) {
	return c.ln.SubscribeInvoices(ctx, &lnrpc.InvoiceSubscription{})
}

// SubscribePayments returns an update stream for every payment that is not in a terminal state.
func (c *client) SubscribePayments(ctx context.Context) (Stream[*lnrpc.Payment], error) {
	return c.router.TrackPayments(ctx, &routerrpc.TrackPaymentsRequest{NoInflightUpdates: true})
}
