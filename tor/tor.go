package tor

import (
	"net/http"
	"net/url"

	"github.com/aftermath2/BTRY/config"

	"github.com/pkg/errors"
)

// NewClient returns an http.Client proxied through Tor.
func NewClient(config config.Tor) (*http.Client, error) {
	proxy, err := url.Parse("socks5://" + config.Address)
	if err != nil {
		return nil, errors.Wrap(err, "parsing tor address")
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy),
		},
		Timeout: config.Timeout,
	}, nil
}
