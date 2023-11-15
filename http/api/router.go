package api

import (
	"io"
	"net/http"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/http/api/handler"
	"github.com/aftermath2/BTRY/http/api/middleware"
	"github.com/aftermath2/BTRY/http/api/sse"
	"github.com/aftermath2/BTRY/lightning"
	"github.com/aftermath2/BTRY/ui"

	"github.com/go-chi/chi/v5"
)

// Router implements http.Handler and io.Closer interfaces.
type Router interface {
	http.Handler
	io.Closer
}

type router struct {
	mux           *chi.Mux
	eventStreamer sse.Streamer
}

// NewRouter returns an HTTP request multiplexer.
func NewRouter(
	config config.API,
	db *db.DB,
	lnd lightning.Client,
	winnersChannel chan []db.Winner,
) (Router, error) {
	rateLimiter, err := middleware.NewRateLimiter(config.RateLimiter)
	if err != nil {
		return nil, err
	}

	loggerMw, err := middleware.NewLogger(config.Logger)
	if err != nil {
		return nil, err
	}

	eventStreamer, err := sse.NewStreamer(config.SSE, db, lnd, winnersChannel)
	if err != nil {
		return nil, err
	}

	mux := chi.NewRouter()
	mux.Use(rateLimiter.Handle, middleware.Cors)

	uiFs, err := ui.FS()
	if err != nil {
		return nil, err
	}
	fs := http.FileServer(http.FS(uiFs))
	mux.Mount("/", fs)
	// Temporary workaround to handle refreshes
	mux.Get("/bet", redirectRoot)
	mux.Get("/bets", redirectRoot)
	mux.Get("/winners", redirectRoot)
	mux.Get("/withdraw", redirectRoot)
	mux.Get("/faq", redirectRoot)

	handler := handler.New(lnd, db, eventStreamer)
	mux.Route("/api", func(r chi.Router) {
		r.Use(rateLimiter.Handle, middleware.Cors, loggerMw.Log)

		r.Get("/bets", handler.GetBets)
		r.Handle("/events", eventStreamer)
		r.Get("/lottery", handler.GetLottery)
		r.Get("/invoice", handler.GetInvoice)
		r.Get("/lnurl/withdraw", handler.LNURLWithdraw)
		r.Get("/prizes", handler.GetPrizes)
		r.Get("/winners", handler.GetWinners)
		r.Get("/winners/history", handler.GetWinnersHistory)
		r.Post("/withdraw", handler.Withdraw)
	})

	return &router{
		mux:           mux,
		eventStreamer: eventStreamer,
	}, nil
}

func (rr *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rr.mux.ServeHTTP(w, r)
}

func (rr *router) Close() error {
	return rr.eventStreamer.Close()
}

func redirectRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusPermanentRedirect)
}
