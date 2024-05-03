package nostr

import (
	"context"
	"net/http"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/logger"

	"github.com/nbd-wtf/go-nostr"
	"github.com/pkg/errors"
	"nhooyr.io/websocket"
)

// Client represents a nostr client.
type Client struct {
	logger     *logger.Logger
	torClient  *http.Client
	privateKey string
	relays     []string
}

// NewClient creates and returns a client to publish messages to nostr.
func NewClient(config config.Nostr, logger *logger.Logger, torClient *http.Client) *Client {
	return &Client{
		logger:     logger,
		privateKey: config.PrivateKey,
		relays:     config.Relays,
		torClient:  torClient,
	}
}

// Publish publishes an event to the configured relays.
func (c *Client) Publish(message string) error {
	event, err := c.createEvent(message)
	if err != nil {
		return errors.Wrap(err, "creating event")
	}

	eventEnvelope := nostr.EventEnvelope{Event: event}
	body, err := eventEnvelope.MarshalJSON()
	if err != nil {
		return errors.Wrap(err, "encoding event")
	}

	dialOpts := &websocket.DialOptions{
		HTTPClient: c.torClient,
		Host:       "",
	}

	for _, relay := range c.relays {
		if err := send(relay, dialOpts, body); err != nil {
			// Errors are not returned so we attempt to to publish the message to other relays
			c.logger.Error(err)
		}
	}

	return nil
}

func (c *Client) createEvent(message string) (nostr.Event, error) {
	event := nostr.Event{
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindTextNote,
		Tags:      nil,
		Content:   message,
	}

	if err := event.Sign(c.privateKey); err != nil {
		return nostr.Event{}, errors.Wrap(err, "signing event")
	}

	return event, nil
}

// send opens a websocket connection with the relay and sends the event.
func send(relay string, dialOpts *websocket.DialOptions, body []byte) error {
	ctx := context.Background()
	conn, resp, err := websocket.Dial(ctx, relay, dialOpts)
	if err != nil {
		return errors.Wrap(err, "opening connection")
	}

	if resp.StatusCode != http.StatusSwitchingProtocols {
		return errors.Errorf("invalid response (%d) from %q", resp.StatusCode, relay)
	}

	if err := conn.Write(ctx, websocket.MessageText, body); err != nil {
		return errors.Wrap(err, "sending event")
	}

	if err := conn.Close(websocket.StatusNormalClosure, ""); err != nil {
		return errors.Wrap(err, "closing connection")
	}

	return nil
}
