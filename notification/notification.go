// Package notification contains objects used to send messages through various platforms and
// services.
package notification

import (
	"net/http"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/logger"
)

// Notification message formats
const (
	Congratulations     = "Congratulations! You have won %d sats, your prizes expire in 5 days."
	welcome             = "Hello @%s! I will send you a notification if you win."
	errInvalidMessage   = "Message not recognized. Enable notifications using `/start <public_key>` or scanning the QR code on BTRY's web client."
	errInvalidPublicKey = "The public key %q is invalid."
	errInternalError    = "Something went wrong. Please try again later or contact an admin."
)

// Notifier represents a service that is used to send messages to winners.
type Notifier interface {
	GetUpdates()
	Notify(chatID int64, message string)
}

type notifier struct {
	telegram Notifier
}

// NewNotifier returns a new notification sender.
func NewNotifier(config config.Notifier, db *db.DB, torClient *http.Client) (Notifier, error) {
	logger, err := logger.New(config.Logger)
	if err != nil {
		return nil, err
	}

	telegram, err := NewTelegramNotifier(config, db, logger, torClient)
	if err != nil {
		return nil, err
	}

	return &notifier{
		telegram: telegram,
	}, nil
}

func (n *notifier) GetUpdates() {
	n.telegram.GetUpdates()
}

func (n *notifier) Notify(chatID int64, message string) {
	n.telegram.Notify(chatID, message)
}
