package lpbot

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/khase/leaseplan-bot/lpbot/config"
	"github.com/khase/leaseplan-bot/lpbot/tgcon"
	"github.com/khase/leaseplanabocarexporter/pkg"
)

var (
	LoginCmd = &tgcon.MessageCommand{
		CommandTrigger:   "login",
		ShortDescription: "loggt dich bei leaseplan ein email/password",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handleLoginCommand(message, UserMap.Users[message.From.ID])
		},
	}
	TokenCmd = &tgcon.MessageCommand{
		CommandTrigger:   "settoken",
		ShortDescription: "loggt dich bei leaseplan ein token",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handleSetTokenCommand(message, UserMap.Users[message.From.ID])
		},
	}
	ConnectCmd = &tgcon.MessageCommand{
		CommandTrigger:   "connect",
		ShortDescription: "verwende den lp-Account eines Kollegen",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handleConnectCommand(message, UserMap.Users[message.From.ID])
		},
	}
)

func handleConnectCommand(message *tgbotapi.Message, user *config.User) ([]tgbotapi.Chattable, error) {
	return nil, tgcon.ErrCommandNotImplemented
}

func handleLoginCommand(message *tgbotapi.Message, user *config.User) ([]tgbotapi.Chattable, error) {
	if user == nil {
		return nil, tgcon.ErrCommandPermittedForUnknownUser
	}

	command := strings.Split(message.Text, " ")

	if len(command) != 3 {
		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			"Bitte sende mir dein Leaseplan login in folgendem Format \"/login <dein username> <dein passwort>\" (du kannst mir alternativ auch gleich ein leaseplan token zusenden \"/setToken\").")
		msg.ReplyToMessageID = message.MessageID

		return []tgbotapi.Chattable{msg}, nil
	}

	token, err := pkg.GetToken(command[1], command[2])
	if err != nil {
		return nil, err
	}

	return setToken(token, message, user)
}

func handleSetTokenCommand(message *tgbotapi.Message, user *config.User) ([]tgbotapi.Chattable, error) {
	if user == nil {
		return nil, tgcon.ErrCommandPermittedForUnknownUser
	}

	command := strings.Split(message.Text, " ")

	if len(command) != 2 {
		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			"Bitte sende mir dein Leaseplan token in folgendem Format \"/setToken <dein Token>\".")
		msg.ReplyToMessageID = message.MessageID

		return []tgbotapi.Chattable{msg}, nil
	}

	return setToken(command[1], message, user)
}

func setToken(token string, message *tgbotapi.Message, user *config.User) ([]tgbotapi.Chattable, error) {
	user.LeaseplanToken = token
	user.Save()

	deleteCredsMsg := tgbotapi.NewDeleteMessage(
		message.Chat.ID,
		message.MessageID)
	msg := tgbotapi.NewMessage(
		message.Chat.ID,
		"Perfekt 🎉, das hat schonmal geklappt 😊.\nSicherheitshalber habe ich das Token aus unserem Verlauf gelöscht.\n\nWenn du benachrichtigt werden willst sobald sich dein angebot bei Leaseplan geändert hat, aktiviere deine Updates mit /resume.\nDu kannst natürlich noch das Nachrichtenformat (/messageFormat) sowie eigene Filter (/filter) einstellen.")

	return []tgbotapi.Chattable{deleteCredsMsg, msg}, nil
}
