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
	Notifications NotificationsStore
	Battles       BattlesStore
	Bets          BetsStore
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
		Battles:       newBattlesStore(db, logger),
		Bets:          newBetsStore(db, logger),
		Notifications: newNotificationsStore(db, logger),
		Winners:       newWinnersStore(db, logger),
	}, nil
}

// Close releases all related resources.
func (db *DB) Close() error {
	return db.db.Close()
}

// AddPagination returns the pagination part of an SQL query.
func AddPagination(offset, limit uint64, sortField string, reverse bool) string {
	// Safety mechanism to avoid SQL injections (although sortField should always be hardcoded),
	// parameterized column names are not supported
	if strings.ContainsAny(sortField, ";'\"%\n\t\r\b-/*\\") {
		return ""
	}

	var sb strings.Builder

	if offset > 0 {
		sign := '>'
		if reverse {
			sign = '<'
		}
		sb.WriteString(" WHERE ")
		sb.WriteString(sortField)
		sb.WriteByte(' ')
		sb.WriteRune(sign)
		sb.WriteString(strconv.FormatUint(offset, 10))
	}

	orderDirection := "ASC"
	if reverse {
		orderDirection = "DESC"
	}
	sb.WriteString(" ORDER BY ")
	sb.WriteString(sortField)
	sb.WriteByte(' ')
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

func scanRows[T ~uint64 | ~string](rows *sql.Rows) ([]T, error) {
	var (
		list []T
		item T
	)
	for rows.Next() {
		if err := rows.Scan(&item); err != nil {
			return nil, errors.Wrap(err, "scanning rows")
		}

		list = append(list, item)
	}

	return list, nil
}

const migrations = `
CREATE TABLE IF NOT EXISTS bets (
	idx INTEGER PRIMARY KEY CHECK (idx > 0),
	tickets INTEGER CHECK (tickets > 0),
	public_key VARCHAR(64) NOT NULL
) WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS winners (
	public_key VARCHAR(64) PRIMARY KEY,
	prizes INTEGER NOT NULL CHECK (prizes >= 0),
	ticket INTEGER NOT NULL,
	created_at INTEGER DEFAULT (unixepoch())
) WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS winners_history (
	public_key VARCHAR(64),
	prizes INTEGER NOT NULL CHECK (prizes >= 0),
	ticket INTEGER NOT NULL,
	created_at INTEGER DEFAULT (unixepoch())
);

CREATE TRIGGER IF NOT EXISTS winners_history_ro_columns
BEFORE UPDATE OF public_key, prizes, ticket, created_at ON winners_history
BEGIN
    SELECT raise(abort, 'updating winners history is not permitted');
END;

CREATE TABLE IF NOT EXISTS notifications (
	public_key VARCHAR(64) PRIMARY KEY,
	chat_id INTEGER NOT NULL,
	service TEXT NOT NULL CHECK (service IN ('telegram')),
	created_at INTEGER DEFAULT (unixepoch())
) WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS battles (
	id INTEGER PRIMARY KEY NOT NULL,
	amount INTEGER NOT NULL CHECK (amount > 0),
	status INTEGER DEFAULT 1 CHECK (status IN (1,2,3,4)),
	number INTEGER,
	created_at DATETIME DEFAULT (unixepoch())
);

CREATE TABLE IF NOT EXISTS players (
	battle_id INTEGER,
	role INTEGER NOT NULL,
	public_key VARCHAR(64) NOT NULL,
	invoice TEXT NOT NULL,
	number INTEGER NOT NULL CHECK (number >= 0 AND number <= 1000),
	FOREIGN KEY (battle_id) REFERENCES battles (id) ON DELETE CASCADE
);

CREATE TRIGGER IF NOT EXISTS battles_ro_columns
BEFORE UPDATE OF id, amount, initiator_public_key, initiator_invoice, initiator_number ON battles
BEGIN
    SELECT raise(abort, 'updating battle read-only fields');
END;

CREATE TABLE IF NOT EXISTS notifications (
	public_key VARCHAR(64) PRIMARY KEY,
	chat_id INTEGER NOT NULL,
	service TEXT NOT NULL CHECK (service IN ('telegram')),
	expires_at INTEGER NOT NULL
) WITHOUT ROWID;`
