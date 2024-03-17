package db

import (
	"database/sql"

	"github.com/aftermath2/BTRY/logger"

	"github.com/pkg/errors"
)

// ErrInsufficientPrizes is returned when the user requests more than he has.
var ErrInsufficientPrizes = errors.New("withdrawal amount is higher than assigned prizes")

// PrizesStore contains the methods used to store and retrieve prizes from the database.
type PrizesStore interface {
	Expire(lotteryHeight uint32) (uint64, error)
	Get(publicKey string) (uint64, error)
	Set(lotteryHeight uint32, winners []Winner) error
	Withdraw(publicKey string, amount uint64) error
}

// PrizesRow represent a prizes table row.
type PrizesRow struct {
	RowID  int    `db:"rowid"`
	Amount uint64 `db:"amount"`
}

type prizes struct {
	db     *sql.DB
	logger *logger.Logger
}

// newPrizesStore returns a new prizes storage service.
func newPrizesStore(db *sql.DB, logger *logger.Logger) PrizesStore {
	return &prizes{
		db:     db,
		logger: logger,
	}
}

// Expire sets prizes won before lotteryHeight as expired.
func (p *prizes) Expire(lotteryHeight uint32) (uint64, error) {
	query := "UPDATE prizes SET expired=1 WHERE lottery_height <= ? RETURNING amount"
	stmt, err := p.db.Prepare(query)
	if err != nil {
		return 0, errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	rows, err := stmt.Query(lotteryHeight)
	if err != nil {
		return 0, errors.Wrap(err, "expiring prizes")
	}
	defer rows.Close()

	expiredAmount := uint64(0)
	for rows.Next() {
		var amount uint64
		if err := rows.Scan(&amount); err != nil {
			return 0, errors.Wrap(err, "scanning rows")
		}

		expiredAmount += amount
	}

	return expiredAmount, nil
}

// Get returns the prizes corresponding to the public key specified.
func (p *prizes) Get(publicKey string) (uint64, error) {
	query := "SELECT COALESCE(SUM(amount), 0) FROM prizes WHERE public_key=? AND expired=0"
	stmt, err := p.db.Prepare(query)
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

// Set stores the prizes for each winner.
func (p *prizes) Set(lotteryHeight uint32, winners []Winner) error {
	query := "INSERT INTO prizes (public_key, amount, lottery_height) VALUES "
	values := BulkInsertValues(len(winners), 3)
	query += values

	stmt, err := p.db.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	args := make([]any, 0, len(winners)*3)
	for _, winner := range winners {
		args = append(args, winner.PublicKey)
		args = append(args, winner.Prize)
		args = append(args, lotteryHeight)
	}

	if _, err := stmt.Exec(args...); err != nil {
		return errors.Wrap(err, "storing prizes")
	}

	return nil
}

// Withdraw substracts the withdrawal amount from the winner prizes.
func (p *prizes) Withdraw(publicKey string, amount uint64) error {
	if amount == 0 {
		return ErrInsufficientPrizes
	}

	tx, err := p.db.Begin()
	if err != nil {
		return errors.Wrap(err, "starting transaction")
	}
	defer tx.Rollback()

	query := "SELECT rowid, amount FROM prizes WHERE public_key=? AND expired=0 AND amount != 0 ORDER BY rowid DESC"
	selectStmt, err := tx.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "preparing statement")
	}
	defer selectStmt.Close()

	rows, err := selectStmt.Query(publicKey)
	if err != nil {
		return errors.Wrap(err, "updating prizes")
	}

	var prizes []*PrizesRow
	for rows.Next() {
		var prize PrizesRow
		if err := rows.Scan(&prize.RowID, &prize.Amount); err != nil {
			return errors.Wrap(err, "scanning rows")
		}

		prizes = append(prizes, &prize)
	}

	if err := UpdatePrizes(amount, prizes); err != nil {
		return err
	}

	for _, prize := range prizes {
		updateStmt, err := tx.Prepare("UPDATE prizes SET amount=? WHERE rowid=?")
		if err != nil {
			return errors.Wrap(err, "preparing statement")
		}
		defer updateStmt.Close()

		if _, err := updateStmt.Exec(prize.Amount, prize.RowID); err != nil {
			return errors.Wrap(err, "updating prizes")
		}
	}

	return tx.Commit()
}

// UpdatePrizes substracts the amount from the prizes.
func UpdatePrizes(amount uint64, prizes []*PrizesRow) error {
	for _, prize := range prizes {
		if amount <= prize.Amount {
			prize.Amount -= amount
			return nil
		}

		if amount > prize.Amount {
			amount -= prize.Amount
			prize.Amount = 0
		}
	}

	return ErrInsufficientPrizes
}
