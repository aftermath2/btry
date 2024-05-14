// Package lightning provides utilities for interacting with a Lightning Netowrk node.
package lightning

import (
	"context"
	"encoding/hex"
	"net/http"
	"os"
	"time"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/logger"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/macaroons"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	"gopkg.in/macaroon.v2"
)

// DefaultInvoiceExpiry is the default time used for invoices expiration.
const DefaultInvoiceExpiry = time.Hour * 3

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
	GetInfo(ctx context.Context) (*lnrpc.GetInfoResponse, error)
	PayInvoice(ctx context.Context, invoice *lnrpc.PayReq, feeSat int64, inflightUpdates bool) (Stream[*lnrpc.Payment], error)
	RemoteBalance(ctx context.Context) (int64, error)
	SendToLightningAddress(ctx context.Context, address string, amountSat int64) (string, error)
	SubscribeBlocks(ctx context.Context) (Stream[*chainrpc.BlockEpoch], error)
	SubscribeChannelEvents(ctx context.Context) (Stream[*lnrpc.ChannelEventUpdate], error)
	SubscribeInvoices(ctx context.Context) (Stream[*lnrpc.Invoice], error)
	SubscribePayments(ctx context.Context) (Stream[*lnrpc.Payment], error)
}

type client struct {
	ln        lnrpc.LightningClient
	chain     chainrpc.ChainNotifierClient
	router    routerrpc.RouterClient
	logger    *logger.Logger
	torClient *http.Client
	maxFeePPM int64
}

// NewClient returns a new client that communicates with a Lightning node.
func NewClient(config config.Lightning, torClient *http.Client) (Client, error) {
	logger, err := logger.New(config.Logger)
	if err != nil {
		return nil, err
	}

	opts, err := loadGRPCOpts(config)
	if err != nil {
		return nil, errors.Wrap(err, "loading gRPC options")
	}

	logger.Infof("Opening gRPC connection to %s...", config.RPCAddress)

	conn, err := grpc.NewClient(config.RPCAddress, opts...)
	if err != nil {
		return nil, err
	}

	if err := waitForLND(conn, logger); err != nil {
		return nil, err
	}

	return &client{
		ln:        lnrpc.NewLightningClient(conn),
		chain:     chainrpc.NewChainNotifierClient(conn),
		router:    routerrpc.NewRouterClient(conn),
		logger:    logger,
		torClient: torClient,
		maxFeePPM: config.MaxFeePPM,
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

	connectionParams := grpc.ConnectParams{
		Backoff: backoff.Config{
			BaseDelay:  5 * time.Second,
			Multiplier: 1.5,
			MaxDelay:   time.Minute,
		},
		MinConnectTimeout: 30 * time.Second,
	}

	return []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCred),
		grpc.WithPerRPCCredentials(macCred),
		grpc.WithConnectParams(connectionParams),
	}, nil
}

// waitForLND blocks the execution until LND is fully ready to accept calls.
func waitForLND(conn *grpc.ClientConn, logger *logger.Logger) error {
	stateClient := lnrpc.NewStateClient(conn)
	stream, err := stateClient.SubscribeState(context.Background(), &lnrpc.SubscribeStateRequest{})
	if err != nil {
		return errors.Wrap(err, "subscribing to state")
	}

	logger.Info("Waiting for LND to be ready to accept connections...")

	for {
		wallet, err := stream.Recv()
		if err != nil {
			return err
		}

		if wallet.State == lnrpc.WalletState_SERVER_ACTIVE {
			logger.Info("The LND server is now active. Initializing connections...")
			return nil
		}

		logger.Infof("Wallet state changed to: %s", wallet.State)
	}
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

// GetInfo returns general information concerning the lightning node.
func (c *client) GetInfo(ctx context.Context) (*lnrpc.GetInfoResponse, error) {
	return c.ln.GetInfo(ctx, &lnrpc.GetInfoRequest{})
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

// PayInvoice attempts to route a payment to the final destination.
func (c *client) PayInvoiceSync(
	ctx context.Context,
	invoice *lnrpc.PayReq,
	feeSat int64,
) (*lnrpc.SendResponse, error) {
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

	req := &lnrpc.SendRequest{
		Amt: invoice.NumSatoshis,
		FeeLimit: &lnrpc.FeeLimit{
			Limit: &lnrpc.FeeLimit_Fixed{
				Fixed: feeSat,
			},
		},
		Dest:           dest,
		PaymentHash:    paymentHash,
		PaymentAddr:    invoice.PaymentAddr[:],
		FinalCltvDelta: 80,
	}
	return c.ln.SendPaymentSync(ctx, req)
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

// SendToLightningAddress uses the LNURL protocol to request invoices based on the address provided
// and it pays them. It returns the payment preimage or an error if it fails.
func (c *client) SendToLightningAddress(ctx context.Context, address string, amountSat int64) (string, error) {
	callback, err := getPayCallback(c.torClient, address, amountSat)
	if err != nil {
		return "", err
	}

	invoice, err := getInvoice(c.torClient, callback)
	if err != nil {
		return "", err
	}

	payReq, err := c.DecodeInvoice(ctx, invoice)
	if err != nil {
		return "", err
	}

	if payReq.NumSatoshis != amountSat {
		return "", errors.New("invalid invoice amount")
	}

	fee := amountSat * c.maxFeePPM / 1_000_000
	resp, err := c.PayInvoiceSync(ctx, payReq, fee)
	if err != nil {
		return "", err
	}

	if resp.PaymentError != "" {
		return "", errors.New(resp.PaymentError)
	}

	return hex.EncodeToString(resp.PaymentPreimage), nil
}

// SubscribeBlocks creates a uni-directional stream from the server to the client in which
// any updates relevant to new blocks are sent over.
func (c *client) SubscribeBlocks(ctx context.Context) (Stream[*chainrpc.BlockEpoch], error) {
	return c.chain.RegisterBlockEpochNtfn(ctx, nil)
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
