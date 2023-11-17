package notification

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/logger"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

func TestProcessUpdate(t *testing.T) {
	publicKey := "345fe256754b1b472e58aede6c2f138ce67d05d431c776bcb4e384edbbdca9cd"
	username := "test"

	cases := []struct {
		desc           string
		publicKey      string
		message        string
		fail           bool
		internalError  bool
		alreadyEnabled bool
	}{
		{
			desc:      "Welcome",
			publicKey: publicKey,
			message:   fmt.Sprintf(welcome, username),
		},
		{
			desc:      "Invalid message",
			publicKey: "",
			message:   errInvalidMessage,
			fail:      true,
		},
		{
			desc:      "Invalid public key",
			publicKey: publicKey[:len(publicKey)/2],
			message:   fmt.Sprintf(errInvalidPublicKey, publicKey[:len(publicKey)/2]),
			fail:      true,
		},
		{
			desc:          "Internal error",
			publicKey:     publicKey,
			message:       errInternalError,
			fail:          true,
			internalError: true,
		},
		{
			desc:           "Already enabled",
			publicKey:      publicKey,
			message:        errAlreadyEnabled,
			alreadyEnabled: true,
		},
	}

	botAPI := NewTelegramBotAPIMock()
	notificationsMock := db.NewNotificationsStoreMock()
	telegram := &telegram{
		logger:  &logger.Logger{},
		botAPI:  botAPI,
		botName: "BTRY",
		db:      &db.DB{Notifications: notificationsMock},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			chatID := int64(123123)
			if tc.alreadyEnabled {
				notificationsMock.On("GetChatID", tc.publicKey).Return(chatID, nil).Once()
			} else {
				notificationsMock.On("GetChatID", tc.publicKey).
					Return(chatID, errors.New("")).Once()
			}
			if !tc.fail {
				notificationsMock.On("Add", tc.publicKey, chatID).Return(nil).Once()
			}
			if tc.internalError {
				notificationsMock.On("Add", tc.publicKey, chatID).Return(errors.New("")).Once()
			}

			tgMessage := createTelegramMessage(chatID, formatMessage(tc.message), telegram.botName)
			botAPI.On("Send", tgMessage).Return(tg.Message{}, nil).Once()

			update := tg.Update{
				Message: &tg.Message{
					From: &tg.User{
						ID:       chatID,
						UserName: username,
					},
					Text: strings.Trim("/start "+tc.publicKey, " "),
				},
			}

			telegram.processUpdate(update)
		})
	}
}

func TestNotify(t *testing.T) {
	botAPI := NewTelegramBotAPIMock()
	telegram := &telegram{
		botAPI:  botAPI,
		botName: "BTRY",
	}

	chatID := int64(123123)
	message := "Hello, world"

	tgMessage := createTelegramMessage(chatID, message, telegram.botName)
	botAPI.On("Send", tgMessage).Return(tg.Message{}, nil)

	telegram.Notify(chatID, message)

	botAPI.AssertExpectations(t)
}

func TestFormatMessage(t *testing.T) {
	cases := []struct {
		message         string
		expectedMessage string
	}{
		{
			message:         "Congratulations!",
			expectedMessage: "Congratulations\\!",
		},
		{
			message:         "Enable notifications using `/start <public_key>`",
			expectedMessage: "Enable notifications using `/start \\<public\\_key\\>`",
		},
		{
			message:         "Something (went) [wrong].",
			expectedMessage: "Something \\(went\\) \\[wrong\\]\\.",
		},
		{
			message:         "Bitcoin = ∞ / 21M",
			expectedMessage: "Bitcoin \\= ∞ / 21M",
		},
	}

	for i, tc := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotMessage := formatMessage(tc.message)

			assert.Equal(t, tc.expectedMessage, gotMessage)
		})
	}
}

func createTelegramMessage(chatID int64, message, botName string) tg.MessageConfig {
	msg := tg.NewMessage(chatID, message)
	msg.ParseMode = tg.ModeMarkdownV2
	msg.ChannelUsername = botName
	return msg
}
