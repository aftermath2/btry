// Package lottery is in charge of handling the logic for the daily lotteries.
package lottery

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"slices"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/lightning"
	"github.com/aftermath2/BTRY/logger"
	"github.com/aftermath2/BTRY/notification"

	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
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
	PrizePool  int64  `json:"prize_pool"`
	Capacity   int64  `json:"capacity"`
	NextHeight uint32 `json:"next_height"`
}

// Lottery is in charge of handling the lottery's logic.
type Lottery struct {
	lnd            lightning.Client
	notifier       notification.Notifier
	logger         *logger.Logger
	db             *db.DB
	winnersCh      chan<- []db.Winner
	blocksCh       <-chan *chainrpc.BlockEpoch
	blocksDuration uint32
}

// New returns a new Lottery object.
func New(
	config config.Lottery,
	db *db.DB,
	lnd lightning.Client,
	notifier notification.Notifier,
	winnersCh chan<- []db.Winner,
	blocksCh <-chan *chainrpc.BlockEpoch,
) (*Lottery, error) {
	logger, err := logger.New(config.Logger)
	if err != nil {
		return nil, err
	}

	return &Lottery{
		blocksDuration: config.Duration,
		logger:         logger,
		db:             db,
		lnd:            lnd,
		notifier:       notifier,
		winnersCh:      winnersCh,
		blocksCh:       blocksCh,
	}, nil
}

// Start executes the loop in charge of doing the periodic lottery.
func (l *Lottery) Start() error {
	ctx := context.Background()

	info, err := l.lnd.GetInfo(ctx)
	if err != nil {
		return err
	}

	nextHeight, err := l.db.Lotteries.GetNextHeight()
	if err != nil {
		return err
	}

	if nextHeight == 0 || info.BlockHeight > nextHeight {
		if nextHeight != 0 && info.BlockHeight > nextHeight {
			// Remove next height to avoid showing one where no lottery has taken place.
			// Means the server was down when the block was mined.
			if err := l.db.Lotteries.DeleteHeight(nextHeight); err != nil {
				return err
			}
		}

		nextHeight = info.BlockHeight + l.blocksDuration
		if err := l.db.Lotteries.AddHeight(nextHeight); err != nil {
			return err
		}
	}

	l.logger.Infof("Next block height target: %d", nextHeight)

	go func() {
		for {
			block := <-l.blocksCh
			if block.Height != nextHeight {
				continue
			}

			// Block hash bytes are reversed, correct it
			slices.Reverse(block.Hash)

			if err := l.raffle(block); err != nil {
				l.logger.Error(err)
			}

			// Add next lottery height
			nextHeight += l.blocksDuration
			if err := l.db.Lotteries.AddHeight(nextHeight); err != nil {
				l.logger.Error(err)
			}

			l.logger.Infof("Next block height target: %d", nextHeight)
		}
	}()

	return nil
}

func (l *Lottery) raffle(block *chainrpc.BlockEpoch) error {
	// Expire prizes assigned blocksDuration*5 or more blocks ago
	expiredPrizes, err := l.db.Prizes.Expire(block.Height - (l.blocksDuration * 5))
	if err != nil {
		return err
	}
	l.logger.Infof("Expired prizes: %d", expiredPrizes)

	bets, err := l.db.Bets.List(block.Height, 0, 0, false)
	if err != nil {
		return errors.Wrap(err, "listing bets")
	}

	if len(bets) == 0 {
		return nil
	}

	prizePool, err := l.db.Bets.GetPrizePool(block.Height)
	if err != nil {
		return err
	}

	winners, err := getWinners(block.Hash, prizePool, bets)
	if err != nil {
		return errors.Wrap(err, "getting winners")
	}

	if err := l.db.Winners.Add(block.Height, winners); err != nil {
		return errors.Wrap(err, "saving winners")
	}

	if err := l.db.Prizes.Set(block.Height, winners); err != nil {
		return errors.Wrap(err, "saving prizes")
	}

	l.winnersCh <- winners

	winnersMap := aggregateWinners(winners)
	l.notifyWinners(winnersMap)
	l.tryAutoWithdrawals(block.Height, winnersMap)

	return nil
}

func (l *Lottery) notify(publicKey, message string) {
	chatID, err := l.db.Notifications.GetChatID(publicKey)
	if err != nil {
		if !errors.Is(err, db.ErrNoChatID) {
			l.logger.Error(errors.Wrap(err, "getting telegram chat ID"))
		}
		return
	}

	l.notifier.Notify(chatID, message)
}

// notifyWinners sends a notification with a congratulations message to the winners if they have
// enabled the notifications.
func (l *Lottery) notifyWinners(winnersMap map[string]uint64) {
	for publicKey, prizes := range winnersMap {
		message := fmt.Sprintf(notification.Congratulations, prizes, l.blocksDuration*5)
		l.notify(publicKey, message)
	}
}

// tryAutoWithdrawals attempts to send winners their prizes via lightning addresses.
func (l *Lottery) tryAutoWithdrawals(lotteryHeight uint32, winnersMap map[string]uint64) {
	ctx := context.Background()

	for publicKey, prizes := range winnersMap {
		address, err := l.db.Lightning.GetAddress(publicKey)
		if err != nil {
			if !errors.Is(err, db.ErrNoAddress) {
				l.logger.Error(err)
			}
			continue
		}

		if err := l.db.Prizes.Withdraw(publicKey, prizes); err != nil {
			l.logger.Error(err)
			continue
		}

		preimage, err := l.lnd.SendToLightningAddress(ctx, address, int64(prizes))
		if err != nil {
			l.logger.Error(errors.Wrap(err, "sending to lightning address"))

			winner := db.Winner{
				PublicKey: publicKey,
				Prize:     prizes,
			}
			if err := l.db.Prizes.Set(lotteryHeight, []db.Winner{winner}); err != nil {
				l.logger.Error(err)
			}
			continue
		}

		message := fmt.Sprintf(notification.AutomaticWithdrawal, prizes, address, preimage)
		l.notify(publicKey, message)
	}
}

// GetInfo returns information about the lottery.
func GetInfo(ctx context.Context, lnd lightning.Client, db *db.DB) (Info, error) {
	remoteBalance, err := lnd.RemoteBalance(ctx)
	if err != nil {
		return Info{}, err
	}

	nextHeight, err := db.Lotteries.GetNextHeight()
	if err != nil {
		return Info{}, err
	}

	prizePool, err := db.Bets.GetPrizePool(nextHeight)
	if err != nil {
		return Info{}, err
	}

	return Info{
		PrizePool:  int64(prizePool),
		Capacity:   remoteBalance / CapacityDivisor,
		NextHeight: nextHeight,
	}, nil
}

// aggregateWinners returns a map with the winners and their prizes aggregated.
func aggregateWinners(winners []db.Winner) map[string]uint64 {
	winnersMap := make(map[string]uint64)

	for _, winner := range winners {
		prizes, ok := winnersMap[winner.PublicKey]
		if ok {
			winnersMap[winner.PublicKey] = prizes + winner.Prize
		} else {
			winnersMap[winner.PublicKey] = winner.Prize
		}
	}

	return winnersMap
}

// getWinners looks for the target or the closest higher number using the binary search algorithm.
//
// The bets slice must be sorted.
func getWinners(blockHash []byte, prizePool uint64, bets []db.Bet) ([]db.Winner, error) {
	if len(bets) <= 0 {
		return nil, nil
	}

	winners := make([]db.Winner, 0, len(prizes))
	i := len(blockHash) - 1

	for _, prize := range prizes {
		winningTicket := getWinningTicket(blockHash, i, prizePool)
		p := (prize / 100) * float64(prizePool)

		winner := db.Winner{
			PublicKey: getPublicKey(bets, winningTicket),
			Ticket:    winningTicket,
			Prize:     uint64(math.Round(p)),
		}

		winners = append(winners, winner)
		i -= 2
	}

	return winners, nil
}

// getWinningTicket takes two bytes from the latest block hash to get the winning number.
func getWinningTicket(hash []byte, i int, prizePool uint64) uint64 {
	num1 := int64(hash[i])
	num2 := int64(hash[i-1])

	// (num1 ^ num2) % prizePool
	result := new(big.Int).Exp(
		big.NewInt(num1),
		big.NewInt(num2),
		big.NewInt(int64(prizePool)),
	)

	// Add one so the index zero is not taken into account and the last one is
	return result.Uint64() + 1
}

func getPublicKey(bets []db.Bet, winningTicket uint64) string {
	left, mid, right := 0, 0, len(bets)-1
	for left <= right {
		mid = (left + right) / 2

		i := bets[mid].Index
		if i == winningTicket {
			return bets[mid].PublicKey
		}
		if i < winningTicket {
			left = mid + 1
			continue
		}

		right = mid - 1
	}

	// The left ends up being the higher value of the two, hence that user has the winning ticket
	return bets[left].PublicKey
}
