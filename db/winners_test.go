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
		Prize:     75,
		Ticket:    21,
	}

	testWinner2 = database.Winner{
		PublicKey: "876baf90c3d2d26c04ba1d208c29605b2c6fd13fbb3f6b46cf7f10ece3dac69d",
		Prize:     12,
		Ticket:    23,
	}
)

const lotteryHeight uint32 = 1

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
		_, err := db.Exec(lotteriesQuery, lotteryHeight)
		w.NoError(err)
		query := `DELETE FROM winners;
		INSERT INTO winners (public_key, prize, ticket, lottery_height) VALUES (?, ?, ?, ?);`
		_, err = db.Exec(query, testWinner.PublicKey, testWinner.Prize, testWinner.Ticket, lotteryHeight)
		w.NoError(err)
	})
	w.db = db.Winners
}

func (w *WinnersSuite) TestAddWinners() {
	winner := database.Winner{
		PublicKey: "pubKey",
		Prize:     2016,
		Ticket:    21,
	}
	err := w.db.Add(lotteryHeight, []database.Winner{winner})
	w.NoError(err)
}

func (w *WinnersSuite) TestList() {
	winners, err := w.db.List(lotteryHeight)
	w.NoError(err)

	w.Len(winners, 1)
	w.Equal(testWinner, winners[0])
}
