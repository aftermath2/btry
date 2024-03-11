package lottery

import (
	"context"
	"database/sql"
	"encoding/hex"
	"math"
	"os"
	"testing"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/lightning"
	"github.com/aftermath2/BTRY/notification"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

var bets = []db.Bet{
	{
		Index:     427_224,
		PublicKey: "1",
		Tickets:   427_224,
	},
	{
		Index:     1_427_224,
		PublicKey: "2",
		Tickets:   1_000_000,
	},
}

func TestStart(t *testing.T) {
	nextHeight := uint32(900_000)
	blocksDuration := uint32(144)

	config := config.Lottery{
		Duration: blocksDuration,
	}

	betsMock := db.NewBetsStoreMock()
	betsMock.On("List", uint64(0), uint64(0), false).Return([]db.Bet{}, nil)

	lotteryMock := db.NewLotteriesStoreMock()
	lotteryMock.On("GetNextHeight").Return(nextHeight, nil)
	lotteryMock.On("AddHeight", nextHeight+blocksDuration).Return(nil)

	db := &db.DB{
		Bets:      betsMock,
		Lotteries: lotteryMock,
	}

	lnd := lightning.NewClientMock()
	info := &lnrpc.GetInfoResponse{BlockHeight: 843_204}
	lnd.On("GetInfo", context.Background()).Return(info, nil)

	blocksCh := make(chan *chainrpc.BlockEpoch)

	lottery, err := New(config, db, lnd, nil, nil, blocksCh)
	assert.NoError(t, err)

	go func() {
		blockHash, err := hex.DecodeString("000000000000000000003bc0544004a6e74beb66b21b1e564eb81dbd478d67c6")
		assert.NoError(t, err)
		blocksCh <- &chainrpc.BlockEpoch{
			Hash:   blockHash,
			Height: nextHeight,
		}
	}()

	err = lottery.Start()
	assert.NoError(t, err)
}

func TestStartNoNextHeight(t *testing.T) {
	nextHeight := uint32(0)
	blockHeight := uint32(843_204)
	blocksDuration := uint32(32)

	config := config.Lottery{
		Duration: blocksDuration,
	}

	lotteryMock := db.NewLotteriesStoreMock()
	lotteryMock.On("GetNextHeight").Return(nextHeight, nil)
	lotteryMock.On("AddHeight", blockHeight+blocksDuration).Return(nil)
	db := &db.DB{
		Lotteries: lotteryMock,
	}

	lnd := lightning.NewClientMock()
	info := &lnrpc.GetInfoResponse{BlockHeight: blockHeight}
	lnd.On("GetInfo", context.Background()).Return(info, nil)

	lottery, err := New(config, db, lnd, nil, nil, nil)
	assert.NoError(t, err)

	err = lottery.Start()
	assert.NoError(t, err)
}

func TestStartPastHeight(t *testing.T) {
	nextHeight := uint32(843_199)
	blockHeight := uint32(843_204)
	blocksDuration := uint32(144)

	config := config.Lottery{
		Duration: blocksDuration,
	}

	lotteryMock := db.NewLotteriesStoreMock()
	lotteryMock.On("GetNextHeight").Return(nextHeight, nil)
	lotteryMock.On("AddHeight", blockHeight+blocksDuration).Return(nil)
	lotteryMock.On("DeleteHeight", nextHeight).Return(nil)
	db := &db.DB{
		Lotteries: lotteryMock,
	}

	lnd := lightning.NewClientMock()
	info := &lnrpc.GetInfoResponse{BlockHeight: blockHeight}
	lnd.On("GetInfo", context.Background()).Return(info, nil)

	lottery, err := New(config, db, lnd, nil, nil, nil)
	assert.NoError(t, err)

	err = lottery.Start()
	assert.NoError(t, err)
}

func TestRaffle(t *testing.T) {
	winnersCh := make(chan []db.Winner)
	blocksCh := make(<-chan *chainrpc.BlockEpoch)
	db := setupDB(t, func(db *sql.DB) {
		_, err := db.Exec("INSERT INTO bets (idx, tickets, public_key) VALUES (?, ?, ?), (?, ?, ?)",
			bets[0].Index, bets[0].Tickets, bets[0].PublicKey,
			bets[1].Index, bets[1].Tickets, bets[1].PublicKey,
		)
		assert.NoError(t, err)
	})
	notifierMock := notification.NewNotifierMock()

	lottery, err := New(config.Lottery{}, db, nil, notifierMock, winnersCh, blocksCh)
	assert.NoError(t, err)

	prizePool, err := db.Bets.GetPrizePool()
	assert.NoError(t, err)

	t.Run("Winners were sent through the channel", func(t *testing.T) {
		go func() {
			winners := <-winnersCh
			assert.Len(t, winners, len(prizes))
		}()
	})

	blockHash, err := hex.DecodeString("000000000000000000003bc0544004a6e74beb66b21b1e564eb81dbd478d67c6")
	assert.NoError(t, err)

	block := &chainrpc.BlockEpoch{
		Hash:   blockHash,
		Height: 833348,
	}

	err = lottery.raffle(block)
	assert.NoError(t, err)

	t.Run("Bets were reset", func(t *testing.T) {
		bets, err := db.Bets.List(0, 0, false)
		assert.NoError(t, err)

		assert.Len(t, bets, 0)
	})

	t.Run("Winners prizes were correctly assigned", func(t *testing.T) {
		winners, err := db.Winners.List(block.Height)
		assert.NoError(t, err)

		assert.LessOrEqual(t, len(winners), 2)

		givenPrizes := uint64(0)
		for _, winner := range winners {
			givenPrizes += winner.Prizes
			if winner.PublicKey != bets[0].PublicKey && winner.PublicKey != bets[1].PublicKey {
				assert.Failf(t, "A winner that did not have bets was assigned: %s", winner.PublicKey)
			}
		}

		fee := float64(prizePool) * (btryFee / 100)
		assert.Equal(t, math.Round(float64(prizePool)-fee), float64(givenPrizes))
	})
}

func TestRaffleWithoutBets(t *testing.T) {
	db := setupDB(t, nil)
	lottery, err := New(config.Lottery{}, db, nil, nil, nil, nil)
	assert.NoError(t, err)

	block := &chainrpc.BlockEpoch{
		Height: 0,
	}

	err = lottery.raffle(block)
	assert.NoError(t, err)

	winners, err := db.Winners.List(block.Height)
	assert.NoError(t, err)

	assert.Empty(t, winners)
}

func TestGetWinners(t *testing.T) {
	lottery, err := New(config.Lottery{}, &db.DB{}, nil, nil, nil, nil)
	assert.NoError(t, err)

	prizePool := uint64(1_427_224)
	blockHash, err := hex.DecodeString("000000000000000000003bc0544004a6e74beb66b21b1e564eb81dbd478d67c6")
	assert.NoError(t, err)

	winners, err := lottery.getWinners(blockHash, prizePool, bets)
	assert.NoError(t, err)

	assert.Len(t, winners, len(prizes))

	for i, winner := range winners {
		validateGetWinner(t, winner.Ticket, winner.PublicKey)

		prize := (prizes[i] / 100) * float64(prizePool)
		assert.Equal(t, uint64(math.Round(prize)), winner.Prizes)
		assert.False(t, winner.Expired)
	}
}

func TestGetWinnersWithoutBets(t *testing.T) {
	lottery, err := New(config.Lottery{}, &db.DB{}, nil, nil, nil, nil)
	assert.NoError(t, err)

	blockHash, err := hex.DecodeString("000000000000000000003bc0544004a6e74beb66b21b1e564eb81dbd478d67c6")
	assert.NoError(t, err)
	winners, err := lottery.getWinners(blockHash, 0, []db.Bet{})
	assert.NoError(t, err)

	assert.Nil(t, winners)
}

func TestGetWinningTickets(t *testing.T) {
	prizePool := uint64(1000)
	results := []uint64{417, 777, 865, 833, 977, 402, 322, 337}
	blockHash, err := hex.DecodeString("000000000000000000001badcbb5d10b486a18a97ac9d6e08d526a62aa9a360e")
	assert.NoError(t, err)
	i := len(blockHash) - 1

	for _, expected := range results {
		target := getWinningTicket(blockHash, i, prizePool)
		assert.Equal(t, expected, target)
		i -= 2
	}
}

func TestPercentages(t *testing.T) {
	// Just in case :)
	total := btryFee
	for _, prize := range prizes {
		total += prize
	}
	assert.Equal(t, 100, int(total))
}

func TestGetInfo(t *testing.T) {
	lndMock := lightning.NewClientMock()
	betsMock := db.NewBetsStoreMock()
	lotteriesMock := db.NewLotteriesStoreMock()
	db := &db.DB{
		Bets:      betsMock,
		Lotteries: lotteriesMock,
	}

	remoteBalance := int64(15_000_000)
	prizePool := uint64(5_000_000)
	nextHeight := uint32(1)

	ctx := context.Background()
	lndMock.On("RemoteBalance", ctx).Return(remoteBalance, nil)
	betsMock.On("GetPrizePool").Return(prizePool, nil)
	lotteriesMock.On("GetNextHeight").Return(nextHeight, nil)

	info, err := GetInfo(ctx, lndMock, db)
	assert.NoError(t, err)

	assert.Equal(t, int64(prizePool), info.PrizePool)
	assert.Equal(t, remoteBalance/CapacityDivisor, info.Capacity)
	assert.Equal(t, nextHeight, info.NextHeight)
}

func setupDB(t *testing.T, setup func(db *sql.DB)) *db.DB {
	t.Helper()

	file, err := os.CreateTemp("", "*")
	assert.NoError(t, err)

	db, err := db.Open(config.DB{
		Path:   file.Name(),
		Logger: config.Logger{},
	})
	assert.NoError(t, err)

	sqlDB, err := sql.Open("sqlite", file.Name())
	assert.NoError(t, err)

	if setup != nil {
		setup(sqlDB)
	}

	err = sqlDB.Close()
	assert.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, file.Close())
		assert.NoError(t, db.Close())
	})

	return db
}

func validateGetWinner(t *testing.T, target uint64, expectedPubKey string) {
	t.Helper()

	for i, bet := range bets {
		if target <= bet.Tickets {
			if i == 0 || target > bets[i-1].Tickets {
				assert.Equal(t, expectedPubKey, bet.PublicKey)
			}
		}
	}
}
