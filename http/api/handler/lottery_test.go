package handler_test

import (
	"encoding/json"
	"net/http"

	"github.com/aftermath2/BTRY/http/api/handler"
	"github.com/aftermath2/BTRY/lottery"
	"github.com/pkg/errors"
)

func (h *HandlerSuite) TestGetLottery() {
	remoteBalance := int64(500000)
	h.lndMock.On("RemoteBalance", h.req.Context()).Return(remoteBalance, nil)
	prizePool := uint64(50000)
	h.betsMock.On("GetPrizePool").Return(prizePool, nil)

	h.handler.GetLottery(h.rec, h.req)

	var response lottery.Info
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusOK, h.rec.Code)

	expectedCapacity := remoteBalance / lottery.CapacityDivisor
	h.Equal(expectedCapacity, response.Capacity)

	h.Equal(prizePool, uint64(response.PrizePool))
}

func (h *HandlerSuite) TestGetLotteryError() {
	expectedErr := errors.New("test error")
	h.lndMock.On("RemoteBalance", h.req.Context()).Return(int64(0), expectedErr)

	h.handler.GetLottery(h.rec, h.req)

	var response handler.ErrorResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusInternalServerError, h.rec.Code)
	h.Equal(expectedErr.Error(), response.Error)
}
