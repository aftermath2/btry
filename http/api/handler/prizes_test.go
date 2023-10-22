package handler_test

import (
	"encoding/json"
	"net/http"

	"github.com/aftermath2/BTRY/http/api/handler"
	"github.com/pkg/errors"
)

func (h *HandlerSuite) TestGetPrizes() {
	publicKey := "e68b99fc5f60c971926fdc3a3af38ccf67e6f4306ab1c388735533e7c5dcc749"
	h.SetAuthorizationKey(publicKey)

	prizes := uint64(100)
	h.winnersMock.On("GetPrizes", publicKey).Return(prizes, nil)

	h.handler.GetPrizes(h.rec, h.req)

	var response handler.PrizesResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusOK, h.rec.Code)
	h.Equal(prizes, response.Prizes)
}

func (h *HandlerSuite) TestGetPrizesInvalidPublicKey() {
	h.SetAuthorizationKey("invalid")

	h.handler.GetPrizes(h.rec, h.req)

	h.Equal(http.StatusBadRequest, h.rec.Code)
}

func (h *HandlerSuite) TestGetPrizesInternalError() {
	publicKey := "e68b99fc5f60c971926fdc3a3af38ccf67e6f4306ab1c388735533e7c5dcc749"
	h.SetAuthorizationKey(publicKey)

	expectedErr := errors.New("test err")
	h.winnersMock.On("GetPrizes", publicKey).Return(uint64(0), expectedErr)

	h.handler.GetPrizes(h.rec, h.req)

	var response handler.ErrorResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusInternalServerError, h.rec.Code)
	h.Equal(expectedErr.Error(), response.Error)
}
