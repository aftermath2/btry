package notification

import (
	"fmt"
	"testing"

	"github.com/aftermath2/BTRY/db"

	"github.com/stretchr/testify/assert"
)

func TestBuildMessage(t *testing.T) {
	blockHeight := uint32(300000)
	winners := []db.Winner{
		{
			PublicKey: "One",
			Prize:     1,
			Ticket:    1,
		},
		{
			PublicKey: "Two",
			Prize:     2,
			Ticket:    2,
		},
		{
			PublicKey: "Three",
			Prize:     3,
			Ticket:    3,
		},
	}
	expectedMessage := fmt.Sprintf(`Lottery winners. Block: %d
------------------------------
1. Ticket #%d from %s won %d sats
2. Ticket #%d from %s won %d sats
3. Ticket #%d from %s won %d sats`,
		blockHeight,
		winners[0].Ticket, winners[0].PublicKey, winners[0].Prize,
		winners[1].Ticket, winners[1].PublicKey, winners[1].Prize,
		winners[2].Ticket, winners[2].PublicKey, winners[2].Prize,
	)
	message := buildMessage(blockHeight, winners)

	assert.Equal(t, expectedMessage, message)
}
