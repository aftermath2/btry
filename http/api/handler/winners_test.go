package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"

	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/http/api/handler"

	"github.com/pkg/errors"
)

func (h *HandlerSuite) TestGetWinners() {
	winners := []db.Winner{
		{
			PublicKey: "pubkey",
			Prizes:    1,
			Ticket:    1,
			CreatedAt: 0,
		},
		{
			PublicKey: "pubkey2",
			Prizes:    100,
			Ticket:    5,
			CreatedAt: 95,
		},
	}
	h.winnersMock.On("List").Return(winners, nil)

	h.handler.GetWinners(h.rec, h.req)

	var response handler.WinnersResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusOK, h.rec.Code)
	h.Equal(winners, response.Winners)
}

func (h *HandlerSuite) TestGetWinnersInternalError() {
	expectedErr := errors.New("test err")
	h.winnersMock.On("List").Return(nil, expectedErr)

	h.handler.GetWinners(h.rec, h.req)

	var response handler.ErrorResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusInternalServerError, h.rec.Code)
	h.Equal(expectedErr.Error(), response.Error)
}

func (h *HandlerSuite) TestGetWinnersHistory() {
	winners := []db.Winner{
		{
			PublicKey: "pubkey",
			Prizes:    1,
			Ticket:    1,
			CreatedAt: 0,
		},
		{
			PublicKey: "pubkey2",
			Prizes:    100,
			Ticket:    5,
			CreatedAt: 95,
		},
	}
	h.winnersMock.On("ListHistory", uint64(0), uint64(0)).Return(winners, nil)

	h.handler.GetWinnersHistory(h.rec, h.req)

	var response handler.WinnersResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusOK, h.rec.Code)
	h.Equal(winners, response.Winners)
}

func (h *HandlerSuite) TestGetWinnersHistoryParameters() {
	from := uint64(1)
	to := uint64(5)

	url := url.Values{}
	url.Add("from", strconv.FormatUint(from, 10))
	url.Add("to", strconv.FormatUint(to, 10))
	h.req = httptest.NewRequest(http.MethodPost, "/winners/history?"+url.Encode(), nil)

	h.winnersMock.On("ListHistory", from, to).Return([]db.Winner{}, nil)

	h.handler.GetWinnersHistory(h.rec, h.req)

	h.winnersMock.AssertExpectations(h.T())
}

func (h *HandlerSuite) TestGetWinnersHistoryInvalidParameter() {
	cases := []struct {
		desc  string
		key   string
		value string
	}{
		{
			desc:  "Invalid from",
			key:   "from",
			value: "false",
		},
		{
			desc:  "Invalid to",
			key:   "to",
			value: "sats",
		},
	}

	for _, tc := range cases {
		h.Run(tc.desc, func() {
			url := url.Values{}
			url.Add(tc.key, tc.value)
			h.req = httptest.NewRequest(http.MethodPost, "/winners/history?"+url.Encode(), nil)

			h.handler.GetWinnersHistory(h.rec, h.req)

			h.Equal(http.StatusBadRequest, h.rec.Code)
		})
	}
}

func (h *HandlerSuite) TestGetWinnersHistoryInternalError() {
	expectedErr := errors.New("test err")
	h.winnersMock.On("ListHistory", uint64(0), uint64(0)).Return(nil, expectedErr)

	h.handler.GetWinnersHistory(h.rec, h.req)

	var response handler.ErrorResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusInternalServerError, h.rec.Code)
	h.Equal(expectedErr.Error(), response.Error)
}
