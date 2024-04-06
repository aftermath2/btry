// Package db contains the logic necessary to store BTRY's information.
package db

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/logger"

	"github.com/pkg/errors"
)

// DB represents the application database.
type DB struct {
	db            *sql.DB
	Bets          BetsStore
	Lightning     LightningStore
	Lotteries     LotteriesStore
	Notifications NotificationsStore
	Prizes        PrizesStore
	Winners       WinnersStore
}

// Open opens the database.
func Open(config config.DB) (*DB, error) {
	logger, err := logger.New(config.Logger)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", config.Path+"?_pragma=busy_timeout=5000")
	if err != nil {
		return nil, errors.Wrap(err, "opening database")
	}
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if _, err := db.Exec(migrations); err != nil {
		return nil, errors.Wrap(err, "executing migrations")
	}

	return &DB{
		db:            db,
		Bets:          newBetsStore(db, logger),
		Lightning:     newLightningStore(db, logger),
		Lotteries:     newLotteriesStore(db, logger),
		Notifications: newNotificationsStore(db, logger),
		Prizes:        newPrizesStore(db, logger),
		Winners:       newWinnersStore(db, logger),
	}, nil
}

// Close releases all related resources.
func (db *DB) Close() error {
	return db.db.Close()
}

// AddPagination returns the pagination part of an SQL query.
func AddPagination(query string, offset, limit uint64, sortField string, reverse bool) string {
	// Safety mechanism to avoid SQL injections (although sortField should always be hardcoded),
	// parameterized column names are not supported
	if strings.ContainsAny(sortField, ";'\"%\n\t\r\b-/*\\") {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(query)

	if offset > 0 {
		if strings.Contains(query, "WHERE") {
			sb.WriteString(" AND ")
		} else {
			sb.WriteString(" WHERE ")
		}

		sb.WriteString(sortField)
		sb.WriteByte(' ')

		sign := '>'
		if reverse {
			sign = '<'
		}
		sb.WriteRune(sign)

		sb.WriteString(strconv.FormatUint(offset, 10))
	}

	sb.WriteString(" ORDER BY ")
	sb.WriteString(sortField)
	sb.WriteByte(' ')

	orderDirection := "ASC"
	if reverse {
		orderDirection = "DESC"
	}
	sb.WriteString(orderDirection)

	if limit > 0 {
		sb.WriteString(" LIMIT ")
		sb.WriteString(strconv.FormatUint(limit, 10))
	}

	return sb.String()
}

// BulkInsertValues builds a query to insert multiple values in a single database call.
func BulkInsertValues(rows, values int) string {
	list := make([]string, 0, rows)
	placeholders := strings.Repeat("?,", values)
	placeholders = "(" + placeholders[:len(placeholders)-1] + ")"

	for i := 0; i < rows; i++ {
		list = append(list, placeholders)
	}

	return strings.Join(list, ",")
}

const migrations = `
CREATE TABLE IF NOT EXISTS bets (
	idx INTEGER NOT NULL CHECK (idx > 0),
	tickets INTEGER CHECK (tickets > 0),
	public_key VARCHAR(64) NOT NULL,
	lottery_height INTEGER NOT NULL,
	FOREIGN KEY (lottery_height) REFERENCES lotteries(height),
	PRIMARY KEY (idx, lottery_height)
);

CREATE TABLE IF NOT EXISTS winners (
	public_key VARCHAR(64) NOT NULL,
	prize INTEGER NOT NULL,
	ticket INTEGER NOT NULL,
	lottery_height INTEGER NOT NULL,
	FOREIGN KEY (lottery_height) REFERENCES lotteries(height)
);

CREATE INDEX IF NOT EXISTS lottery_heights ON winners(lottery_height);

CREATE TABLE IF NOT EXISTS prizes (
	amount INTEGER NOT NULL CHECK (amount >= 0),
	public_key VARCHAR(64) NOT NULL,
	lottery_height INTEGER NOT NULL,
	expired BOOLEAN DEFAULT 0 CHECK (expired IN (0, 1)),
	FOREIGN KEY (public_key) REFERENCES winners(public_key),
	FOREIGN KEY (lottery_height) REFERENCES lotteries(height)
);

CREATE TABLE IF NOT EXISTS notifications (
	public_key VARCHAR(64) PRIMARY KEY,
	chat_id INTEGER NOT NULL,
	service TEXT NOT NULL CHECK (service IN ('telegram'))
) WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS lotteries (
	height INTEGER PRIMARY KEY CHECK (height > 0)
);

CREATE TABLE IF NOT EXISTS lightning (
	public_key VARCHAR(64) NOT NULL,
	address VARCHAR(255) NOT NULL,
	PRIMARY KEY (public_key, address)
);`
