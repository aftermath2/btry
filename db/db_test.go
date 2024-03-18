package db_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/db"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestOpen(t *testing.T) {
	file, err := os.CreateTemp("", "*")
	assert.NoError(t, err)
	defer file.Close()

	_, err = db.Open(config.DB{
		Path:   file.Name(),
		Logger: config.Logger{Label: "TEST", Level: 0},
	})
	assert.NoError(t, err)
}

func TestClose(t *testing.T) {
	file, err := os.CreateTemp("", "*")
	assert.NoError(t, err)
	defer file.Close()

	db, err := db.Open(config.DB{
		Path: file.Name(),
	})
	assert.NoError(t, err)

	err = db.Close()
	assert.NoError(t, err)
}

func TestAddPagination(t *testing.T) {
	cases := []struct {
		desc          string
		query         string
		sortField     string
		expectedQuery string
		limit         uint64
		offset        uint64
		reverse       bool
	}{
		{
			desc:          "All",
			query:         "SELECT idx FROM bets",
			offset:        4,
			limit:         2,
			sortField:     "idx",
			reverse:       false,
			expectedQuery: "SELECT idx FROM bets WHERE idx >4 ORDER BY idx ASC LIMIT 2",
		},
		{
			desc:          "All (reverse)",
			query:         "SELECT idx FROM bets",
			offset:        4,
			limit:         2,
			sortField:     "idx",
			reverse:       true,
			expectedQuery: "SELECT idx FROM bets WHERE idx <4 ORDER BY idx DESC LIMIT 2",
		},
		{
			desc:          "Query with WHERE",
			query:         "SELECT idx FROM bets WHERE lottery_height=?",
			offset:        4,
			sortField:     "idx",
			expectedQuery: "SELECT idx FROM bets WHERE lottery_height=? AND idx >4 ORDER BY idx ASC",
		},
		{
			desc:          "Query without WHERE",
			query:         "SELECT idx FROM bets",
			offset:        4,
			sortField:     "idx",
			expectedQuery: "SELECT idx FROM bets WHERE idx >4 ORDER BY idx ASC",
		},
		{
			desc:          "No limit",
			offset:        4,
			sortField:     "idx",
			expectedQuery: " WHERE idx >4 ORDER BY idx ASC",
		},
		{
			desc:          "No offset",
			limit:         2,
			sortField:     "idx",
			expectedQuery: " ORDER BY idx ASC LIMIT 2",
		},
		{
			desc:          "Sort field",
			sortField:     "created_at",
			expectedQuery: " ORDER BY created_at ASC",
		},
		{
			desc:          "Sort field (reverse)",
			sortField:     "created_at",
			reverse:       true,
			expectedQuery: " ORDER BY created_at DESC",
		},
		{
			desc:          "Sort field (invalid characters)",
			sortField:     "created_at; --",
			expectedQuery: "",
		},
		{
			desc:          "Sort field (invalid characters 2)",
			sortField:     "idx%%\n",
			expectedQuery: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			gotQuery := db.AddPagination(tc.query, tc.offset, tc.limit, tc.sortField, tc.reverse)

			assert.Equal(t, tc.expectedQuery, gotQuery)
		})
	}
}

func TestBulkInsert(t *testing.T) {
	cases := []struct {
		desc          string
		expectedQuery string
		rows          int
		values        int
	}{
		{
			desc:          "1-1",
			rows:          1,
			values:        1,
			expectedQuery: "(?)",
		},
		{
			desc:          "1-2",
			rows:          1,
			values:        2,
			expectedQuery: "(?,?)",
		},
		{
			desc:          "3-2",
			rows:          3,
			values:        2,
			expectedQuery: "(?,?),(?,?),(?,?)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			gotQuery := db.BulkInsertValues(tc.rows, tc.values)
			assert.Equal(t, tc.expectedQuery, gotQuery)
		})
	}
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

	setup(sqlDB)

	err = sqlDB.Close()
	assert.NoError(t, err)

	t.Cleanup(func() {
		err := file.Close()
		assert.NoError(t, err)
		err = db.Close()
		assert.NoError(t, err)
	})

	return db
}
