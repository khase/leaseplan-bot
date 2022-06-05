package lpbot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/khase/leaseplan-bot/lpbot/config"
	"github.com/khase/leaseplan-bot/lpbot/tgcon"
)

var (
	FilterCmd = &tgcon.MessageCommand{
		CommandTrigger:   "filter",
		ShortDescription: "",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handleFilterCommand(message, UserMap.Users[message.From.ID])
		},
	}
)

func handleFilterCommand(message *tgbotapi.Message, user *config.User) ([]tgbotapi.Chattable, error) {
	return nil, tgcon.ErrCommandNotImplemented
}
