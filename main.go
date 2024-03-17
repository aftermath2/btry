// Main package
package main

import (
	"context"
	"log"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/http/api"
	"github.com/aftermath2/BTRY/http/server"
	"github.com/aftermath2/BTRY/lightning"
	"github.com/aftermath2/BTRY/lottery"
	"github.com/aftermath2/BTRY/notification"
	"github.com/aftermath2/BTRY/tor"

	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	_ "modernc.org/sqlite"
)

func main() {
	config, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	lnd, err := lightning.NewClient(config.Lightning)
	if err != nil {
		log.Fatal(err)
	}

	winnersCh := make(chan []db.Winner)
	blocksCh := make(chan *chainrpc.BlockEpoch)

	db, err := db.Open(config.DB)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	torClient, err := tor.NewClient(config.Tor)
	if err != nil {
		log.Fatal(err)
	}

	notifier, err := notification.NewNotifier(config.Notifier, db, torClient)
	if err != nil {
		log.Fatal(err)
	}
	go notifier.GetUpdates()

	lottery, err := lottery.New(config.Lottery, db, lnd, notifier, winnersCh, blocksCh)
	if err != nil {
		log.Fatal(err)
	}

	if err := lottery.Start(); err != nil {
		log.Fatal(err)
	}

	router, err := api.NewRouter(config.API, db, lnd, winnersCh, blocksCh)
	if err != nil {
		log.Fatal(err)
	}

	server, err := server.New(config.Server, router)
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
