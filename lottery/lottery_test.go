package lottery

import (
	"context"
	"database/sql"
	"math"
	"os"
	"testing"
	"time"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/lightning"
	"github.com/aftermath2/BTRY/notification"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

var bets = []db.Bet{
	{
		Index:     2016,
		PublicKey: "1",
		Tickets:   2016,
	},
	{
		Index:     6048,
		PublicKey: "2",
		Tickets:   4032,
	},
}

func TestStart(t *testing.T) {
	config := config.Lottery{
		Time: "12:00",
	}
	lottery, err := New(config, &db.DB{}, nil, nil, nil)
	assert.NoError(t, err)

	lottery.Start()
}

func TestJobFunc(t *testing.T) {
	db := setupDB(t, nil)
	lottery, err := New(config.Lottery{}, db, nil, nil, nil)
	assert.NoError(t, err)

	lottery.jobFunc()
}

func TestRaffle(t *testing.T) {
	winnersCh := make(chan []db.Winner)
	db := setupDB(t, func(db *sql.DB) {
		_, err := db.Exec("INSERT INTO bets (idx, tickets, public_key) VALUES (?, ?, ?), (?, ?, ?)",
			bets[0].Index, bets[0].Tickets, bets[0].PublicKey,
			bets[1].Index, bets[1].Tickets, bets[1].PublicKey,
		)
		assert.NoError(t, err)
	})
	notifierMock := notification.NewNotifierMock()

	lottery, err := New(config.Lottery{}, db, nil, notifierMock, winnersCh)
	assert.NoError(t, err)

	prizePool, err := db.Bets.GetPrizePool()
	assert.NoError(t, err)

	t.Run("Winners were sent through the channel", func(t *testing.T) {
		go func() {
			winners := <-winnersCh
			assert.Len(t, winners, len(prizes))
		}()
	})

	err = lottery.raffle()
	assert.NoError(t, err)

	t.Run("Bets were reset", func(t *testing.T) {
		bets, err := db.Bets.List(0, 0, false)
		assert.NoError(t, err)

		assert.Len(t, bets, 0)
	})

	t.Run("Winners prizes were correctly assigned", func(t *testing.T) {
		winners, err := db.Winners.List()
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
		assert.Equal(t, math.Floor(float64(prizePool)-fee), float64(givenPrizes))
	})

	t.Run("Winners history table was updated", func(t *testing.T) {
		winners, err := db.Winners.ListHistory(0, uint64(time.Now().Unix()+1000))
		assert.NoError(t, err)

		assert.Len(t, winners, len(prizes))

		givenPrizes := uint64(0)
		for _, winner := range winners {
			givenPrizes += winner.Prizes
			if winner.PublicKey != bets[0].PublicKey && winner.PublicKey != bets[1].PublicKey {
				assert.Failf(t, "A winner that did not have bets was assigned: %s", winner.PublicKey)
			}
		}

		fee := float64(prizePool) * (btryFee / 100)
		assert.Equal(t, math.Floor(float64(prizePool)-fee), float64(givenPrizes))
	})
}

func TestRaffleWithoutBets(t *testing.T) {
	db := setupDB(t, nil)
	lottery, err := New(config.Lottery{}, db, nil, nil, nil)
	assert.NoError(t, err)

	err = lottery.raffle()
	assert.NoError(t, err)

	winners, err := db.Winners.List()
	assert.NoError(t, err)

	assert.Empty(t, winners)
}

func TestGetWinners(t *testing.T) {
	lottery, err := New(config.Lottery{}, &db.DB{}, nil, nil, nil)
	assert.NoError(t, err)

	prizePool := uint64(1_000_000)
	now := time.Now()

	winners, err := lottery.getWinners(prizePool, bets)
	assert.NoError(t, err)

	assert.Len(t, winners, len(prizes))

	for i, winner := range winners {
		validateGetWinner(t, winner.Ticket, winner.PublicKey)

		prize := (prizes[i] / 100) * float64(prizePool)
		assert.Equal(t, uint64(prize), winner.Prizes)
		assert.LessOrEqual(t, winner.CreatedAt, now.Add(time.Second).Unix())
	}
}

func TestGetWinnersWithoutBets(t *testing.T) {
	lottery, err := New(config.Lottery{}, &db.DB{}, nil, nil, nil)
	assert.NoError(t, err)

	winners, err := lottery.getWinners(0, []db.Bet{})
	assert.NoError(t, err)

	assert.Nil(t, winners)
}

func TestGetWinner(t *testing.T) {
	target, pubKey, err := getWinner(bets)
	assert.NoError(t, err)

	validateGetWinner(t, target, pubKey)
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
	db := &db.DB{Bets: betsMock}

	remoteBalance := int64(15_000_000)
	prizePool := uint64(5_000_000)

	ctx := context.Background()
	lndMock.On("RemoteBalance", ctx).Return(remoteBalance, nil)

	betsMock.On("GetPrizePool").Return(prizePool, nil)

	info, err := GetInfo(ctx, lndMock, db)
	assert.NoError(t, err)

	assert.Equal(t, int64(prizePool), info.PrizePool)
	assert.Equal(t, remoteBalance/CapacityDivisor, info.Capacity)
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
