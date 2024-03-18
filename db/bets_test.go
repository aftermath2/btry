package db_test

import (
	"database/sql"
	"sort"
	"testing"

	database "github.com/aftermath2/BTRY/db"

	"github.com/stretchr/testify/suite"
)

var (
	firstBet = database.Bet{
		Index:     15,
		PublicKey: "7d959d6d552c7d38b3ecafb72805fa03a6dee6b7f0c5f63f57a371736cb004b1",
		Tickets:   15,
	}
	secondBet = database.Bet{
		Index:     33,
		PublicKey: "876baf90c3d2d26c04ba1d208c29605b2c6fd13fbb3f6b46cf7f10ece3dac69d",
		Tickets:   18,
	}
)

type BetsSuite struct {
	suite.Suite

	db database.BetsStore
}

func TestBetsSuite(t *testing.T) {
	suite.Run(t, &BetsSuite{})
}

func (b *BetsSuite) SetupTest() {
	db := setupDB(b.T(), func(db *sql.DB) {
		lotteriesQuery := `INSERT INTO lotteries (height) VALUES (?)`
		_, err := db.Exec(lotteriesQuery, lotteryHeight)
		b.NoError(err)
		query := `DELETE FROM bets;
		INSERT INTO bets (idx, public_key, tickets, lottery_height) VALUES (?,?,?,?), (?,?,?,?);`
		_, err = db.Exec(query, firstBet.Index, firstBet.PublicKey, firstBet.Tickets, lotteryHeight,
			secondBet.Index, secondBet.PublicKey, secondBet.Tickets, lotteryHeight)
		b.NoError(err)
	})
	b.db = db.Bets
}

func (b *BetsSuite) TestAdd() {
	bet := database.Bet{
		PublicKey: "99c1c20dfa84ca4a475359d8f6e711cb42923c7d2792b9e278766cb4801b8917",
		Tickets:   10,
	}
	err := b.db.Add(bet)
	b.NoError(err)

	bets, err := b.db.List(lotteryHeight, 0, 0, false)
	b.NoError(err)

	b.Len(bets, 3)
	b.Equal(bet.PublicKey, bets[2].PublicKey)
	expectedIndex := firstBet.Tickets + secondBet.Tickets + bet.Tickets
	b.Equal(expectedIndex, bets[2].Index)
	b.Equal(bet.Tickets, bets[2].Tickets)
}

func (b *BetsSuite) TestGetPrizePool() {
	prizePool, err := b.db.GetPrizePool(lotteryHeight)
	b.NoError(err)

	expectedPrizePool := firstBet.Tickets + secondBet.Tickets
	b.Equal(expectedPrizePool, prizePool)
}

func (b *BetsSuite) TestList() {
	cases := []struct {
		desc     string
		expected []database.Bet
		offset   uint64
		limit    uint64
		reverse  bool
	}{
		{
			desc:     "Forward",
			expected: []database.Bet{firstBet, secondBet},
		},
		{
			desc:     "Offset",
			offset:   firstBet.Index,
			expected: []database.Bet{secondBet},
		},
		{
			desc:     "Limit",
			limit:    1,
			expected: []database.Bet{firstBet},
		},
		{
			desc:     "Reverse",
			reverse:  true,
			expected: []database.Bet{secondBet, firstBet},
		},
		{
			desc:     "Reverse and offset",
			reverse:  true,
			offset:   secondBet.Index,
			expected: []database.Bet{firstBet},
		},
		{
			desc:     "Reverse and limit",
			limit:    1,
			reverse:  true,
			expected: []database.Bet{secondBet},
		},
	}

	for _, tc := range cases {
		b.Run(tc.desc, func() {
			bets, err := b.db.List(lotteryHeight, tc.offset, tc.limit, tc.reverse)
			b.NoError(err)

			b.Equal(tc.expected, bets)

			ok := sort.SliceIsSorted(bets, func(i, j int) bool {
				if tc.reverse {
					return bets[i].Tickets > bets[j].Tickets
				}
				return bets[i].Tickets < bets[j].Tickets
			})
			b.True(ok)
		})
	}
}
