package db_test

import (
	"database/sql"
	"sort"
	"testing"

	database "github.com/aftermath2/BTRY/db"

	"github.com/stretchr/testify/suite"
)

const (
	firstHeight  uint32 = 1
	secondHeight uint32 = 145
)

type LotteriesSuite struct {
	suite.Suite

	db database.LotteriesStore
}

func TestLotteriesSuite(t *testing.T) {
	suite.Run(t, &LotteriesSuite{})
}

func (l *LotteriesSuite) SetupTest() {
	db := setupDB(l.T(), func(db *sql.DB) {
		query := `DELETE FROM lotteries;
		INSERT INTO lotteries (height) VALUES (?), (?);`
		_, err := db.Exec(query, firstHeight, secondHeight)
		l.NoError(err)
	})
	l.db = db.Lotteries
}

func (l *LotteriesSuite) TestAddHeight() {
	thirdHeight := secondHeight + 144
	err := l.db.AddHeight(thirdHeight)
	l.NoError(err)

	heights, err := l.db.ListHeights(0, 0, false)
	l.NoError(err)

	l.Len(heights, 3)
	l.Equal(firstHeight, heights[0])
	l.Equal(secondHeight, heights[1])
	l.Equal(thirdHeight, heights[2])
}

func (l *LotteriesSuite) TestGetNextHeight() {
	nextHeight, err := l.db.GetNextHeight()
	l.NoError(err)

	l.Equal(secondHeight, nextHeight)
}

func (l *LotteriesSuite) TestListHeights() {
	cases := []struct {
		desc     string
		expected []uint32
		offset   uint64
		limit    uint64
		reverse  bool
	}{
		{
			desc:     "Forward",
			expected: []uint32{firstHeight, secondHeight},
		},
		{
			desc:     "Offset",
			offset:   uint64(firstHeight),
			expected: []uint32{secondHeight},
		},
		{
			desc:     "Limit",
			limit:    1,
			expected: []uint32{firstHeight},
		},
		{
			desc:     "Reverse",
			reverse:  true,
			expected: []uint32{secondHeight, firstHeight},
		},
		{
			desc:     "Reverse and offset",
			reverse:  true,
			offset:   uint64(secondHeight),
			expected: []uint32{firstHeight},
		},
		{
			desc:     "Reverse and limit",
			limit:    1,
			reverse:  true,
			expected: []uint32{secondHeight},
		},
	}

	for _, tc := range cases {
		l.Run(tc.desc, func() {
			heights, err := l.db.ListHeights(tc.offset, tc.limit, tc.reverse)
			l.NoError(err)

			l.Equal(tc.expected, heights)

			ok := sort.SliceIsSorted(heights, func(i, j int) bool {
				if tc.reverse {
					return heights[i] > heights[j]
				}
				return heights[i] < heights[j]
			})
			l.True(ok)
		})
	}
}
