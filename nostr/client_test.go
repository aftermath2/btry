package nostr

import (
	"testing"

	"github.com/aftermath2/BTRY/config"

	nostrlib "github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"
)

func TestPublish(t *testing.T) {
	privateKey := nostrlib.GeneratePrivateKey()
	client := NewClient(config.Nostr{PrivateKey: privateKey}, nil, nil)

	err := client.Publish("test")
	assert.NoError(t, err)
}

func TestCreateEvent(t *testing.T) {
	message := "test"
	privateKey := nostrlib.GeneratePrivateKey()
	publicKey, err := nostrlib.GetPublicKey(privateKey)
	assert.NoError(t, err)

	client := NewClient(config.Nostr{PrivateKey: privateKey}, nil, nil)
	expectedEvent := nostrlib.Event{
		PubKey:  publicKey,
		Kind:    nostrlib.KindTextNote,
		Tags:    make(nostrlib.Tags, 0),
		Content: message,
	}

	event, err := client.createEvent(message)
	assert.NoError(t, err)

	assert.Len(t, event.Sig, 128)
	assert.Len(t, event.ID, 64)
	assert.NotNil(t, event.CreatedAt)
	// Remove dynamic fields, not calculated to avoid mocking the time
	event.Sig = ""
	event.ID = ""
	event.CreatedAt = 0
	assert.Equal(t, expectedEvent, event)
}
