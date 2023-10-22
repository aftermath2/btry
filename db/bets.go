package db

import (
	"database/sql"

	"github.com/aftermath2/BTRY/logger"

	"github.com/pkg/errors"
)

// BetsStore contains the methods used to store and retrieve bets from the database.
type BetsStore interface {
	Add(bet Bet) error
	GetPrizePool() (uint64, error)
	List(offset, limit uint64, reverse bool) ([]Bet, error)
	Reset() error
}

// Bet represents a user bet.
type Bet struct {
	PublicKey string `json:"public_key,omitempty" db:"public_key"`
	Index     uint64 `json:"index,omitempty"`
	Tickets   uint64 `json:"tickets,omitempty"`
}

type bets struct {
	db     *sql.DB
	logger *logger.Logger
}

// newBetsStore returns a new bets storage service.
func newBetsStore(db *sql.DB, logger *logger.Logger) BetsStore {
	return &bets{
		db:     db,
		logger: logger,
	}
}

// Add saves a bet in the database.
func (b *bets) Add(bet Bet) error {
	tx, err := b.db.Begin()
	if err != nil {
		return errors.Wrap(err, "starting transaction")
	}
	defer tx.Rollback()

	highestIndex, err := getHighestIndex(tx)
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO bets (idx, tickets, public_key) VALUES (?,?,?)")
	if err != nil {
		return errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	index := highestIndex + bet.Tickets
	if _, err := stmt.Exec(index, bet.Tickets, bet.PublicKey); err != nil {
		return errors.Wrap(err, "adding bet")
	}

	return tx.Commit()
}

// GetPrizePool returns the prize pool size.
func (b *bets) GetPrizePool() (uint64, error) {
	tx, err := b.db.Begin()
	if err != nil {
		return 0, errors.Wrap(err, "starting transaction")
	}
	defer tx.Rollback()

	highestIndex, err := getHighestIndex(tx)
	if err != nil {
		return 0, err
	}

	return highestIndex, nil
}

// List returns a list of bets.
//
// A limit value of 0 means there's no limit.
func (b *bets) List(offset, limit uint64, reverse bool) ([]Bet, error) {
	// Cap limit to avoid creating a slice with too big capacity
	if limit > 100 {
		limit = 100
	}

	query := "SELECT idx, tickets, public_key FROM bets"
	clauses := AddPagination(offset, limit, "idx", reverse)
	query += clauses

	stmt, err := b.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, errors.Wrap(err, "listing bets")
	}
	defer rows.Close()

	bets := make([]Bet, 0, limit)
	// Reuse object
	var bet Bet
	for rows.Next() {
		if err := rows.Scan(&bet.Index, &bet.Tickets, &bet.PublicKey); err != nil {
			return nil, err
		}

		bets = append(bets, bet)
	}

	return bets, nil
}

// Reset removes all bets from the database.
func (b *bets) Reset() error {
	stmt, err := b.db.Prepare("DELETE FROM bets")
	if err != nil {
		return errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	if _, err := stmt.Exec(); err != nil {
		return errors.Wrap(err, "deleting all bets")
	}

	return nil
}

func getHighestIndex(tx *sql.Tx) (uint64, error) {
	stmt, err := tx.Prepare("SELECT MAX(idx) FROM bets")
	if err != nil {
		return 0, errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	var index *uint64
	row := stmt.QueryRow()
	if err := row.Scan(&index); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, errors.Wrap(err, "scanning highest index")
	}

	if index == nil {
		return 0, nil
	}

	return *index, nil
}
