package db_test

import (
	"database/sql"
	"testing"

	database "github.com/aftermath2/BTRY/db"

	"github.com/stretchr/testify/suite"
)

var (
	testWinner = database.Winner{
		PublicKey: "7d959d6d552c7d38b3ecafb72805fa03a6dee6b7f0c5f63f57a371736cb004b1",
		Prizes:    75,
		Ticket:    21,
	}

	testWinner2 = database.Winner{
		PublicKey: "876baf90c3d2d26c04ba1d208c29605b2c6fd13fbb3f6b46cf7f10ece3dac69d",
		Prizes:    12,
		Ticket:    23,
	}
)

const height uint32 = 1

type WinnersSuite struct {
	suite.Suite

	db database.WinnersStore
}

func TestWinnersSuite(t *testing.T) {
	suite.Run(t, &WinnersSuite{})
}

func (w *WinnersSuite) SetupTest() {
	db := setupDB(w.T(), func(db *sql.DB) {
		lotteriesQuery := `INSERT INTO lotteries (height) VALUES (?)`
		_, err := db.Exec(lotteriesQuery, height)
		w.NoError(err)
		query := `DELETE FROM winners;
		INSERT INTO winners (public_key, prizes, ticket, lottery_height) VALUES (?, ?, ?, ?);`
		_, err = db.Exec(query, testWinner.PublicKey, testWinner.Prizes, testWinner.Ticket, height)
		w.NoError(err)
	})
	w.db = db.Winners
}

func (w *WinnersSuite) TestAddWinners() {
	winner := database.Winner{
		PublicKey: "pubKey",
		Prizes:    2016,
		Ticket:    21,
	}
	err := w.db.Add(height, []database.Winner{winner})
	w.NoError(err)
}

func (w *WinnersSuite) TestAddWinnersPrizesAccumulation() {
	winner := database.Winner{
		PublicKey: "pubKey",
		Prizes:    2016,
		Ticket:    21,
	}
	err := w.db.Add(height, []database.Winner{winner, winner, winner})
	w.NoError(err)

	prizes, err := w.db.GetPrizes(winner.PublicKey)
	w.NoError(err)

	w.Equal(winner.Prizes*3, prizes)
}

func (w *WinnersSuite) TestClaimPrizes() {
	amount := uint64(20)
	err := w.db.ClaimPrizes(testWinner.PublicKey, amount)
	w.NoError(err)

	winners, err := w.db.List(height)
	w.NoError(err)

	w.Equal(testWinner.Prizes-amount, winners[0].Prizes)
}

func (w *WinnersSuite) TestClaimPrizesInsufficientPrizes() {
	amount := testWinner.Prizes + 125
	err := w.db.ClaimPrizes(testWinner.PublicKey, amount)
	w.Error(err)
	w.ErrorIs(err, database.ErrInsufficientPrizes)
}

func (w *WinnersSuite) TestExpirePrizes() {
	height := uint32(2)
	expiredAmount, err := w.db.ExpirePrizes(height)
	w.NoError(err)

	w.Equal(testWinner.Prizes, expiredAmount)

	prizes, err := w.db.GetPrizes(testWinner.PublicKey)
	w.NoError(err)

	w.Zero(prizes)
}

func (w *WinnersSuite) TestGetPrizes() {
	prizes, err := w.db.GetPrizes(testWinner.PublicKey)
	w.NoError(err)

	w.Equal(testWinner.Prizes, prizes)
}

func (w *WinnersSuite) TestGetMultiplePrizes() {
	err := w.db.Add(height, []database.Winner{testWinner2, testWinner2})
	w.NoError(err)

	prizes, err := w.db.GetPrizes(testWinner2.PublicKey)
	w.NoError(err)

	w.Equal(testWinner2.Prizes*2, prizes)
}

func (w *WinnersSuite) TestList() {
	winners, err := w.db.List(height)
	w.NoError(err)

	w.Len(winners, 1)
	w.Equal(testWinner, winners[0])
}
