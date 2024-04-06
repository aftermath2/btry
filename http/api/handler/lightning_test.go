package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/http/api/handler"

	"github.com/pkg/errors"
)

const lightningAddress = "satoshi@bitcoin.org"

func (h *HandlerSuite) TestGetLightningAddress() {
	address := "satoshi@test.xyz"
	h.lightningMock.On("GetAddress", validPublicKey).Return(address, nil)

	h.SetAuthorizationKey(validPublicKey)
	h.handler.GetLightningAddress(h.rec, h.req)

	var response handler.GetLightningAddressResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusOK, h.rec.Code)
	h.Equal(address, response.Address)
}

func (h *HandlerSuite) TestGetLightningAddressNoAddress() {
	h.lightningMock.On("GetAddress", validPublicKey).Return("", db.ErrNoAddress)

	h.SetAuthorizationKey(validPublicKey)
	h.handler.GetLightningAddress(h.rec, h.req)

	var response handler.GetLightningAddressResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusOK, h.rec.Code)
	h.False(response.HasAddress)
	h.Empty(response.Address)
}

func (h *HandlerSuite) TestGetLightningAddressError() {
	expectedErr := errors.New("test error")
	h.lightningMock.On("GetAddress", validPublicKey).Return("", expectedErr)

	h.SetAuthorizationKey(validPublicKey)
	h.handler.GetLightningAddress(h.rec, h.req)

	var response handler.ErrorResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusInternalServerError, h.rec.Code)
	h.Equal(expectedErr.Error(), response.Error)
}

func (h *HandlerSuite) TestGetLightningNoAuthError() {
	h.handler.GetLightningAddress(h.rec, h.req)

	h.Equal(http.StatusBadRequest, h.rec.Code)
}

func (h *HandlerSuite) TestSetLightningAddress() {
	url := url.Values{}
	url.Add("address", lightningAddress)
	h.req = httptest.NewRequest(http.MethodPost, "/lightning/address?"+url.Encode(), nil)
	h.SetAuthorizationKey(validPublicKey)

	h.lightningMock.On("SetAddress", validPublicKey, lightningAddress).Return(nil)

	h.handler.SetLightningAddress(h.rec, h.req)

	var response handler.SetLightningAddressResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusOK, h.rec.Code)
	h.True(response.Success)
}

func (h *HandlerSuite) TestSetLightningAddressNoAuth() {
	h.req = httptest.NewRequest(http.MethodPost, "/lightning/address", nil)

	h.handler.SetLightningAddress(h.rec, h.req)

	h.Equal(http.StatusBadRequest, h.rec.Code)
}

func (h *HandlerSuite) TestSetLightningAddressInvalidAddress() {
	url := url.Values{}
	url.Add("address", "invalid_address")
	h.req = httptest.NewRequest(http.MethodPost, "/lightning/address?"+url.Encode(), nil)
	h.SetAuthorizationKey(validPublicKey)

	h.handler.SetLightningAddress(h.rec, h.req)

	h.Equal(http.StatusBadRequest, h.rec.Code)
}

func (h *HandlerSuite) TestSetLightningAddressInternalError() {
	url := url.Values{}
	url.Add("address", lightningAddress)
	h.req = httptest.NewRequest(http.MethodPost, "/lightning/address?"+url.Encode(), nil)
	h.SetAuthorizationKey(validPublicKey)

	expectedErr := errors.New("test err")
	h.lightningMock.On("SetAddress", validPublicKey, lightningAddress).Return(expectedErr)

	h.handler.SetLightningAddress(h.rec, h.req)

	var response handler.ErrorResponse
	err := json.NewDecoder(h.rec.Body).Decode(&response)
	h.NoError(err)

	h.Equal(http.StatusInternalServerError, h.rec.Code)
	h.Equal(expectedErr.Error(), response.Error)
}
