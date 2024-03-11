package db

import (
	"database/sql"

	"github.com/aftermath2/BTRY/logger"

	"github.com/pkg/errors"
)

// LotteriesStore contains the methods used to store and retrieve lotteries from the database.
type LotteriesStore interface {
	AddHeight(height uint32) error
	DeleteHeight(height uint32) error
	GetNextHeight() (uint32, error)
	ListHeights(offset, limit uint64, reverse bool) ([]uint32, error)
}

type lotteries struct {
	db     *sql.DB
	logger *logger.Logger
}

// newLotteriesStore returns a new notifications storage service.
func newLotteriesStore(db *sql.DB, logger *logger.Logger) LotteriesStore {
	return &lotteries{
		db:     db,
		logger: logger,
	}
}

func (l *lotteries) AddHeight(height uint32) error {
	query := "INSERT OR IGNORE INTO lotteries (height) VALUES (?)"
	stmt, err := l.db.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	if _, err := stmt.Exec(height); err != nil {
		return errors.Wrap(err, "adding height")
	}

	return nil
}

func (l *lotteries) DeleteHeight(height uint32) error {
	query := "DELETE FROM lotteries WHERE height=?"
	stmt, err := l.db.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	if _, err := stmt.Exec(height); err != nil {
		return errors.Wrap(err, "deleting height")
	}

	return nil
}

func (l *lotteries) GetNextHeight() (uint32, error) {
	query := "SELECT COALESCE(MAX(height), 0) FROM lotteries"
	stmt, err := l.db.Prepare(query)
	if err != nil {
		return 0, errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	var height uint32
	if err := stmt.QueryRow(query).Scan(&height); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, errors.Wrap(err, "scanning height")
	}

	return height, nil
}

func (l *lotteries) ListHeights(offset, limit uint64, reverse bool) ([]uint32, error) {
	// Cap limit to avoid creating a slice with too big capacity
	if limit > 100 {
		limit = 100
	}

	query := "SELECT height FROM lotteries"
	clauses := AddPagination(offset, limit, "height", reverse)
	query += clauses

	stmt, err := l.db.Prepare(query)
	if err != nil {
		return nil, errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, errors.Wrap(err, "listing heights")
	}

	heights := make([]uint32, 0, limit)
	// Reuse object
	var height uint32
	for rows.Next() {
		if err := rows.Scan(&height); err != nil {
			return nil, err
		}

		heights = append(heights, height)
	}

	return heights, nil
}
