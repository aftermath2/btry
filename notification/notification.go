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
	AutomaticWithdrawal = "%d sats were withdrawn to %s. Preimage: %s"
	Congratulations     = "Congratulations! You have won %d sats, your prizes expire at block %d."
	welcome             = "Hello @%s! I will send you a notification if you win."
	errInvalidMessage   = "Message not recognized. Enable notifications using `/start " +
		"<public_key>` or scanning the QR code on BTRY's web client."
	errInvalidPublicKey = "The public key %q is invalid."
	errInternalError    = "Something went wrong. Please try again later or contact an admin."
	errAlreadyEnabled   = "The public key already has notifications enabled"
)

// Notifier represents a service that is used to send messages to winners.
type Notifier interface {
	GetUpdates()
	Notify(chatID int64, message string)
	PublishWinners(blockHeight uint32, winners []db.Winner) error
}

type notifier struct {
	telegram *telegram
	nostr    *nostrc
	disabled bool
}

// NewNotifier returns a new notification sender.
func NewNotifier(config config.Notifier, db *db.DB, torClient *http.Client) (Notifier, error) {
	logger, err := logger.New(config.Logger)
	if err != nil {
		return nil, err
	}

	if config.Disabled {
		logger.Info("Notifier disabled")
		return &notifier{disabled: config.Disabled}, nil
	}

	telegram, err := newTelegramNotifier(config.Telegram, db, logger, torClient)
	if err != nil {
		return nil, err
	}

	return &notifier{
		disabled: config.Disabled,
		telegram: telegram,
		nostr:    newNostrNotifier(config.Nostr, logger, torClient),
	}, nil
}

func (n *notifier) GetUpdates() {
	if n.disabled {
		return
	}
	n.telegram.GetUpdates()
}

func (n *notifier) Notify(chatID int64, message string) {
	if n.disabled {
		return
	}
	n.telegram.Notify(chatID, message)
}

func (n *notifier) PublishWinners(blockHeight uint32, winners []db.Winner) error {
	if n.disabled {
		return nil
	}
	return n.nostr.PublishWinners(blockHeight, winners)
}
