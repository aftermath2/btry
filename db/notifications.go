package db

import (
	"database/sql"
	"time"

	"github.com/aftermath2/BTRY/logger"

	"github.com/pkg/errors"
)

// NotificationExpiry is the duration for which public keys and chat IDs are associated
const NotificationExpiry = time.Hour * 24 * 7

// NotificationsStore contains the methods used to store and retrieve notifications from the database.
type NotificationsStore interface {
	Add(publicKey string, chatID int64) error
	GetChatID(publicKey string) (int64, error)
	Expire() error
}

type notifications struct {
	db     *sql.DB
	logger *logger.Logger
}

// newNotificationsStore returns a new notifications storage service.
func newNotificationsStore(db *sql.DB, logger *logger.Logger) NotificationsStore {
	return &notifications{
		db:     db,
		logger: logger,
	}
}

// Add stores a public key to chat ID link in the database.
func (n *notifications) Add(publicKey string, chatID int64) error {
	query := "INSERT INTO notifications (public_key, chat_id, service) VALUES (?,?,?)"
	stmt, err := n.db.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	if _, err := stmt.Exec(publicKey, chatID, "telegram"); err != nil {
		return errors.Wrap(err, "adding notification")
	}

	return nil
}

// GetChatID looks for the chat ID corresponding to the public key.
func (n *notifications) GetChatID(publicKey string) (int64, error) {
	stmt, err := n.db.Prepare("SELECT chat_id FROM notifications WHERE public_key=?")
	if err != nil {
		return 0, errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	var chatID int64
	if err := stmt.QueryRow(publicKey).Scan(&chatID); err != nil {
		return 0, errors.Wrap(err, "scanning notification chat ID")
	}

	return chatID, nil
}

// Expire deletes expired notifications.
func (n *notifications) Expire() error {
	stmt, err := n.db.Prepare("DELETE FROM notifications WHERE created_at < ?")
	if err != nil {
		return errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	t := time.Now().Add(-NotificationExpiry).Unix()
	if _, err := stmt.Exec(t); err != nil {
		return errors.Wrap(err, "deleting expired notifications")
	}

	return nil
}
