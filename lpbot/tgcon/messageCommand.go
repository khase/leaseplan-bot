package tgcon

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageCommand struct {
	CommandTrigger   string
	ShortDescription string
	Description      string
	Execute          func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error)
}
