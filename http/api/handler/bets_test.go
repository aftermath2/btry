package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/http/api/handler"
	"github.com/aftermath2/BTRY/http/api/sse"
	"github.com/aftermath2/BTRY/lightning"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
)

type HandlerSuite struct {
	suite.Suite

	rec               *httptest.ResponseRecorder
	req               *http.Request
	betsMock          *db.BetsStoreMock
	lotteriesMock     *db.LotteriesStoreMock
	prizesMock        *db.PrizesStoreMock
	winnersMock       *db.WinnersStoreMock
	lndMock           *lightning.ClientMock
	handler           *handler.Handler
	eventStreamerMock *sse.StreamerMock
}

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, &HandlerSuite{})
}

func (h *HandlerSuite) SetupTest() {
	h.rec = httptest.NewRecorder()
	h.req = httptest.NewRequest(http.MethodGet, "/", nil)
	h.betsMock = db.NewBetsStoreMock()
	h.lotteriesMock = db.NewLotteriesStoreMock()
	h.prizesMock = db.NewPrizesStoreMock()
	h.winnersMock = db.NewWinnersStoreMock()
	h.lndMock = lightning.NewClientMock()
	h.eventStreamerMock = sse.NewStreamerMock()
	db := &db.DB{
		Bets:      h.betsMock,
		Lotteries: h.lotteriesMock,
		Prizes:    h.prizesMock,
		Winners:   h.winnersMock,
	}
	h.handler = handler.New(h.lndMock, db, h.eventStreamerMock)
}

func (h *HandlerSuite) SetAuthorizationKey(publicKey string) {
	h.req.Header.Set("Authorization", "Bearer "+publicKey)
}

func (h *HandlerSuite) SetDefaultAuthorizationKey() {
	h.req.Header.Set("Authorization", "Bearer e68b99fc5f60c971926fdc3a3af38ccf67e6f4306ab1c388735533e7c5dcc749")
}

func (h *HandlerSuite) TestGetBets() {
	bets := []db.Bet{
		{
			PublicKey: "pubkey",
			Index:     17,
			Tickets:   17,
		},
		{
			PublicKey: "pubkey2",
			Index:     25,
			Tickets:   8,
		},
	}
	h.betsMock.On("List", uint64(0), uint64(0), false).Return(bets, nil)

	h.handler.GetBets(h.rec, h.req)

	var response handler.BetsResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusOK, h.rec.Code)
	h.Equal(bets, response.Bets)
}

func (h *HandlerSuite) TestGetBetsParameters() {
	offset := uint64(1)
	limit := uint64(5)
	reverse := true

	url := url.Values{}
	url.Add("offset", strconv.FormatUint(offset, 10))
	url.Add("limit", strconv.FormatUint(limit, 10))
	url.Add("reverse", strconv.FormatBool(reverse))
	h.req = httptest.NewRequest(http.MethodPost, "/bets?"+url.Encode(), nil)

	h.betsMock.On("List", offset, limit, reverse).Return([]db.Bet{}, nil)

	h.handler.GetBets(h.rec, h.req)

	h.betsMock.AssertExpectations(h.T())
}

func (h *HandlerSuite) TestGetBetsInvalidParameters() {
	cases := []struct {
		desc  string
		key   string
		value string
	}{
		{
			desc:  "Invalid offset",
			key:   "offset",
			value: "false",
		},
		{
			desc:  "Invalid limit",
			key:   "limit",
			value: "false",
		},
		{
			desc:  "Invalid reverse",
			key:   "reverse",
			value: "five",
		},
	}

	for _, tc := range cases {
		h.Run(tc.desc, func() {
			url := url.Values{}
			url.Add(tc.key, tc.value)
			h.req = httptest.NewRequest(http.MethodPost, "/bets?"+url.Encode(), nil)

			h.handler.GetBets(h.rec, h.req)

			h.Equal(http.StatusBadRequest, h.rec.Code)
		})
	}
}

func (h *HandlerSuite) TestGetBetsInternalError() {
	expectedErr := errors.New("test error")
	h.betsMock.On("List", uint64(0), uint64(0), false).Return([]db.Bet{}, expectedErr)

	h.handler.GetBets(h.rec, h.req)

	var response handler.ErrorResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusInternalServerError, h.rec.Code)
	h.Equal(expectedErr.Error(), response.Error)
}
