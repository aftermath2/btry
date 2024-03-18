package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"

	"github.com/aftermath2/BTRY/http/api/handler"
	"github.com/aftermath2/BTRY/lottery"

	"github.com/pkg/errors"
)

func (h *HandlerSuite) TestGetLottery() {
	remoteBalance := int64(500000)
	h.lndMock.On("RemoteBalance", h.req.Context()).Return(remoteBalance, nil)
	prizePool := uint64(50000)
	nextHeight := uint32(145)
	h.lotteriesMock.On("GetNextHeight").Return(nextHeight, nil)
	h.betsMock.On("GetPrizePool", nextHeight).Return(prizePool, nil)

	h.handler.GetLottery(h.rec, h.req)

	var response lottery.Info
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusOK, h.rec.Code)

	expectedCapacity := remoteBalance / lottery.CapacityDivisor
	h.Equal(expectedCapacity, response.Capacity)

	h.Equal(prizePool, uint64(response.PrizePool))
	h.Equal(nextHeight, response.NextHeight)
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

func (h *HandlerSuite) TestListHeights() {
	heights := []uint32{
		2,
		3,
	}
	h.lotteriesMock.On("ListHeights", uint64(0), uint64(0), false).Return(heights, nil)

	h.handler.GetHeights(h.rec, h.req)

	var response handler.HeightsResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusOK, h.rec.Code)
	h.Equal(heights, response.Heights)
}

func (h *HandlerSuite) TestGetHeightsParameters() {
	offset := uint64(1)
	limit := uint64(5)
	reverse := true

	url := url.Values{}
	url.Add("offset", strconv.FormatUint(offset, 10))
	url.Add("limit", strconv.FormatUint(limit, 10))
	url.Add("reverse", strconv.FormatBool(reverse))
	h.req = httptest.NewRequest(http.MethodPost, "/heights?"+url.Encode(), nil)

	h.lotteriesMock.On("ListHeights", offset, limit, reverse).Return([]uint32{}, nil)

	h.handler.GetHeights(h.rec, h.req)

	h.lotteriesMock.AssertExpectations(h.T())
}

func (h *HandlerSuite) TestGetHeightsInvalidParameters() {
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
			h.req = httptest.NewRequest(http.MethodPost, "/heights?"+url.Encode(), nil)

			h.handler.GetHeights(h.rec, h.req)

			h.Equal(http.StatusBadRequest, h.rec.Code)
		})
	}
}

func (h *HandlerSuite) TestGetHeightsInternalError() {
	expectedErr := errors.New("test error")
	h.lotteriesMock.On("ListHeights", uint64(0), uint64(0), false).Return([]uint32{}, expectedErr)

	h.handler.GetHeights(h.rec, h.req)

	var response handler.ErrorResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusInternalServerError, h.rec.Code)
	h.Equal(expectedErr.Error(), response.Error)
}
