package db

import (
	"database/sql"
	"time"

	"github.com/aftermath2/BTRY/logger"

	"github.com/pkg/errors"
)

// ErrInsufficientPrizes is returned when the user requests more than he has.
var ErrInsufficientPrizes = errors.New("claimed amount is higher than assigned prizes")

// PrizesExpiry indicates the time window in which a prize can be redeemed.
const PrizesExpiry = time.Hour * 24 * 5

// WinnersStore contains the methods used to store and retrieve notifications from the database.
type WinnersStore interface {
	Add(winners []Winner) error
	ClaimPrizes(publicKey string, amount uint64) error
	ExpirePrizes() (uint64, error)
	GetPrizes(publicKey string) (uint64, error)
	List() ([]Winner, error)
	ListHistory(from, to uint64) ([]Winner, error)
	WriteHistory(winners []Winner) error
}

// Winner represents a user that had a winning ticket.
type Winner struct {
	PublicKey string `json:"public_key,omitempty" db:"public_key"`
	Prizes    uint64 `json:"prizes,omitempty"`
	Ticket    uint64 `json:"ticket,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty" db:"created_at"`
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
func (w *winners) Add(winners []Winner) error {
	// TODO: if a user wins using the same public key on multiple days and does not withdraw the
	// funds, then the latter prizes will expire earlier because created_at is not updated
	query := "INSERT INTO winners (public_key, prizes, ticket) VALUES "
	values := BulkInsertValues(len(winners), 3)
	query += values
	query += " ON CONFLICT (public_key) DO UPDATE SET prizes=prizes+excluded.prizes"

	stmt, err := w.db.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	args := make([]any, 0, len(winners)*3)
	for _, winner := range winners {
		args = append(args, winner.PublicKey)
		args = append(args, winner.Prizes)
		args = append(args, winner.Ticket)
	}

	if _, err := stmt.Exec(args...); err != nil {
		return errors.Wrap(err, "storing winners")
	}

	return nil
}

// ClaimPrizes removes the specified amount from the prizes.
func (w *winners) ClaimPrizes(publicKey string, amount uint64) error {
	prizes, err := w.GetPrizes(publicKey)
	if err != nil {
		return err
	}

	if amount > prizes {
		return ErrInsufficientPrizes
	}

	stmt, err := w.db.Prepare("UPDATE winners SET prizes=prizes-? WHERE public_key=?")
	if err != nil {
		return errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	if _, err := stmt.Exec(amount, publicKey); err != nil {
		return errors.Wrap(err, "updating winners")
	}

	return nil
}

// ExpirePrizes looks for prizes that were given more than 3 days ago. It takes a moveFunds function
// that should send the specified amount outside of the channel.
//
// It returns the amount of funds expired.
func (w *winners) ExpirePrizes() (uint64, error) {
	stmt, err := w.db.Prepare("DELETE FROM winners WHERE created_at < ? OR prizes=0 RETURNING prizes")
	if err != nil {
		return 0, errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	t := time.Now().Add(-PrizesExpiry).Unix()
	rows, err := stmt.Query(t)
	if err != nil {
		return 0, errors.Wrap(err, "deleting expired prizes")
	}
	defer rows.Close()

	expiredFunds := uint64(0)
	for rows.Next() {
		var prizes uint64
		if err := rows.Scan(&prizes); err != nil {
			return 0, errors.Wrap(err, "scanning rows")
		}

		expiredFunds += prizes
	}

	return expiredFunds, nil
}

// GetPrizes returns the number of satoshis available for the public key to withdraw.
func (w *winners) GetPrizes(publicKey string) (uint64, error) {
	stmt, err := w.db.Prepare("SELECT prizes FROM winners WHERE public_key=?")
	if err != nil {
		return 0, errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	var prizes uint64
	if err := stmt.QueryRow(publicKey).Scan(&prizes); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, errors.Wrap(err, "scanning prizes")
	}

	return prizes, nil
}

// List returns all the entries from the winners file.
func (w *winners) List() ([]Winner, error) {
	stmt, err := w.db.Prepare("SELECT public_key, prizes, ticket, created_at FROM winners")
	if err != nil {
		return nil, errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	rows, err := stmt.Query()
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
		if err := rows.Scan(
			&winner.PublicKey, &winner.Prizes, &winner.Ticket, &winner.CreatedAt,
		); err != nil {
			return nil, errors.Wrap(err, "scanning rows")
		}

		winners = append(winners, winner)
	}

	return winners, nil
}

// ListHistory returns the winners from the previous lottery.
func (w *winners) ListHistory(from, to uint64) ([]Winner, error) {
	query := "SELECT public_key, prizes, ticket, created_at FROM winners_history WHERE created_at >= ? AND created_at <= ?"
	stmt, err := w.db.Prepare(query)
	if err != nil {
		return nil, errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	rows, err := stmt.Query(from, to)
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
		if err := rows.Scan(
			&winner.PublicKey, &winner.Prizes, &winner.Ticket, &winner.CreatedAt,
		); err != nil {
			return nil, errors.Wrap(err, "scanning rows")
		}

		winners = append(winners, winner)
	}

	return winners, nil
}

// WriteHistory writes the winners list from the previous lottery to a different file that won't be
// modified until the next lottery.
func (w *winners) WriteHistory(winners []Winner) error {
	query := "INSERT INTO winners_history (public_key, prizes, ticket) VALUES "
	values := BulkInsertValues(len(winners), 3)
	query += values

	stmt, err := w.db.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "preparing insert statement")
	}
	defer stmt.Close()

	args := make([]any, 0, len(winners)*3)
	for _, winner := range winners {
		args = append(args, winner.PublicKey)
		args = append(args, winner.Prizes)
		args = append(args, winner.Ticket)
	}

	if _, err := stmt.Exec(args...); err != nil {
		return errors.Wrap(err, "storing winners history")
	}

	return nil
}
