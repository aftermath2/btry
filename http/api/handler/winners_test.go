package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/http/api/handler"

	"github.com/pkg/errors"
)

func (h *HandlerSuite) TestGetWinners() {
	winners := []db.Winner{
		{
			PublicKey: "pubkey",
			Prize:     1,
			Ticket:    1,
		},
		{
			PublicKey: "pubkey2",
			Prize:     100,
			Ticket:    5,
		},
	}
	h.winnersMock.On("List", uint32(0)).Return(winners, nil)

	h.handler.GetWinners(h.rec, h.req)

	var response handler.WinnersResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusOK, h.rec.Code)
	h.Equal(winners, response.Winners)
}

func (h *HandlerSuite) TestGetWinnersInvalidHeight() {
	h.req = httptest.NewRequest(http.MethodGet, "/winners?height=one", nil)
	h.handler.GetWinners(h.rec, h.req)
	h.Equal(http.StatusBadRequest, h.rec.Code)
}

func (h *HandlerSuite) TestGetWinnersInternalError() {
	expectedErr := errors.New("test err")
	h.winnersMock.On("List", uint32(0)).Return(nil, expectedErr)

	h.handler.GetWinners(h.rec, h.req)

	var response handler.ErrorResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusInternalServerError, h.rec.Code)
	h.Equal(expectedErr.Error(), response.Error)
}
