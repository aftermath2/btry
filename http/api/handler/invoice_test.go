package handler_test

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/aftermath2/BTRY/http/api/handler"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/pkg/errors"
)

func (h *HandlerSuite) TestGetInvoice() {
	amount := uint64(2000)
	h.req = httptest.NewRequest(http.MethodGet, "/invoice?amount="+strconv.FormatUint(amount, 10), nil)
	publicKey := "e68b99fc5f60c971926fdc3a3af38ccf67e6f4306ab1c388735533e7c5dcc749"
	h.SetAuthorizationKey(publicKey)

	ctx := h.req.Context()
	h.betsMock.On("GetPrizePool").Return(uint64(0), nil)
	h.lndMock.On("RemoteBalance", ctx).Return(int64(1_000_000), nil)
	h.lotteriesMock.On("GetNextHeight").Return(uint32(1), nil)

	addInvoiceResp := &lnrpc.AddInvoiceResponse{
		RHash:          []byte("rhash"),
		PaymentRequest: "pr",
		AddIndex:       0,
		PaymentAddr:    []byte("addr"),
	}
	h.lndMock.On("AddInvoice", ctx, amount).Return(addInvoiceResp, nil)

	paymentID := uint64(123456)
	h.eventStreamerMock.On("TrackPayment", hex.EncodeToString(addInvoiceResp.RHash), publicKey, amount).Return(paymentID)

	h.handler.GetInvoice(h.rec, h.req)

	var response handler.InvoiceResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusOK, h.rec.Code)
	h.Equal(paymentID, response.PaymentID)
	h.Equal(addInvoiceResp.PaymentRequest, response.Invoice)
}

func (h *HandlerSuite) TestGetInvoiceInvalidPublicKey() {
	h.SetAuthorizationKey("invalid")

	h.handler.GetInvoice(h.rec, h.req)

	h.Equal(http.StatusBadRequest, h.rec.Code)
}

func (h *HandlerSuite) TestGetInvoiceInvalidAmount() {
	h.req = httptest.NewRequest(http.MethodGet, "/invoice?amount=five", nil)
	h.SetDefaultAuthorizationKey()

	h.handler.GetInvoice(h.rec, h.req)

	h.Equal(http.StatusBadRequest, h.rec.Code)
}

func (h *HandlerSuite) TestGetInvoiceExceedsCapacity() {
	h.req = httptest.NewRequest(http.MethodGet, "/invoice?amount=5000000", nil)
	h.SetDefaultAuthorizationKey()

	h.lndMock.On("RemoteBalance", h.req.Context()).Return(int64(5000000), nil)
	h.betsMock.On("GetPrizePool").Return(uint64(0), nil)
	h.lotteriesMock.On("GetNextHeight").Return(uint32(1), nil)

	h.handler.GetInvoice(h.rec, h.req)

	h.Equal(http.StatusBadRequest, h.rec.Code)
}

func (h *HandlerSuite) TestGetInvoiceGetInfoError() {
	amount := uint64(21000)
	h.req = httptest.NewRequest(http.MethodGet, "/invoice?amount="+strconv.FormatUint(amount, 10), nil)
	h.SetDefaultAuthorizationKey()

	ctx := h.req.Context()
	expectedErr := errors.New("test err")
	h.lndMock.On("RemoteBalance", ctx).Return(int64(0), expectedErr)

	h.handler.GetInvoice(h.rec, h.req)

	var response handler.ErrorResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusInternalServerError, h.rec.Code)
	h.Equal(expectedErr.Error(), response.Error)
}

func (h *HandlerSuite) TestGetInvoiceAddInvoiceError() {
	amount := uint64(21000)
	h.req = httptest.NewRequest(http.MethodGet, "/invoice?amount="+strconv.FormatUint(amount, 10), nil)
	h.SetDefaultAuthorizationKey()

	ctx := h.req.Context()
	h.betsMock.On("GetPrizePool").Return(uint64(0), nil)
	h.lndMock.On("RemoteBalance", ctx).Return(int64(1_000_000), nil)
	h.lotteriesMock.On("GetNextHeight").Return(uint32(1), nil)

	expectedErr := errors.New("test err")
	h.lndMock.On("AddInvoice", ctx, amount).Return(nil, expectedErr)

	h.handler.GetInvoice(h.rec, h.req)

	var response handler.ErrorResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusInternalServerError, h.rec.Code)
	h.Equal(expectedErr.Error(), response.Error)
}
