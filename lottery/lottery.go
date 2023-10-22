// Package lottery is in charge of handling the logic for the daily lotteries.
package lottery

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/lightning"
	"github.com/aftermath2/BTRY/logger"
	"github.com/aftermath2/BTRY/notification"

	"github.com/go-co-op/gocron"
	"github.com/pkg/errors"
)

// Prize pool percentages
const (
	first   float64 = 50
	second          = first / 2
	third           = second / 2
	fourth          = third / 2
	fifth           = fourth / 2
	sixth           = fifth / 2
	seventh         = sixth / 2
	eighth          = seventh / 2
	btryFee         = eighth

	// Lottery capacity divisor
	CapacityDivisor = 5
)

var prizes = [8]float64{first, second, third, fourth, fifth, sixth, seventh, eighth}

// Info contains details about the lottery.
type Info struct {
	PrizePool int64 `json:"prize_pool"`
	Capacity  int64 `json:"capacity"`
}

// Lottery is in charge of handling the lottery's logic.
type Lottery struct {
	logger         *logger.Logger
	db             *db.DB
	lnd            lightning.Client
	notifier       notification.Notifier
	winnersChannel chan<- []db.Winner
	scheduler      *gocron.Scheduler
	time           string
}

// New returns a new Lottery object.
func New(
	config config.Lottery,
	db *db.DB,
	lnd lightning.Client,
	notifier notification.Notifier,
	winnersChannel chan<- []db.Winner,
) (*Lottery, error) {
	logger, err := logger.New(config.Logger)
	if err != nil {
		return nil, err
	}

	return &Lottery{
		logger:         logger,
		db:             db,
		lnd:            lnd,
		notifier:       notifier,
		winnersChannel: winnersChannel,
		scheduler:      gocron.NewScheduler(time.UTC),
		time:           config.Time,
	}, nil
}

// Start executes the loop in charge of doing the periodic lottery.
func (l *Lottery) Start() {
	l.scheduler.Every(1).Day().At(l.time).WaitForSchedule().Do(l.jobFunc)
	l.scheduler.StartAsync()
}

func (l *Lottery) jobFunc() {
	l.logger.Info("Lottery started")

	if err := l.raffle(); err != nil {
		l.logger.Error(err)
	}
}

func (l *Lottery) raffle() error {
	bets, err := l.db.Bets.List(0, 0, false)
	if err != nil {
		return errors.Wrap(err, "listing bets")
	}

	if len(bets) == 0 {
		return nil
	}

	prizePool, err := l.db.Bets.GetPrizePool()
	if err != nil {
		return err
	}

	if err := l.db.Bets.Reset(); err != nil {
		return errors.Wrap(err, "deleting bets")
	}

	winners, err := l.getWinners(prizePool, bets)
	if err != nil {
		return errors.Wrap(err, "getting winners")
	}

	if err := l.db.Winners.Add(winners); err != nil {
		return errors.Wrap(err, "saving winners")
	}

	if err := l.db.Winners.WriteHistory(winners); err != nil {
		return errors.Wrap(err, "saving winners history")
	}

	l.winnersChannel <- winners

	l.notifyWinners(winners)

	expiredPrizes, err := l.db.Winners.ExpirePrizes()
	if err != nil {
		return err
	}

	l.logger.Infof("Expired prizes: %d", expiredPrizes)

	if err := l.db.Notifications.Expire(); err != nil {
		return err
	}

	// TODO:
	// Trigger loop out if there's a excess of local balance to exchange for remote
	// Keep winners prizes and a portion of the balance to avoid draining channels
	// Loop outs must not affect users' withdrawals reliability
	// loopOutAmount := localBalance - winnersPrizes - 10_000_000

	return nil
}

// getWinners looks for the target or the closest higher number using the binary search algorithm.
//
// The bets slice must be sorted.
func (l *Lottery) getWinners(prizePool uint64, bets []db.Bet) ([]db.Winner, error) {
	if len(bets) <= 0 {
		return nil, nil
	}

	now := time.Now()
	winners := make([]db.Winner, 0, len(prizes))

	for i := 0; i < len(prizes); i++ {
		ticket, publicKey, err := getWinner(bets)
		if err != nil {
			return nil, err
		}

		prize := (prizes[i] / 100) * float64(prizePool)
		winner := db.Winner{
			PublicKey: publicKey,
			Ticket:    ticket,
			Prizes:    uint64(prize),
			// Set just for the clients, server-side is generated automatically by SQL
			CreatedAt: now.Unix(),
		}
		winners = append(winners, winner)
	}

	return winners, nil
}

// getWinner returns the winning ticket and the winner public key given a cryptographically
// secure random number.
func getWinner(bets []db.Bet) (uint64, string, error) {
	highestIndex := int64(bets[len(bets)-1].Index)
	targetBig, err := rand.Int(rand.Reader, big.NewInt(highestIndex))
	if err != nil {
		return 0, "", errors.Wrap(err, "failed generating random number")
	}

	// Add 1 so 0 is not a target and the last ticket is taken into account
	target := targetBig.Uint64() + 1
	left, mid, right := 0, 0, len(bets)-1
	for left <= right {
		mid = (left + right) / 2

		i := bets[mid].Index
		if i == target {
			return target, bets[mid].PublicKey, nil
		}
		if i < target {
			left = mid + 1
			continue
		}

		right = mid - 1
	}

	// The left ends up being the higher value of the two, hence the user that has the target ticket
	return target, bets[left].PublicKey, nil
}

// notifyWinners sends a notification with a congratulations message to the winners if they have
// enabled the notifications.
func (l *Lottery) notifyWinners(winners []db.Winner) {
	winnersMap := make(map[string]uint64, len(prizes))

	// Aggregate prizes to avoid sending multiple notifications to the same winner
	for _, winner := range winners {
		prizes, ok := winnersMap[winner.PublicKey]
		if ok {
			winnersMap[winner.PublicKey] = prizes + winner.Prizes
		} else {
			winnersMap[winner.PublicKey] = winner.Prizes
		}
	}

	for publicKey, prizes := range winnersMap {
		chatID, err := l.db.Notifications.GetChatID(publicKey)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				l.logger.Error(errors.Wrap(err, "getting telegram chat ID"))
			}
			continue
		}

		message := fmt.Sprintf(notification.Congratulations, prizes)
		l.notifier.Notify(chatID, message)
	}
}

// GetInfo returns information about the lottery.
func GetInfo(ctx context.Context, lnd lightning.Client, db *db.DB) (Info, error) {
	remoteBalance, err := lnd.RemoteBalance(ctx)
	if err != nil {
		return Info{}, err
	}

	prizePool, err := db.Bets.GetPrizePool()
	if err != nil {
		return Info{}, err
	}

	return Info{
		PrizePool: int64(prizePool),
		Capacity:  remoteBalance / CapacityDivisor,
	}, nil
}
