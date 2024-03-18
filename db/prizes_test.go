package db_test

import (
	"database/sql"
	"testing"

	database "github.com/aftermath2/BTRY/db"

	"github.com/stretchr/testify/suite"
)

type PrizesSuite struct {
	suite.Suite

	db database.PrizesStore
}

func TestPrizesSuite(t *testing.T) {
	suite.Run(t, &PrizesSuite{})
}

func (p *PrizesSuite) SetupTest() {
	db := setupDB(p.T(), func(db *sql.DB) {
		lotteriesQuery := `INSERT INTO lotteries (height) VALUES (?)`
		_, err := db.Exec(lotteriesQuery, lotteryHeight)
		p.NoError(err)
		query := `DELETE FROM prizes;
		INSERT INTO prizes (public_key, amount, lottery_height) VALUES (?, ?, ?);`
		_, err = db.Exec(query, testWinner.PublicKey, testWinner.Prize, lotteryHeight)
		p.NoError(err)
	})
	p.db = db.Prizes
}

func (p *PrizesSuite) TestPrizesAccumulation() {
	winner := database.Winner{
		PublicKey: "pubKey",
		Prize:     2016,
		Ticket:    21,
	}
	err := p.db.Set(lotteryHeight, []database.Winner{winner, winner, winner})
	p.NoError(err)

	prizes, err := p.db.Get(winner.PublicKey)
	p.NoError(err)

	p.Equal(winner.Prize*3, prizes)
}

func (p *PrizesSuite) TestExpire() {
	height := uint32(2)
	expiredAmount, err := p.db.Expire(height)
	p.NoError(err)

	p.Equal(testWinner.Prize, expiredAmount)

	prizes, err := p.db.Get(testWinner.PublicKey)
	p.NoError(err)

	p.Zero(prizes)
}

func (p *PrizesSuite) TestGet() {
	prizes, err := p.db.Get(testWinner.PublicKey)
	p.NoError(err)

	p.Equal(testWinner.Prize, prizes)
}

func (p *PrizesSuite) TestGetMany() {
	err := p.db.Set(lotteryHeight, []database.Winner{testWinner2, testWinner2})
	p.NoError(err)

	prizes, err := p.db.Get(testWinner2.PublicKey)
	p.NoError(err)

	p.Equal(testWinner2.Prize*2, prizes)
}

func (p *PrizesSuite) TestSet() {
	err := p.db.Set(1, []database.Winner{testWinner2, testWinner2})
	p.NoError(err)

	prizes, err := p.db.Get(testWinner2.PublicKey)
	p.NoError(err)

	p.Equal(testWinner2.Prize*2, prizes)
}

func (p *PrizesSuite) TestWithdraw() {
	amount := uint64(20)
	err := p.db.Withdraw(testWinner.PublicKey, amount)
	p.NoError(err)

	prizes, err := p.db.Get(testWinner.PublicKey)
	p.NoError(err)

	p.Equal(testWinner.Prize-amount, prizes)
}

func (p *PrizesSuite) TestWithdrawMultiplePrizes() {
	winner := database.Winner{
		PublicKey: "17dc39e569bbeab0b1a1e2da5198d217c855fe5041a0b04f94030fdaf15c0bcd",
		Prize:     100,
		Ticket:    1,
	}
	err := p.db.Set(1, []database.Winner{winner, winner, winner})
	p.NoError(err)

	amount := uint64(200)
	err = p.db.Withdraw(winner.PublicKey, amount)
	p.NoError(err)

	prizes, err := p.db.Get(winner.PublicKey)
	p.NoError(err)

	p.Equal((winner.Prize*3)-amount, prizes)
}

func (p *PrizesSuite) TestWithdrawInsufficientPrizes() {
	amount := testWinner.Prize + 125
	err := p.db.Withdraw(testWinner.PublicKey, amount)
	p.Error(err)
	p.ErrorIs(err, database.ErrInsufficientPrizes)
}

func (p *PrizesSuite) TestUpdatePrizes() {
	cases := []struct {
		err            error
		desc           string
		prizes         []*database.PrizesRow
		expectedPrizes []*database.PrizesRow
		amount         uint64
	}{
		{
			desc:   "No amount",
			amount: 0,
			prizes: []*database.PrizesRow{
				{RowID: 2, Amount: 10},
				{RowID: 1, Amount: 20},
			},
			expectedPrizes: []*database.PrizesRow{
				{RowID: 2, Amount: 10},
				{RowID: 1, Amount: 20},
			},
		},
		{
			desc:   "Amount equal to prizes",
			amount: 60,
			prizes: []*database.PrizesRow{
				{RowID: 3, Amount: 10},
				{RowID: 2, Amount: 20},
				{RowID: 1, Amount: 30},
			},
			expectedPrizes: []*database.PrizesRow{
				{RowID: 3, Amount: 0},
				{RowID: 2, Amount: 0},
				{RowID: 1, Amount: 0},
			},
		},
		{
			desc:   "Amount lower than prizes",
			amount: 20,
			prizes: []*database.PrizesRow{
				{RowID: 3, Amount: 10},
				{RowID: 2, Amount: 20},
				{RowID: 1, Amount: 30},
			},
			expectedPrizes: []*database.PrizesRow{
				{RowID: 3, Amount: 0},
				{RowID: 2, Amount: 10},
				{RowID: 1, Amount: 30},
			},
		},
		{
			desc:   "Amount higher than prizes",
			amount: 20,
			prizes: []*database.PrizesRow{
				{Amount: 10},
			},
			err: database.ErrInsufficientPrizes,
		},
		{
			desc:   "Insufficient prizes",
			amount: 50,
			prizes: []*database.PrizesRow{
				{Amount: 10},
				{Amount: 20},
			},
			err: database.ErrInsufficientPrizes,
		},
		{
			desc:   "No prizes",
			amount: 50,
			prizes: []*database.PrizesRow{},
			err:    database.ErrInsufficientPrizes,
		},
	}

	for _, tc := range cases {
		p.Run(tc.desc, func() {
			err := database.UpdatePrizes(tc.amount, tc.prizes)
			p.Equal(tc.err, err)

			if err == nil {
				p.Equal(tc.expectedPrizes, tc.prizes)
			}
		})
	}
}
