// Package notification contains utilities for notifying the winners of the lottery.
package notification

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/crypto"
	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/logger"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

// TODO:
// - If the user did not withdraw the funds within one day of the message being sent,
// send a reminder every one day until it expires

var escapedCharacters = map[rune]struct{}{
	'_': {}, '.': {}, '!': {}, '<': {}, '>': {}, '*': {}, '+': {}, '-': {}, '=': {},
	'(': {}, ')': {}, '[': {}, ']': {}, '{': {}, '}': {}, '~': {}, '#': {}, '|': {},
}

type botAPI interface {
	GetUpdatesChan(config tg.UpdateConfig) tg.UpdatesChannel
	Send(c tg.Chattable) (tg.Message, error)
}

type telegram struct {
	logger  *logger.Logger
	botAPI  botAPI
	db      *db.DB
	botName string
}

func newTelegramNotifier(
	config config.Telegram,
	db *db.DB,
	logger *logger.Logger,
	torClient *http.Client,
) (*telegram, error) {
	botAPI, err := tg.NewBotAPIWithClient(config.BotAPIToken, tg.APIEndpoint, torClient)
	if err != nil {
		return nil, errors.Wrap(err, "creating telegram bot API")
	}

	return &telegram{
		logger:  logger,
		botAPI:  botAPI,
		botName: config.BotName,
		db:      db,
	}, nil
}

func (t *telegram) GetUpdates() {
	updatesChannel := t.botAPI.GetUpdatesChan(tg.UpdateConfig{Timeout: 10})

	for {
		select {
		case update := <-updatesChannel:
			t.processUpdate(update)
		}
	}
}

func (t *telegram) processUpdate(update tg.Update) {
	chatID := update.Message.From.ID
	// Message should have the format `/start <public_key>`
	split := strings.Split(update.Message.Text, " ")
	if len(split) != 2 {
		t.Notify(chatID, errInvalidMessage)
		return
	}

	publicKey := split[1]
	if err := crypto.ValidatePublicKey(publicKey); err != nil {
		t.Notify(chatID, fmt.Sprintf(errInvalidPublicKey, publicKey))
		return
	}

	if _, err := t.db.Notifications.GetChatID(publicKey); err == nil {
		t.Notify(chatID, errAlreadyEnabled)
		return
	}

	if err := t.db.Notifications.Add(publicKey, chatID); err != nil {
		t.Notify(chatID, errInternalError)
		t.logger.Error(errors.Wrap(err, "adding telegram entry"))
		return
	}

	t.Notify(chatID, fmt.Sprintf(welcome, update.Message.From.UserName))
}

func (t *telegram) Notify(chatID int64, message string) {
	msg := tg.NewMessage(chatID, formatMessage(message))
	msg.ParseMode = tg.ModeMarkdownV2
	msg.ChannelUsername = t.botName

	if _, err := t.botAPI.Send(msg); err != nil {
		t.logger.Error(errors.Wrapf(err, "sending message to chat %d", chatID))
	}
}

func formatMessage(message string) string {
	newMessage := make([]rune, 0, len(message))
	for _, c := range message {
		if _, ok := escapedCharacters[c]; ok {
			newMessage = append(newMessage, '\\')
		}
		newMessage = append(newMessage, c)
	}

	return string(newMessage)
}
