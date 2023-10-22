package notification

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/mock"
)

// TelegramBotAPIMock is a mocked implementation of a telegram.
type TelegramBotAPIMock struct {
	mock.Mock
}

// NewTelegramBotAPIMock returns a mocked telegram service.
func NewTelegramBotAPIMock() *TelegramBotAPIMock {
	return &TelegramBotAPIMock{}
}

// GetUpdatesChan mock.
func (t *TelegramBotAPIMock) GetUpdatesChan(config tg.UpdateConfig) tg.UpdatesChannel {
	args := t.Called(config)
	return args.Get(0).(chan tg.Update)
}

// Send mock.
func (t *TelegramBotAPIMock) Send(c tg.Chattable) (tg.Message, error) {
	args := t.Called(c)
	return args.Get(0).(tg.Message), args.Error(1)
}
