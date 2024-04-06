// Package sse contains utilities for working with server sent events.
package sse

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/lightning"
	"github.com/aftermath2/BTRY/logger"
	"github.com/aftermath2/BTRY/lottery"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/pkg/errors"
	"github.com/r3labs/sse"
)

const (
	failed status = iota
	success

	streamID       = "events"
	expirationTime = time.Hour * 24
)

var (
	infoEvent     = []byte("info")
	invoicesEvent = []byte("invoices")
	paymentsEvent = []byte("payments")
)

type status uint8

type infoPayload struct {
	Winners    *[]db.Winner `json:"winners,omitempty"`
	Capacity   *int64       `json:"capacity,omitempty"`
	PrizePool  *int64       `json:"prize_pool,omitempty"`
	NextHeight *uint32      `json:"next_height,omitempty"`
}

type invoicesPayload struct {
	Error     string `json:"error,omitempty"`
	PublicKey string `json:"public_key,omitempty"`
	PaymentID uint64 `json:"payment_id,omitempty"`
	Amount    uint64 `json:"amount,omitempty"`
	Status    status `json:"status,omitempty"`
}

type paymentsPayload struct {
	Error     string `json:"error,omitempty"`
	PaymentID uint64 `json:"payment_id,omitempty"`
	Status    status `json:"status,omitempty"`
}

// entry is the concurrent map values structure.
// It contains information to track invoices and payments.
type entry struct {
	publicKey string
	id        uint64
	amount    uint64
	timestamp int64
}

func (e *entry) String() string {
	return fmt.Sprintf("public key: %s, amount: %d, timestamp: %d", e.publicKey, e.amount, e.timestamp)
}

// Server is in charge of managing the streams to which the events are sent to.
type Server interface {
	Close()
	CreateStream(id string) *sse.Stream
	HTTPHandler(w http.ResponseWriter, r *http.Request)
	Publish(id string, event *sse.Event)
}

// Streamer is in charge of tracking payments and executing actions based on them.
type Streamer interface {
	http.Handler
	io.Closer
	TrackPayment(rHash, publicKey string, amount uint64) uint64
}

type streamer struct {
	trackedPayments cmap.ConcurrentMap[string, entry]
	lnd             lightning.Client
	db              *db.DB
	server          Server
	logger          *logger.Logger
	winnersCh       <-chan []db.Winner
	blocksCh        chan<- *chainrpc.BlockEpoch
	config          config.SSE
}

// NewStreamer returns a new event streamer.
func NewStreamer(
	config config.SSE,
	db *db.DB,
	lnd lightning.Client,
	winnersCh <-chan []db.Winner,
	blocksCh chan<- *chainrpc.BlockEpoch,
) (Streamer, error) {
	logger, err := logger.New(config.Logger)
	if err != nil {
		return nil, err
	}

	server := sse.New()
	server.AutoReplay = false
	server.CreateStream(streamID)

	streamer := &streamer{
		config:          config,
		server:          server,
		lnd:             lnd,
		db:              db,
		trackedPayments: cmap.New[entry](),
		logger:          logger,
		winnersCh:       winnersCh,
		blocksCh:        blocksCh,
	}

	// Start listening for events to stream
	ctx := context.Background()
	go streamer.subscribeBlocks(ctx)
	go streamer.subscribeChannelEvents(ctx)
	go streamer.subscribeInvoices(ctx)
	go streamer.subscribePayments(ctx)
	go streamer.subscribeWinners(ctx)

	return streamer, nil
}

// Close closes all of the streams and connections.
func (s *streamer) Close() error {
	s.server.Close()
	return nil
}

func (s *streamer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Overwrite server.WriteTimeout to avoid a short timeout.
	// See https://github.com/golang/go/issues/16100
	rc := http.NewResponseController(w)
	// This will close the connection after the deadline
	// Use time.Time{} to leave the connection open indefinitely
	rc.SetWriteDeadline(time.Now().UTC().Add(s.config.Deadline))
	s.server.HTTPHandler(w, r)
}

// TrackPayment watches the rHash for updates to execute a certain action and returns the ID of the
// payment.
//
// Takes payment hashes from both invoices (in) and payments (out).
func (s *streamer) TrackPayment(rHash, publicKey string, amount uint64) uint64 {
	entry := entry{
		id:        rand.Uint64(),
		publicKey: publicKey,
		amount:    amount,
		timestamp: time.Now().Unix(),
	}
	s.trackedPayments.Set(rHash, entry)
	return entry.id
}

// removeExpiredPayments deletes all payments that were tracked more than one day ago.
func (s *streamer) removeExpiredPayments() {
	minTimestamp := time.Now().Add(-lightning.DefaultInvoiceExpiry).Unix()
	for item := range s.trackedPayments.IterBuffered() {
		if item.Val.timestamp < minTimestamp {
			s.trackedPayments.Remove(item.Key)
		}
	}
}

// subscribeBlocks listens to a stream of blocks and sends an event to all subscribers when a new
// block is found.
func (s *streamer) subscribeBlocks(ctx context.Context) {
	stream, err := s.lnd.SubscribeBlocks(ctx)
	if err != nil {
		s.logger.Error(errors.Wrap(err, "subscribing to blocks stream"))
		return
	}

	for {
		block, err := stream.Recv()
		if err != nil {
			s.logger.Error(errors.Wrap(err, "receiving events from blocks stream"))
			return
		}

		s.blocksCh <- block
	}
}

func (s *streamer) subscribeChannelEvents(ctx context.Context) {
	stream, err := s.lnd.SubscribeChannelEvents(ctx)
	if err != nil {
		s.logger.Fatal(errors.Wrap(err, "subscribing to channel events stream"))
		return
	}

	for {
		update, err := stream.Recv()
		if err != nil {
			s.logger.Error(errors.Wrap(err, "receiving events from channel events stream"))
			return
		}

		if ch, ok := update.Channel.(*lnrpc.ChannelEventUpdate_OpenChannel); ok {
			// Skip streaming events about private channels openings
			if ch.OpenChannel.Private || !ch.OpenChannel.Active {
				continue
			}
		}

		// When a new channel is opened or closed, get the updated balance and send it through
		// the stream
		if update.Type == lnrpc.ChannelEventUpdate_OPEN_CHANNEL ||
			update.Type == lnrpc.ChannelEventUpdate_CLOSED_CHANNEL {
			// Wait one second for the LND backend to update the channel list
			time.Sleep(time.Second)

			lotteryInfo, err := lottery.GetInfo(ctx, s.lnd, s.db)
			if err != nil {
				s.logger.Error(errors.Wrap(err, "getting lottery information"))
				return
			}

			payload := &infoPayload{
				PrizePool: &lotteryInfo.PrizePool,
				Capacity:  &lotteryInfo.Capacity,
			}
			s.publish(infoEvent, payload)
		}
	}
}

// subscribeInvoices listens to a stream of invoices and performs a specific action when finds one
// that succeeded and was being tracked by the API.
func (s *streamer) subscribeInvoices(ctx context.Context) {
	stream, err := s.lnd.SubscribeInvoices(ctx)
	if err != nil {
		s.logger.Fatal(errors.Wrap(err, "subscribing to invoices stream"))
		return
	}

	for {
		invoice, err := stream.Recv()
		if err != nil {
			s.logger.Error(errors.Wrap(err, "receiving events from invoices stream"))
			return
		}

		// Stream only settled invoices
		if invoice.State == lnrpc.Invoice_SETTLED {
			rHash := hex.EncodeToString(invoice.RHash)
			entry, ok := s.trackedPayments.Get(rHash)
			if !ok {
				continue
			}

			s.addBet(rHash, entry)
			payload := &invoicesPayload{
				PaymentID: entry.id,
				PublicKey: entry.publicKey,
				Amount:    entry.amount,
				Status:    success,
			}
			s.publish(invoicesEvent, payload)
		}
	}
}

// subscribePayments listens to a stream of payments and performs a specific action when finds one
// that succeeded and was being tracked by the API.
func (s *streamer) subscribePayments(ctx context.Context) {
	stream, err := s.lnd.SubscribePayments(ctx)
	if err != nil {
		s.logger.Fatal(errors.Wrap(err, "subscribing to payments stream"))
		return
	}

	for {
		payment, err := stream.Recv()
		if err != nil {
			s.logger.Error(errors.Wrap(err, "receiving events from payments stream"))
			return
		}

		entry, ok := s.trackedPayments.Get(payment.PaymentHash)
		if !ok {
			continue
		}

		switch payment.Status {
		case lnrpc.Payment_FAILED:
			// Give the funds back to the user
			s.restoreFunds(payment.PaymentHash, entry)
			s.logger.Errorf("Failed to pay invoice %s: %s",
				payment.PaymentHash, payment.FailureReason)

			payload := &paymentsPayload{
				PaymentID: entry.id,
				Status:    failed,
				Error:     payment.FailureReason.String(),
			}
			s.publish(paymentsEvent, payload)

		case lnrpc.Payment_SUCCEEDED:
			payload := &paymentsPayload{
				PaymentID: entry.id,
				Status:    success,
			}
			s.publish(paymentsEvent, payload)
		}
	}
}

// subscribeWinners streams winners when they are known and restarts the prize pool.
func (s *streamer) subscribeWinners(ctx context.Context) {
	for {
		select {
		case winners := <-s.winnersCh:
			lotteryInfo, err := lottery.GetInfo(ctx, s.lnd, s.db)
			if err != nil {
				s.logger.Error(errors.Wrap(err, "getting lottery information"))
				return
			}

			payload := &infoPayload{
				PrizePool:  &lotteryInfo.PrizePool,
				Capacity:   &lotteryInfo.Capacity,
				Winners:    &winners,
				NextHeight: &lotteryInfo.NextHeight,
			}
			s.publish(infoEvent, payload)

			// Remove invoices/payments created more than 1 day ago
			s.removeExpiredPayments()

		case <-ctx.Done():
			return
		}
	}
}

func (s *streamer) publish(event []byte, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error(errors.Wrap(err, "encoding stream payload"))
		return
	}
	s.server.Publish(streamID, &sse.Event{
		Event: event,
		Data:  data,
	})
}

func (s *streamer) addBet(rHash string, e entry) {
	// Stop tracking payment
	s.trackedPayments.Remove(rHash)

	bet := db.Bet{
		PublicKey: e.publicKey,
		Tickets:   e.amount,
	}
	if err := s.db.Bets.Add(bet); err != nil {
		s.logger.Error(errors.Wrapf(err, "adding bet: %s from %s", rHash, e.publicKey))
	}
}

// restoreFunds gives the user back the prizes that were discounted from him. It should be executed
// only after a payment has failed.
func (s *streamer) restoreFunds(rHash string, e entry) {
	// Stop tracking payment
	s.trackedPayments.Remove(rHash)

	// We should restore the prizes only if the public key is stored as a winner
	if prizes, err := s.db.Prizes.Get(e.publicKey); err != nil || prizes == 0 {
		s.logger.Error("tried restoring prizes to a user that is not a winner")
		return
	}

	height, err := s.db.Lotteries.GetNextHeight()
	if err != nil {
		s.logger.Error("getting next height")
		return
	}

	winner := db.Winner{
		PublicKey: e.publicKey,
		Prize:     e.amount,
	}
	if err := s.db.Winners.Add(height, []db.Winner{winner}); err != nil {
		s.logger.Error(
			errors.Wrapf(err, "restoring funds. Public key %s, payment %s", e.publicKey, rHash),
		)
	}
}
