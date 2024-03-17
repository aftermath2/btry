package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"time"

	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/http/api/handler"

	"github.com/fiatjaf/go-lnurl"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/pkg/errors"
)

const (
	validPublicKey = "e68b99fc5f60c971926fdc3a3af38ccf67e6f4306ab1c388735533e7c5dcc749"
	validSignature = "b347d0accfdc0e9446921c9d61699432826c4d552201d35f5e81482ea0d7d8ecbc99e10c0c5" +
		"90858d44389a49bccd20c7c2f414f88230d418a2be43157336005"
)

func (h *HandlerSuite) TestWithdraw() {
	paymentRequest := "lnbcrt"
	fee := int64(10)

	url := url.Values{}
	url.Add("k1", validSignature)
	url.Add("pubkey", validPublicKey)
	url.Add("pr", paymentRequest)
	url.Add("fee", strconv.FormatInt(fee, 10))

	h.req = httptest.NewRequest(http.MethodPost, "/withdraw?"+url.Encode(), nil)
	ctx := h.req.Context()

	invoice := &lnrpc.PayReq{
		PaymentHash: "hash",
		NumSatoshis: 1000,
		Timestamp:   time.Now().Unix(),
		Expiry:      150000,
	}
	h.lndMock.On("DecodeInvoice", ctx, paymentRequest).Return(invoice, nil)

	withdrawAmount := uint64(invoice.NumSatoshis + fee)
	h.prizesMock.On("Withdraw", validPublicKey, withdrawAmount).Return(nil)

	h.lndMock.On("PayInvoice", ctx, invoice, fee, false).Return(nil, nil)

	paymentID := uint64(789)
	h.eventStreamerMock.On("TrackPayment", invoice.PaymentHash, validPublicKey, withdrawAmount).
		Return(paymentID)

	h.handler.Withdraw(h.rec, h.req)

	var response handler.WithdrawResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusOK, h.rec.Code)
	h.Equal("OK", response.Status)
	h.Equal(paymentID, response.PaymentID)
}

func (h *HandlerSuite) TestWithdrawInvalidParameters() {
	cases := []struct {
		desc      string
		publicKey string
		signature string
		invoice   string
		fee       string
	}{
		{
			desc:      "Invalid public key/signature",
			publicKey: "470541ae525b58f98160e5d85a20697e16096020b833b84fb394e0099c874736",
			signature: "df26b9f6e916a57a2493dd18230a7d3071ade982eae2d40e41ae328a0ea9c1654395a2d8c" +
				"ceef9927c60a9bc8dfdf319e30f7945660fb84740c6155b8855f60c",
		},
		{
			desc:      "Empty public key",
			signature: validSignature,
			publicKey: "",
		},
		{
			desc:      "Empty signature",
			signature: "",
		},
		{
			desc:      "Empty invoice",
			publicKey: validPublicKey,
			signature: validSignature,
			invoice:   "",
		},
		{
			desc:      "Invalid fee",
			publicKey: validPublicKey,
			signature: validSignature,
			invoice:   "pr",
			fee:       "a",
		},
	}

	for _, tc := range cases {
		h.Run(tc.desc, func() {
			url := url.Values{}
			url.Add("k1", tc.signature)
			url.Add("pubkey", tc.publicKey)
			url.Add("pr", tc.invoice)
			url.Add("fee", tc.fee)

			h.req = httptest.NewRequest(http.MethodPost, "/withdraw?"+url.Encode(), nil)

			h.handler.Withdraw(h.rec, h.req)

			h.Equal(http.StatusBadRequest, h.rec.Code)
		})
	}
}

func (h *HandlerSuite) TestWithdrawInsufficientPrizes() {
	paymentRequest := "lnbcrt"
	fee := int64(10)

	url := url.Values{}
	url.Add("k1", validSignature)
	url.Add("pubkey", validPublicKey)
	url.Add("pr", paymentRequest)
	url.Add("fee", strconv.FormatInt(fee, 10))

	h.req = httptest.NewRequest(http.MethodPost, "/withdraw?"+url.Encode(), nil)
	ctx := h.req.Context()

	invoice := &lnrpc.PayReq{
		PaymentHash: "hash",
		NumSatoshis: 1000,
		Timestamp:   time.Now().Unix(),
		Expiry:      150000,
	}
	h.lndMock.On("DecodeInvoice", ctx, paymentRequest).Return(invoice, nil)

	withdrawAmount := uint64(invoice.NumSatoshis + fee)
	h.prizesMock.On("Withdraw", validPublicKey, withdrawAmount).Return(db.ErrInsufficientPrizes)

	h.handler.Withdraw(h.rec, h.req)

	var response lnurl.LNURLErrorResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusBadRequest, h.rec.Code)
	h.Equal(db.ErrInsufficientPrizes.Error(), response.Reason)
}

func (h *HandlerSuite) TestWithdrawInvoiceDecodeError() {
	url := url.Values{}
	url.Add("k1", validSignature)
	url.Add("pubkey", validPublicKey)
	url.Add("fee", "1")
	paymentRequest := "lnbc"
	url.Set("pr", paymentRequest)

	h.req = httptest.NewRequest(http.MethodPost, "/withdraw?"+url.Encode(), nil)
	ctx := h.req.Context()

	expectedErr := errors.New("invalid payment request")
	h.lndMock.On("DecodeInvoice", ctx, paymentRequest).Return(nil, expectedErr)

	h.handler.Withdraw(h.rec, h.req)

	var response lnurl.LNURLErrorResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusBadRequest, h.rec.Code)
	h.Equal(expectedErr.Error(), response.Reason)
}

func (h *HandlerSuite) TestWithdrawInvoiceInvalidAmount() {
	url := url.Values{}
	url.Add("k1", validSignature)
	url.Add("pubkey", validPublicKey)
	url.Add("fee", "1")
	paymentRequest := "lntc"
	url.Set("pr", paymentRequest)

	h.req = httptest.NewRequest(http.MethodPost, "/withdraw?"+url.Encode(), nil)
	ctx := h.req.Context()

	invoice := &lnrpc.PayReq{
		NumSatoshis: 0,
	}
	h.lndMock.On("DecodeInvoice", ctx, paymentRequest).Return(invoice, nil)

	h.handler.Withdraw(h.rec, h.req)

	h.Equal(http.StatusBadRequest, h.rec.Code)
}

func (h *HandlerSuite) TestWithdrawExpiredInvoice() {
	url := url.Values{}
	url.Add("k1", validSignature)
	url.Add("pubkey", validPublicKey)
	url.Add("fee", "1")
	paymentRequest := "lntc"
	url.Set("pr", paymentRequest)

	h.req = httptest.NewRequest(http.MethodPost, "/withdraw?"+url.Encode(), nil)
	ctx := h.req.Context()

	invoice := &lnrpc.PayReq{
		NumSatoshis: 9899,
		Timestamp:   102893,
		Expiry:      10,
	}
	h.lndMock.On("DecodeInvoice", ctx, paymentRequest).Return(invoice, nil)

	h.handler.Withdraw(h.rec, h.req)

	h.Equal(http.StatusBadRequest, h.rec.Code)
}

func (h *HandlerSuite) TestWithdrawPayInvoiceError() {
	paymentRequest := "lnbcrt"
	fee := int64(10)

	url := url.Values{}
	url.Add("k1", validSignature)
	url.Add("pubkey", validPublicKey)
	url.Add("pr", paymentRequest)
	url.Add("fee", strconv.FormatInt(fee, 10))

	h.req = httptest.NewRequest(http.MethodPost, "/withdraw?"+url.Encode(), nil)
	ctx := h.req.Context()

	invoice := &lnrpc.PayReq{
		PaymentHash: "hash",
		NumSatoshis: 1000,
		Timestamp:   time.Now().Unix(),
		Expiry:      150000,
	}
	h.lndMock.On("DecodeInvoice", ctx, paymentRequest).Return(invoice, nil)

	withdrawAmount := uint64(invoice.NumSatoshis + fee)
	h.prizesMock.On("Withdraw", validPublicKey, withdrawAmount).Return(nil)

	paymentID := uint64(654)
	h.eventStreamerMock.On("TrackPayment", invoice.PaymentHash, validPublicKey, withdrawAmount).
		Return(paymentID)

	expectedErr := errors.New("test err")
	h.lndMock.On("PayInvoice", ctx, invoice, fee, false).Return(nil, expectedErr)

	h.handler.Withdraw(h.rec, h.req)

	var response lnurl.LNURLErrorResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusInternalServerError, h.rec.Code)
	h.Equal(expectedErr.Error(), response.Reason)
}
