package db

import (
	"database/sql"

	"github.com/aftermath2/BTRY/logger"

	"github.com/pkg/errors"
)

// ErrNoAddress is thrown when a public key does not have any lightning address linked to it.
var ErrNoAddress = errors.New("no lightning address linked to this public key")

// LightningStore contains the methods used to store and retrieve lightning addresses from the database.
type LightningStore interface {
	GetAddress(publicKey string) (string, error)
	SetAddress(publicKey, address string) error
}

type lightning struct {
	db     *sql.DB
	logger *logger.Logger
}

// newLightningStore returns a new lightning storage service.
func newLightningStore(db *sql.DB, logger *logger.Logger) LightningStore {
	return &lightning{
		db:     db,
		logger: logger,
	}
}

// GetAddress returns the specified public key's linked lightning address.
func (l *lightning) GetAddress(publicKey string) (string, error) {
	query := "SELECT address FROM lightning WHERE public_key=?"
	stmt, err := l.db.Prepare(query)
	if err != nil {
		return "", errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	var address string
	if err := stmt.QueryRow(publicKey).Scan(&address); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNoAddress
		}
		return "", errors.Wrap(err, "getting lightning address")
	}

	return address, nil
}

// SetAddress links a public key with an lightning address.
func (l *lightning) SetAddress(publicKey, address string) error {
	query := "INSERT INTO lightning (public_key, address) VALUES (?,?) " +
		"ON CONFLICT (public_key, address) DO UPDATE SET address=excluded.address"
	stmt, err := l.db.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	if _, err := stmt.Exec(publicKey, address); err != nil {
		return errors.Wrap(err, "storing lightning address")
	}

	return nil
}
