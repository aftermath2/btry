package middleware

import (
	"hash/crc32"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/aftermath2/BTRY/config"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-limiter/httplimit"
	"github.com/sethvargo/go-limiter/memorystore"
)

// Headers and corresponding parser to look up the real client IP.
// They will be check in order, the first non-empty one will be picked,
// or else the remote address is selected.
// CF-Connecting-IP is a header added by Cloudflare:
// https://support.cloudflare.com/hc/en-us/articles/206776727-What-is-True-Client-IP-
var ipHeaders = []header{
	{parseXForwardedForHeader, "CF-Connecting-IP"},
	{parseXForwardedForHeader, "X-Forwarded-For"},
	{parseForwardedHeader, "Forwarded"},
	{parseXRealIPHeader, "X-Real-IP"},
}

type header struct {
	parser func(string) string
	name   string
}

// NewRateLimiter returns a rate limiter with the configuration values passed.
func NewRateLimiter(config config.RateLimiter) (*httplimit.Middleware, error) {
	store, err := memorystore.New(&memorystore.Config{
		Tokens:   config.Tokens,
		Interval: config.Interval,
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating rate limiter memory store")
	}

	mw, err := httplimit.NewMiddleware(store, keyFunc)
	if err != nil {
		return nil, errors.Wrap(err, "creating rate limiter middleware")
	}

	return mw, nil
}

func keyFunc(r *http.Request) (string, error) {
	// Store a hash of the IP to preserve the privacy of the users
	ip := getClientIP(r)
	hash := crc32.ChecksumIEEE([]byte(ip))
	key := strconv.Itoa(int(hash))
	return key, nil
}

func getClientIP(r *http.Request) string {
	ip := r.RemoteAddr

	for _, header := range ipHeaders {
		value := r.Header.Get(header.name)
		if value != "" {
			parsedIP := header.parser(value)
			if parsedIP != "" {
				ip = parsedIP
				break
			}
		}
	}

	if strings.Contains(ip, ":") {
		host, _, err := net.SplitHostPort(ip)
		if err != nil {
			return ip
		}

		return host
	}

	return ip
}

func parseForwardedHeader(value string) string {
	parts := strings.Split(value, ",")
	parts = strings.Split(parts[0], ";")

	for _, part := range parts {
		kv := strings.Split(part, "=")

		if len(kv) == 2 {
			k := strings.ToLower(strings.TrimSpace(kv[0]))
			v := strings.TrimSpace(kv[1])

			if k == "for" {
				return v
			}
		}
	}

	return ""
}

func parseXForwardedForHeader(value string) string {
	parts := strings.Split(value, ",")
	return strings.TrimSpace(parts[0])
}

func parseXRealIPHeader(value string) string {
	return value
}
