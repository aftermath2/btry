package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/fiatjaf/go-lnurl"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestGetAuthPublicKey(t *testing.T) {
	publicKey := "e68b99fc5f60c971926fdc3a3af38ccf67e6f4306ab1c388735533e7c5dcc749"

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+publicKey)

	gotPublicKey, err := getAuthPublicKey(req)
	assert.NoError(t, err)

	assert.Equal(t, publicKey, gotPublicKey)
}

func TestGetAuthPublicKeyErrors(t *testing.T) {
	cases := []struct {
		desc      string
		publicKey string
	}{
		{
			desc:      "Invalid length",
			publicKey: "public_key",
		},
		{
			desc:      "Invalid encoding",
			publicKey: "168b99fc5f60c971926fdc3a3af38ccf67e6f4306ab1c388735533e7c5dcc74z",
		},
		{
			desc:      "Invalid header value",
			publicKey: "Bearer e68b99fc5f60c971926fdc3a3af38ccf67e6f4306ab1c388735533e7c5dcc749",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer "+tc.publicKey)

			_, err := getAuthPublicKey(req)
			assert.Error(t, err)
		})
	}
}

func TestParseIntParam(t *testing.T) {
	cases := []struct {
		desc          string
		value         string
		expectedValue uint64
		required      bool
		fail          bool
	}{
		{
			desc:          "Default value",
			expectedValue: 0,
			required:      false,
		},
		{
			desc:          "Specified value",
			value:         "10",
			expectedValue: 10,
		},
		{
			desc:          "Specified required value",
			value:         "1",
			expectedValue: 1,
			required:      true,
		},
		{
			desc:     "Required but empty value",
			required: true,
			fail:     true,
		},
		{
			desc:  "Invalid value (negative)",
			value: "-823",
			fail:  true,
		},
		{
			desc:  "Invalid value (characters)",
			value: "satoshi",
			fail:  true,
		},
		{
			desc:  "Invalid value (boolean)",
			value: "true",
			fail:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			key := "fee"
			values := url.Values{}
			values.Set(key, tc.value)

			req := httptest.NewRequest(http.MethodGet, "/?"+values.Encode(), nil)

			actualValue, err := parseIntParam(req.URL.Query(), key, tc.required)
			if tc.fail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedValue, actualValue)
		})
	}
}

func TestSendResponse(t *testing.T) {
	rec := httptest.NewRecorder()

	code := http.StatusOK
	value := true
	sendResponse(rec, code, value)

	var response bool
	err := json.NewDecoder(rec.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, rec.Header().Get("Content-Type"), "application/json; charset=UTF-8")
	assert.Equal(t, code, rec.Code)
	assert.Equal(t, value, response)
}

func TestSendLNURLError(t *testing.T) {
	rec := httptest.NewRecorder()

	code := http.StatusInternalServerError
	testErr := errors.New("fiat")
	sendLNURLError(rec, code, testErr)

	var response lnurl.LNURLErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, code, rec.Code)
	var expectedURL *url.URL
	assert.Equal(t, expectedURL, response.URL)
	assert.Equal(t, testErr.Error(), response.Reason)
	assert.Equal(t, "ERROR", response.Status)
}

func TestSendError(t *testing.T) {
	rec := httptest.NewRecorder()

	code := http.StatusInternalServerError
	testErr := errors.New("fiat")
	sendError(rec, code, testErr)

	var response ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, code, rec.Code)
	assert.Equal(t, testErr.Error(), response.Error)
}
