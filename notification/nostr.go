package notification

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/logger"
	"github.com/aftermath2/BTRY/nostr"

	"github.com/pkg/errors"
)

type nostrc struct {
	client *nostr.Client
}

// newNostrNotifier returns a notifier that sends nostr events.
func newNostrNotifier(config config.Nostr, logger *logger.Logger, torClient *http.Client) *nostrc {
	return &nostrc{
		client: nostr.NewClient(config, logger, torClient),
	}
}

func (n *nostrc) PublishWinners(blockHeight uint32, winners []db.Winner) error {
	message := buildMessage(blockHeight, winners)

	if err := n.client.Publish(message); err != nil {
		return errors.Wrap(err, "publishing event")
	}

	return nil
}

func buildMessage(blockHeight uint32, winners []db.Winner) string {
	var msg strings.Builder
	msg.WriteString("Lottery winners. Block: ")
	msg.WriteString(strconv.FormatUint(uint64(blockHeight), 10))
	msg.WriteString("\n------------------------------\n")

	for i, winner := range winners {
		if i != 0 {
			msg.WriteByte('\n')
		}

		msg.WriteString(strconv.FormatInt(int64(i+1), 10))
		msg.WriteString(". Ticket #")
		msg.WriteString(strconv.FormatUint(winner.Ticket, 10))
		msg.WriteString(" from ")
		msg.WriteString(winner.PublicKey)
		msg.WriteString(" won ")
		msg.WriteString(strconv.FormatUint(winner.Prize, 10))
		msg.WriteString(" sats")
	}

	return msg.String()
}
