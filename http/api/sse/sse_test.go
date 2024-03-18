package sse

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/lightning"
	"github.com/aftermath2/BTRY/logger"
	"github.com/aftermath2/BTRY/lottery"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/r3labs/sse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestNewStreamer(t *testing.T) {
	lndMock := lightning.NewClientMock()

	lndMock.On("SubscribeBlocks", context.Background()).
		Return(lightning.BlockedStreamMock[*chainrpc.BlockEpoch]{}, nil)
	lndMock.On("SubscribeChannelEvents", context.Background()).
		Return(lightning.BlockedStreamMock[*lnrpc.ChannelEventUpdate]{}, nil)
	lndMock.On("SubscribeInvoices", context.Background()).
		Return(lightning.BlockedStreamMock[*lnrpc.Invoice]{}, nil)
	lndMock.On("SubscribePayments", context.Background()).
		Return(lightning.BlockedStreamMock[*lnrpc.Payment]{}, nil)

	streamer, err := NewStreamer(
		config.SSE{Logger: config.Logger{Level: uint8(logger.DISABLED)}},
		&db.DB{},
		lndMock,
		make(<-chan []db.Winner),
		make(chan<- *chainrpc.BlockEpoch),
	)
	assert.NoError(t, err)

	err = streamer.Close()
	assert.NoError(t, err)
}

type SSESuite struct {
	suite.Suite

	betsMock      *db.BetsStoreMock
	lotteriesMock *db.LotteriesStoreMock
	prizesMock    *db.PrizesStoreMock
	winnersMock   *db.WinnersStoreMock
	lndMock       *lightning.ClientMock
	server        *ServerMock
	winnersCh     chan []db.Winner
	sse           streamer
}

func TestSSESuite(t *testing.T) {
	suite.Run(t, &SSESuite{})
}

func (s *SSESuite) SetupTest() {
	logger, err := logger.New(config.Logger{Level: uint8(logger.DISABLED)})
	s.NoError(err)

	s.betsMock = db.NewBetsStoreMock()
	s.lotteriesMock = db.NewLotteriesStoreMock()
	s.prizesMock = db.NewPrizesStoreMock()
	s.winnersMock = db.NewWinnersStoreMock()
	s.lndMock = lightning.NewClientMock()
	s.server = NewServerMock()
	s.winnersCh = make(chan []db.Winner)
	s.sse = streamer{
		server:          s.server,
		winnersCh:       s.winnersCh,
		logger:          logger,
		lnd:             s.lndMock,
		trackedPayments: cmap.New[entry](),
		db: &db.DB{
			Bets:      s.betsMock,
			Lotteries: s.lotteriesMock,
			Prizes:    s.prizesMock,
			Winners:   s.winnersMock,
		},
	}
}

func (s *SSESuite) TestClose() {
	s.server.On("Close").Return(nil)

	err := s.sse.Close()
	s.NoError(err)
}

func (s *SSESuite) TestTrackPayment() {
	rHash := "f0c92fd3aaf3"
	publicKey := "e68b99fc5f60c971926fdc3a3af38ccf67e6f4306ab1c388735533e7c5dcc749"
	amount := uint64(21000000)
	timestamp := time.Now().Unix()
	id := s.sse.TrackPayment(rHash, publicKey, amount)

	count := s.sse.trackedPayments.Count()
	s.Equal(1, count)

	payment, ok := s.sse.trackedPayments.Get(rHash)
	s.True(ok)

	s.Equal(id, payment.id)
	s.Equal(publicKey, payment.publicKey)
	s.Equal(amount, payment.amount)
	s.LessOrEqual(timestamp, payment.timestamp)
}

func (s *SSESuite) TestTrackConcurrentPayments() {
	count := 5
	var wg sync.WaitGroup
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func(i int) {
			str := strconv.Itoa(i)
			s.sse.TrackPayment(str, str, uint64(i))
			wg.Done()
		}(i)
	}

	wg.Wait()

	gotCount := s.sse.trackedPayments.Count()
	s.Equal(count, gotCount)
}

func (s *SSESuite) TestRemoveExpiredPayments() {
	s.sse.trackedPayments.Set("expired", entry{timestamp: 12763721})
	key := "not_expired"
	s.sse.trackedPayments.Set(key, entry{timestamp: time.Now().Unix()})

	s.sse.removeExpiredPayments()

	count := s.sse.trackedPayments.Count()
	s.Equal(1, count)

	_, ok := s.sse.trackedPayments.Get(key)
	s.True(ok)
}

func (s *SSESuite) TestServeHTTP() {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	s.server.On("HTTPHandler", rec, req)

	s.sse.ServeHTTP(rec, req)
}

type customEventsStreamMock[T any] struct {
	events  []T
	counter int
}

// Recv returns an event on the first n calls and then fails to exit inifite loops.
func (s *customEventsStreamMock[T]) Recv() (T, error) {
	if s.counter == len(s.events) {
		var v T
		return v, errors.New("exit")
	}

	event := s.events[s.counter]
	s.counter++

	return event, nil
}

func (s *SSESuite) TestSubscribeChannelEvents() {
	ctx := context.Background()
	stream := &customEventsStreamMock[*lnrpc.ChannelEventUpdate]{
		events: []*lnrpc.ChannelEventUpdate{{
			Channel: &lnrpc.ChannelEventUpdate_OpenChannel{
				OpenChannel: &lnrpc.Channel{Active: true, Private: false},
			},
		}},
	}

	s.lndMock.On("SubscribeChannelEvents", ctx).Return(stream, nil)

	remoteBalance := int64(2_500_000)
	prizePool := uint64(1_000_000)
	blockHeight := uint32(1)

	s.lndMock.On("RemoteBalance", ctx).Return(remoteBalance, nil)
	s.lotteriesMock.On("GetNextHeight").Return(blockHeight, nil)
	s.betsMock.On("GetPrizePool", blockHeight).Return(prizePool, nil)

	pp := int64(prizePool)
	capacity := remoteBalance / lottery.CapacityDivisor
	payload := &infoPayload{
		PrizePool: &pp,
		Capacity:  &capacity,
	}

	data, err := json.Marshal(payload)
	s.NoError(err)

	event := &sse.Event{Event: infoEvent, Data: data}
	s.server.On("Publish", streamID, event)

	s.sse.subscribeChannelEvents(ctx)
}

func (s *SSESuite) TestSubscribeChannelEventsPrivateChannel() {
	ctx := context.Background()
	stream := &customEventsStreamMock[*lnrpc.ChannelEventUpdate]{
		events: []*lnrpc.ChannelEventUpdate{{
			Channel: &lnrpc.ChannelEventUpdate_OpenChannel{
				OpenChannel: &lnrpc.Channel{Active: true, Private: true},
			},
		}},
	}

	s.lndMock.On("SubscribeChannelEvents", ctx).Return(stream, nil)

	s.sse.subscribeChannelEvents(ctx)
}

func (s *SSESuite) TestSubscribeChannelEventsInactiveChannel() {
	ctx := context.Background()
	stream := &customEventsStreamMock[*lnrpc.ChannelEventUpdate]{
		events: []*lnrpc.ChannelEventUpdate{{
			Channel: &lnrpc.ChannelEventUpdate_OpenChannel{
				OpenChannel: &lnrpc.Channel{Active: false, Private: true},
			},
		}},
	}

	s.lndMock.On("SubscribeChannelEvents", ctx).Return(stream, nil)

	s.sse.subscribeChannelEvents(ctx)
}

func (s *SSESuite) TestSubscribeInvoices() {
	ctx := context.Background()
	rHash := []byte("rHash")
	publicKey := "publicKey"
	amount := uint64(200)

	stream := &customEventsStreamMock[*lnrpc.Invoice]{
		events: []*lnrpc.Invoice{{
			RHash: rHash, State: lnrpc.Invoice_SETTLED,
		}},
	}
	s.lndMock.On("SubscribeInvoices", ctx).Return(stream, nil)

	bet := db.Bet{
		PublicKey: publicKey,
		Tickets:   amount,
	}
	s.betsMock.On("Add", bet).Return(nil)

	id := s.sse.TrackPayment(hex.EncodeToString(rHash), publicKey, amount)
	payload := &invoicesPayload{
		PaymentID: id,
		PublicKey: publicKey,
		Amount:    amount,
		Status:    success,
	}

	data, err := json.Marshal(payload)
	s.NoError(err)

	event := &sse.Event{Event: invoicesEvent, Data: data}
	s.server.On("Publish", streamID, event)

	s.sse.subscribeInvoices(ctx)
}

func (s *SSESuite) TestSubscribeInvoicesUntracked() {
	ctx := context.Background()
	rHash := []byte("rHash")

	stream := &customEventsStreamMock[*lnrpc.Invoice]{
		events: []*lnrpc.Invoice{{
			RHash: rHash, State: lnrpc.Invoice_SETTLED,
		}},
	}
	s.lndMock.On("SubscribeInvoices", ctx).Return(stream, nil)

	_, ok := s.sse.trackedPayments.Get(hex.EncodeToString(rHash))
	s.False(ok)

	s.sse.subscribeInvoices(ctx)
}

func (s *SSESuite) TestSubscribePayments() {
	ctx := context.Background()
	rHash := "rHash"
	publicKey := "publicKey"
	amount := uint64(2016)

	stream := &customEventsStreamMock[*lnrpc.Payment]{
		events: []*lnrpc.Payment{{
			PaymentHash: rHash, Status: lnrpc.Payment_SUCCEEDED,
		}},
	}

	id := s.sse.TrackPayment(rHash, publicKey, amount)
	payload := &paymentsPayload{PaymentID: id, Status: success}
	data, err := json.Marshal(payload)
	s.NoError(err)

	event := &sse.Event{Event: paymentsEvent, Data: data}
	s.server.On("Publish", streamID, event)
	s.lndMock.On("SubscribePayments", ctx).Return(stream, nil)

	s.sse.subscribePayments(ctx)
}

func (s *SSESuite) TestSubscribePaymentsFailed() {
	ctx := context.Background()
	rHash := "rHash"
	publicKey := "publicKey"
	amount := uint64(2016)
	payment := &lnrpc.Payment{
		PaymentHash:   rHash,
		Status:        lnrpc.Payment_FAILED,
		FailureReason: lnrpc.PaymentFailureReason_FAILURE_REASON_TIMEOUT,
	}

	stream := &customEventsStreamMock[*lnrpc.Payment]{
		events: []*lnrpc.Payment{payment},
	}
	id := s.sse.TrackPayment(rHash, publicKey, amount)

	s.prizesMock.On("Get", publicKey).Return(uint64(0), nil)
	winners := []db.Winner{{PublicKey: publicKey, Prize: amount}}
	s.winnersMock.On("Add", winners).Return(nil)

	payload := &paymentsPayload{
		PaymentID: id,
		Status:    failed,
		Error:     payment.FailureReason.String(),
	}
	data, err := json.Marshal(payload)
	s.NoError(err)

	event := &sse.Event{Event: paymentsEvent, Data: data}
	s.server.On("Publish", streamID, event)
	s.lndMock.On("SubscribePayments", ctx).Return(stream, nil)

	s.sse.subscribePayments(ctx)
}

func (s *SSESuite) TestSubscribePaymentsUntracked() {
	ctx := context.Background()

	stream := &customEventsStreamMock[*lnrpc.Payment]{
		events: []*lnrpc.Payment{{
			PaymentHash: "rHash", Status: lnrpc.Payment_SUCCEEDED,
		}},
	}

	s.lndMock.On("SubscribePayments", ctx).Return(stream, nil)

	s.sse.subscribePayments(ctx)
}

func (s *SSESuite) TestSubscribeWinners() {
	ctx, cancel := context.WithCancel(context.Background())
	remoteBalance := int64(10)
	prizePool := uint64(25000)
	nextHeight := uint32(1)
	winners := []db.Winner{{PublicKey: "winner", Prize: 10, Ticket: 1}}

	s.lndMock.On("RemoteBalance", ctx).Return(remoteBalance, nil)
	s.lotteriesMock.On("GetNextHeight").Return(nextHeight, nil)
	s.betsMock.On("GetPrizePool", nextHeight).Return(prizePool, nil)

	pp := int64(prizePool)
	capacity := remoteBalance / lottery.CapacityDivisor
	payload := &infoPayload{
		PrizePool:  &pp,
		Capacity:   &capacity,
		Winners:    &winners,
		NextHeight: &nextHeight,
	}

	data, err := json.Marshal(payload)
	s.NoError(err)

	event := &sse.Event{Event: infoEvent, Data: data}
	s.server.On("Publish", streamID, event)

	go func() {
		s.winnersCh <- winners

		// Force subscribeWinners infinite loop to exit
		cancel()
	}()

	s.sse.subscribeWinners(ctx)
}

func (s *SSESuite) TestPublish() {
	event := []byte("event")
	payload := 1
	sseEvent := &sse.Event{
		Event: event,
		// 1
		Data: []byte{0x31},
	}

	s.server.On("Publish", streamID, sseEvent)

	s.sse.publish(event, payload)
}

func (s *SSESuite) TestAddBet() {
	rHash := "hj432kl2ñ"
	entry := entry{
		publicKey: "publicKey",
		amount:    100,
	}
	s.sse.trackedPayments.Set(rHash, entry)

	bet := db.Bet{
		PublicKey: entry.publicKey,
		Tickets:   entry.amount,
	}
	s.betsMock.On("Add", bet).Return(nil)

	s.sse.addBet(rHash, entry)

	count := s.sse.trackedPayments.Count()
	s.Zero(count)
}

func (s *SSESuite) TestRestoreFunds() {
	rHash := "hj432kl2ñ"
	entry := entry{
		publicKey: "publicKey",
		amount:    100,
	}
	s.sse.trackedPayments.Set(rHash, entry)

	s.prizesMock.On("Get", entry.publicKey).Return(uint64(0), nil)
	nextHeight := uint32(1)
	s.lotteriesMock.On("GetNextHeight").Return(nextHeight, nil)

	winner := db.Winner{PublicKey: entry.publicKey, Prize: entry.amount}
	s.winnersMock.On("Add", nextHeight, []db.Winner{winner}).Return(nil)

	s.sse.restoreFunds(rHash, entry)

	count := s.sse.trackedPayments.Count()
	s.Zero(count)
}
