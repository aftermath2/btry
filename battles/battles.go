// Package battles is in charge of handling the battles logic.
package battles

import (
	"context"
	"crypto/rand"
	"math"
	"math/big"
	mrand "math/rand"
	"time"

	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/lightning"
	"github.com/aftermath2/BTRY/logger"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/pkg/errors"
)

const maxNumber = 1000

// Result represents the state of the battle after it has finished.
type Result struct {
	Players []db.Player
	Number  uint64
}

// Battle ..
type Battle struct {
	HODLInvoice string
	BattleID    uint64
}

type entry struct {
	player db.Player
	battle db.Battle
}

// Battles ..
type Battles struct {
	trackedInvoices cmap.ConcurrentMap[string, entry]
	db              *db.DB
	lnd             lightning.Client
	logger          *logger.Logger
}

// New ..
func New(db *db.DB, lnd lightning.Client, logger *logger.Logger) *Battles {
	return &Battles{
		db:     db,
		lnd:    lnd,
		logger: logger,
	}
}

// Create creates and stores a battle.
func (b *Battles) Create(ctx context.Context, battle db.Battle, player db.Player) (Battle, error) {
	invoice, err := b.lnd.DecodeInvoice(ctx, player.Invoice)
	if err != nil {
		return Battle{}, err
	}

	if battle.Amount != uint64(invoice.NumSatoshis) {
		return Battle{}, errors.Errorf("invalid invoice amount, it should be %d", battle.Amount)
	}

	hodlInvoice, err := b.lnd.AddHODLInvoice(ctx, invoice)
	if err != nil {
		return Battle{}, err
	}

	// We use random IDs instead of a serial counter to avoid leaking the number of battles that
	// were created.
	battle.ID = mrand.Uint64()
	if err := b.db.Battles.Create(battle, player); err != nil {
		return Battle{}, err
	}

	// TODO: listen for payment hash to see if it's accepted

	return Battle{
		BattleID:    battle.ID,
		HODLInvoice: hodlInvoice,
	}, nil
}

// Expire marks as expired the items that are older than a specified date and cancels all pending
// invoices as well as deleting the players.
func (b *Battles) Expire(ctx context.Context) error {
	t := time.Now().UTC().Unix() - lightning.HODLInvoiceExpiry
	invoices, err := b.db.Battles.Expire(t)
	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		inv, err := b.lnd.DecodeInvoice(ctx, invoice)
		if err != nil {
			return err
		}

		if err := b.lnd.CancelInvoice(ctx, inv.PaymentHash); err != nil {
			b.logger.Errorf("cancelling invoice %s: %v", inv.PaymentHash, err)
			continue
		}
	}

	return nil
}

// Play ..
func (b *Battles) Play(ctx context.Context, battle db.Battle, player db.Player) error {
	// The creator's invoice has been locked by the challenger, add him to the battle
	if player.Role == db.BattleCreator {
		if err := b.db.Battles.AddPlayer(battle.ID, player); err != nil {
			return err
		}
	}

	// The challenger's invoice has been locked by the creator, both invoices are locked and
	// the battle is ready to be started
	if player.Role == db.BattleChallenger {
		numberBig, err := rand.Int(rand.Reader, big.NewInt(maxNumber+1))
		if err != nil {
			return errors.Wrap(err, "failed generating random number")
		}

		players, err := b.db.Battles.GetPlayers(battle.ID)
		if err != nil {
			return err
		}

		var (
			winner     *int
			closestNum *float64
		)
		number := numberBig.Uint64()

		for i, player := range players {
			diff := math.Abs(float64(number - player.Number))

			if closestNum == nil {
				closestNum = &diff
				winner = &i
				continue
			}

			if diff == *closestNum {
				winner = nil
			}

			if diff < *closestNum {
				n := float64(player.Number)
				closestNum = &n
				winner = &i
			}
		}

		for j, player := range players {
			if winner != nil && j == *winner {
				if err := b.lnd.SettleInvoice(ctx, players[*winner].Invoice); err != nil {
					return err
				}

				continue
			}

			inv, err := b.lnd.DecodeInvoice(ctx, player.Invoice)
			if err != nil {
				return err
			}

			if err := b.lnd.CancelInvoice(ctx, inv.PaymentHash); err != nil {
				return err
			}
		}

		if err := b.db.Battles.Update(battle.ID, db.BattleFinished, number); err != nil {
			return err
		}

		// TODO: send result to the users via SSE
		// result := Result{
		// 	Players: players,
		// 	Number:  number,
		// }

	}

	return nil
}

// RequestJoin ..
func (b *Battles) RequestJoin(ctx context.Context, battleID uint64, player db.Player) (string, error) {
	invoice, err := b.lnd.DecodeInvoice(ctx, player.Invoice)
	if err != nil {
		return "", err
	}

	battle, err := b.db.Battles.GetByID(battleID)
	if err != nil {
		return "", err
	}

	if battle.Status != db.BattleCreated {
		return "", errors.New("battle can't be joined")
	}

	if battle.Amount != uint64(invoice.NumSatoshis) {
		return "", errors.Errorf("invalid invoice amount, it should be %d", battle.Amount)
	}

	hodlInvoice, err := b.lnd.AddHODLInvoice(ctx, invoice)
	if err != nil {
		return "", err
	}

	b.trackInvoice(invoice.PaymentHash, battle, player)

	return hodlInvoice, nil
}

func (b *Battles) trackInvoice(rHash string, battle db.Battle, player db.Player) {
	entry := entry{
		battle: battle,
		player: player,
	}
	b.trackedInvoices.Set(rHash, entry)
}
