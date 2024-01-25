package notification

import (
	"context"
	"strconv"
	"strings"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/db"

	"github.com/nbd-wtf/go-nostr"
	"github.com/pkg/errors"
)

type nostrc struct {
	privateKey string
	relays     []string
}

// newNostrNotifier returns a notifier that sends nostr events.
func newNostrNotifier(config config.Nostr) *nostrc {
	return &nostrc{
		privateKey: config.PrivateKey,
	}
}

func (n *nostrc) PublishWinners(winners []db.Winner) error {
	pub, err := nostr.GetPublicKey(n.privateKey)
	if err != nil {
		return errors.Wrap(err, "obtaining public key")
	}

	var message strings.Builder
	message.WriteString("BTRY lottery winners\n---------------\n")
	for i, winner := range winners {
		if i != 0 {
			message.WriteByte('\n')
		}

		message.WriteString("Ticket ")
		message.WriteString(strconv.FormatUint(winner.Ticket, 10))
		message.WriteString(" from ")
		message.WriteString(winner.PublicKey)
		message.WriteString(" won ")
		message.WriteString(strconv.FormatUint(winner.Prize, 10))
		message.WriteString(" sats")
	}

	event := nostr.Event{
		PubKey:    pub,
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindTextNote,
		Tags:      nil,
		Content:   message.String(),
	}

	if err := event.Sign(n.privateKey); err != nil {
		return err
	}

	ctx := context.Background()
	for _, url := range n.relays {
		// TODO: add the posibility of specifying a custom proxy
		// https://github.com/nbd-wtf/go-nostr/issues/75
		relay, err := nostr.RelayConnect(ctx, url)
		if err != nil {
			return err
		}

		if err := relay.Publish(ctx, event); err != nil {
			return err
		}
	}

	return nil
}
