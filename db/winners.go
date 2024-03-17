package db

import (
	"database/sql"

	"github.com/aftermath2/BTRY/logger"

	"github.com/pkg/errors"
)

// WinnersStore contains the methods used to store and retrieve winners from the database.
type WinnersStore interface {
	Add(lotteryHeight uint32, winners []Winner) error
	List(lotteryHeight uint32) ([]Winner, error)
}

// Winner represents a user that had a winning ticket.
type Winner struct {
	PublicKey string `json:"public_key,omitempty" db:"public_key"`
	Prize     uint64 `json:"prize,omitempty"`
	Ticket    uint64 `json:"ticket,omitempty"`
}

type winners struct {
	db     *sql.DB
	logger *logger.Logger
}

// newWinnersStore returns a new winners storage service.
func newWinnersStore(db *sql.DB, logger *logger.Logger) WinnersStore {
	return &winners{
		db:     db,
		logger: logger,
	}
}

// Add adds winners to the database.
func (w *winners) Add(lotteryHeight uint32, winners []Winner) error {
	query := "INSERT INTO winners (public_key, prize, ticket, lottery_height) VALUES "
	values := BulkInsertValues(len(winners), 4)
	query += values

	stmt, err := w.db.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	args := make([]any, 0, len(winners)*4)
	for _, winner := range winners {
		args = append(args, winner.PublicKey)
		args = append(args, winner.Prize)
		args = append(args, winner.Ticket)
		args = append(args, lotteryHeight)
	}

	if _, err := stmt.Exec(args...); err != nil {
		return errors.Wrap(err, "storing winners")
	}

	return nil
}

// List returns the winners from the lottery at the lottery height specified.
func (w *winners) List(lotteryHeight uint32) ([]Winner, error) {
	query := "SELECT public_key, prize, ticket FROM winners WHERE lottery_height=?"
	stmt, err := w.db.Prepare(query)
	if err != nil {
		return nil, errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	rows, err := stmt.Query(lotteryHeight)
	if err != nil {
		return nil, errors.Wrap(err, "selecting winners")
	}
	defer rows.Close()

	var (
		winners []Winner
		// Reuse object
		winner Winner
	)
	for rows.Next() {
		if err := rows.Scan(&winner.PublicKey, &winner.Prize, &winner.Ticket); err != nil {
			return nil, errors.Wrap(err, "scanning rows")
		}

		winners = append(winners, winner)
	}

	return winners, nil
}
