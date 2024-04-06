package lightning

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/fiatjaf/go-lnurl"
	"github.com/pkg/errors"
)

func getPayCallback(client *http.Client, address string, amountSat int64) (string, error) {
	name, host, ok := lnurl.ParseInternetIdentifier(address)
	if !ok {
		return "", errors.New("invalid lightning address")
	}

	uri := fmt.Sprintf("https://%s/.well-known/lnurlp/%s", host, name)
	callback, err := url.Parse(uri)
	if err != nil {
		return "", errors.Wrap(err, "parsing LNURL callback")
	}

	amountMsat := amountSat * 1000
	query := callback.Query()
	query.Set("amount", strconv.FormatInt(amountMsat, 10))
	callback.RawQuery = query.Encode()

	resp, err := client.Get(callback.String())
	if err != nil {
		return "", errors.Wrapf(err, "calling %s", callback)
	}

	var params lnurl.LNURLPayParams
	if err := json.NewDecoder(resp.Body).Decode(&params); err != nil {
		return "", errors.Wrap(err, "decoding LNURL params")
	}
	defer resp.Body.Close()

	if params.Status == "ERROR" {
		return "", errors.New(params.Reason)
	}

	if amountMsat < params.MinSendable {
		return "", errors.Errorf("amount %d is lower than the minimum allowed %d",
			amountSat,
			params.MinSendable/1000,
		)
	}

	if amountMsat > params.MaxSendable {
		return "", errors.Errorf("amount %d is higher than the maximum allowed %d",
			amountSat,
			params.MaxSendable/1000,
		)
	}

	return params.Callback, nil
}

func getInvoice(client *http.Client, callback string) (string, error) {
	resp, err := client.Get(callback)
	if err != nil {
		return "", errors.Wrapf(err, "calling %s", callback)
	}

	var values lnurl.LNURLPayValues
	if err := json.NewDecoder(resp.Body).Decode(&values); err != nil {
		return "", errors.Wrap(err, "decoding LNURL values")
	}
	defer resp.Body.Close()

	return values.PR, nil
}
