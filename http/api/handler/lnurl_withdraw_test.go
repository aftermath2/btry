package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/aftermath2/BTRY/http/api/handler"
	"github.com/fiatjaf/go-lnurl"
	"github.com/pkg/errors"
)

func (h *HandlerSuite) TestLNURLWithdraw() {
	url := url.Values{}
	url.Add("signature", validSignature)
	h.req = httptest.NewRequest(http.MethodPost, "/lnurl/withdraw?"+url.Encode(), nil)
	h.SetAuthorizationKey(validPublicKey)

	prizes := uint64(2_000_000)
	h.winnersMock.On("GetPrizes", validPublicKey).Return(prizes, nil)

	h.handler.LNURLWithdraw(h.rec, h.req)

	fee := prizes * handler.LNURLWithdrawFeePPM / 1_000_000
	minWithdrawableMsat := int64(1000)
	maxWithdrawableMsat := int64(prizes-fee) * 1000
	callback := fmt.Sprintf("%s/withdraw?fee=%d&pubkey=%s",
		strings.TrimSuffix(h.req.RequestURI, h.req.URL.Path),
		fee,
		validPublicKey,
	)

	var response *lnurl.LNURLWithdrawResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusOK, h.rec.Code)
	h.Equal("withdrawalRequest", response.Tag)
	h.Equal("BTRY withdrawal", response.DefaultDescription)
	h.Equal(callback, response.Callback)
	h.Equal(validSignature, response.K1)
	h.Equal(minWithdrawableMsat, response.MinWithdrawable)
	h.Equal(maxWithdrawableMsat, response.MaxWithdrawable)
}

func (h *HandlerSuite) TestLNURLWithdrawNoPrizes() {
	url := url.Values{}
	url.Add("signature", validSignature)
	h.req = httptest.NewRequest(http.MethodPost, "/lnurl/withdraw?"+url.Encode(), nil)
	h.SetAuthorizationKey(validPublicKey)

	prizes := uint64(0)
	h.winnersMock.On("GetPrizes", validPublicKey).Return(prizes, nil)

	h.handler.LNURLWithdraw(h.rec, h.req)

	callback := fmt.Sprintf("%s/withdraw?fee=%d&pubkey=%s",
		strings.TrimSuffix(h.req.RequestURI, h.req.URL.Path),
		0,
		validPublicKey,
	)

	var response *lnurl.LNURLWithdrawResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusOK, h.rec.Code)
	h.Equal("withdrawalRequest", response.Tag)
	h.Equal("BTRY withdrawal", response.DefaultDescription)
	h.Equal(callback, response.Callback)
	h.Equal(validSignature, response.K1)
	h.Equal(int64(0), response.MinWithdrawable)
	h.Equal(int64(0), response.MaxWithdrawable)
}

func (h *HandlerSuite) TestLNURLWithdrawInvaliAuth() {
	testCases := []struct {
		desc      string
		signature string
		publicKey string
	}{
		{
			desc:      "Invalid public key",
			publicKey: "public key",
		},
		{
			desc:      "Invalid signature",
			publicKey: "470541ae525b58f98160e5d85a20697e16096020b833b84fb394e0099c874736",
			signature: "df26b9f6e916a57a2493dd18230a7d3071ade982eae2d40e41ae328a0ea9c1654395a2d8cceef9927c60a9bc8dfdf319e30f7945660fb84740c6155b8855f60c",
		},
		{
			desc:      "Empty signature",
			publicKey: validPublicKey,
			signature: "",
		},
	}

	for _, tc := range testCases {
		h.Run(tc.desc, func() {
			url := url.Values{}
			url.Add("signature", tc.signature)
			h.req = httptest.NewRequest(http.MethodPost, "/lnurl/withdraw?"+url.Encode(), nil)
			h.SetAuthorizationKey(tc.publicKey)

			h.handler.LNURLWithdraw(h.rec, h.req)

			h.Equal(http.StatusBadRequest, h.rec.Code)
		})
	}
}

func (h *HandlerSuite) TestLNURLWithdrawInternalError() {
	url := url.Values{}
	url.Add("signature", validSignature)
	h.req = httptest.NewRequest(http.MethodPost, "/lnurl/withdraw?"+url.Encode(), nil)
	h.SetAuthorizationKey(validPublicKey)

	expectedErr := errors.New("test err")
	h.winnersMock.On("GetPrizes", validPublicKey).Return(uint64(0), expectedErr)

	h.handler.LNURLWithdraw(h.rec, h.req)

	var response lnurl.LNURLErrorResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusInternalServerError, h.rec.Code)
	h.Equal(expectedErr.Error(), response.Reason)
}
