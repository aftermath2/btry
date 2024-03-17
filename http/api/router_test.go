package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/http/api"
	"github.com/aftermath2/BTRY/lightning"
	"github.com/aftermath2/BTRY/logger"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	"github.com/stretchr/testify/assert"
)

func TestRouter(t *testing.T) {
	rateLimiterTokens := uint64(5)
	apiConfig := config.API{
		Logger: config.Logger{
			Level: uint8(logger.DISABLED),
		},
		RateLimiter: config.RateLimiter{
			Tokens:   rateLimiterTokens,
			Interval: 60,
		},
		SSE: config.SSE{
			Logger: config.Logger{
				Level: uint8(logger.DISABLED),
			},
		},
	}
	winnersCh := make(chan []db.Winner)
	blocksCh := make(chan *chainrpc.BlockEpoch)
	lndMock := lightning.NewClientMock()

	lndMock.On("SubscribeBlocks", context.Background()).
		Return(lightning.BlockedStreamMock[*chainrpc.BlockEpoch]{}, nil)
	lndMock.On("SubscribeChannelEvents", context.Background()).
		Return(lightning.BlockedStreamMock[*lnrpc.ChannelEventUpdate]{}, nil)
	lndMock.On("SubscribeInvoices", context.Background()).
		Return(lightning.BlockedStreamMock[*lnrpc.Invoice]{}, nil)
	lndMock.On("SubscribePayments", context.Background()).
		Return(lightning.BlockedStreamMock[*lnrpc.Payment]{}, nil)

	handler, err := api.NewRouter(apiConfig, &db.DB{}, lndMock, winnersCh, blocksCh)
	assert.NoError(t, err)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	res, err := srv.Client().Get(srv.URL)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
}
