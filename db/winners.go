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
	Add(lotteryHeight uint32, winners []Winner) error
	ClaimPrizes(publicKey string, amount uint64) error
	ExpirePrizes(lotteryHeight uint32) (uint64, error)
	GetPrizes(publicKey string) (uint64, error)
	List(lotteryHeight uint32) ([]Winner, error)
}

// Winner represents a user that had a winning ticket.
type Winner struct {
	PublicKey string `json:"public_key,omitempty" db:"public_key"`
	Prizes    uint64 `json:"prizes,omitempty"`
	Ticket    uint64 `json:"ticket,omitempty"`
	Expired   bool   `json:"expired,omitempty"`
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
	query := "INSERT INTO winners (public_key, prizes, ticket, lottery_height) VALUES "
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
		args = append(args, winner.Prizes)
		args = append(args, winner.Ticket)
		args = append(args, lotteryHeight)
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

	stmt, err := w.db.Prepare("UPDATE winners SET prizes=prizes-? WHERE public_key=? AND expired=0")
	if err != nil {
		return errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	if _, err := stmt.Exec(amount, publicKey); err != nil {
		return errors.Wrap(err, "updating winners")
	}

	return nil
}

// ExpirePrizes looks for prizes that were given more than 3 lotteries ago.
//
// It returns the amount of funds expired.
func (w *winners) ExpirePrizes(lotteryHeight uint32) (uint64, error) {
	stmt, err := w.db.Prepare("UPDATE winners SET expired=1 WHERE lottery_height < ? RETURNING prizes")
	if err != nil {
		return 0, errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	rows, err := stmt.Query(lotteryHeight)
	if err != nil {
		return 0, errors.Wrap(err, "expiring prizes")
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
	stmt, err := w.db.Prepare("SELECT prizes FROM winners WHERE public_key=? AND expired=0")
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

// List returns the winners from the lottery at the lottery height specified.
func (w *winners) List(lotteryHeight uint32) ([]Winner, error) {
	query := "SELECT public_key, prizes, ticket, expired FROM winners WHERE lottery_height=?"
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
		if err := rows.Scan(
			&winner.PublicKey, &winner.Prizes, &winner.Ticket, &winner.Expired,
		); err != nil {
			return nil, errors.Wrap(err, "scanning rows")
		}

		winners = append(winners, winner)
	}

	return winners, nil
}
