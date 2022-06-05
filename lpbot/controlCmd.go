package lpbot

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/khase/leaseplan-bot/lpbot/config"
	"github.com/khase/leaseplan-bot/lpbot/tgcon"
)

var (
	StartCmd = &tgcon.MessageCommand{
		CommandTrigger:   "start",
		ShortDescription: "",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handleStartCommand(message, UserMap)
		},
	}
	ResumeCmd = &tgcon.MessageCommand{
		CommandTrigger:   "resume",
		ShortDescription: "",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handleResumeCommand(message, UserMap.Users[message.From.ID])
		},
	}
	PauseCmd = &tgcon.MessageCommand{
		CommandTrigger:   "pause",
		ShortDescription: "",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handlePauseCommand(message, UserMap.Users[message.From.ID])
		},
	}
	WhoamiCmd = &tgcon.MessageCommand{
		CommandTrigger:   "whoami",
		ShortDescription: "",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handleWhoamiCommand(message, UserMap.Users[message.From.ID])
		},
	}
)

func handleResumeCommand(message *tgbotapi.Message, user *config.User) ([]tgbotapi.Chattable, error) {
	if user == nil {
		return nil, tgcon.ErrCommandPermittedForUnknownUser
	}

	user.StartWatcher()
	// handler := NewUserHandler(user)
	// handler.StartWatcher(bot)
	user.Save()

	return nil, nil
}

func handlePauseCommand(message *tgbotapi.Message, user *config.User) ([]tgbotapi.Chattable, error) {
	if user == nil {
		return nil, tgcon.ErrCommandPermittedForUnknownUser
	}

	user.StopWatcher()
	user.Save()

	return nil, nil
}

func handleStartCommand(message *tgbotapi.Message, userMap *config.UserMap) ([]tgbotapi.Chattable, error) {
	user, exists := userMap.Users[message.From.ID]
	if exists {
		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			fmt.Sprintf(
				"Hallo %s,\nwir kennen uns bereits 🤗.",
				user.FriendlyName))

		return []tgbotapi.Chattable{msg}, nil
	}

	user, err := userMap.CreateNewUser(message.From.ID, message.From.FirstName)
	if err != nil {
		return nil, err
	}

	msg := tgbotapi.NewMessage(
		message.Chat.ID,
		fmt.Sprintf(
			"Hallo %s,\nich kenne dich jetzt und wir können beginnen 🎉🎊\nTeile mir am besten deinen Leaseplan Token (/setToken, /login) oder connecte dich mit einem deiner Kollegen (/connect).",
			user.FriendlyName))

	return []tgbotapi.Chattable{msg}, nil
}

func handleWhoamiCommand(message *tgbotapi.Message, user *config.User) ([]tgbotapi.Chattable, error) {
	if user == nil {
		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			"Das weiß ich leider auch nicht 🤷‍♂️")
		msg.ReplyToMessageID = message.MessageID

		return []tgbotapi.Chattable{msg}, nil
	}

	infos, err := user.GetHumanReadableUserInfo()
	if err != nil {
		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			fmt.Sprintf("Hallo %s,\nich kenne deinen Namen aber irgendwas stimmt mit meinen Schaltkreisen nicht und ich kann dir mehr leider auch nicht sagen 😨", user.FriendlyName))
		msg.ReplyToMessageID = message.MessageID

		return []tgbotapi.Chattable{msg}, nil
	}

	msg := tgbotapi.NewMessage(
		message.Chat.ID,
		fmt.Sprintf("Hallo %s 🙂,\nfolgende Infos habe ich über dich:\n%s\n\nAm besten löschst du die Nachricht wieder, damit dein Token hier nicht im Verlauf stehen bleibt 😉.", user.FriendlyName, infos))
	msg.ReplyToMessageID = message.MessageID

	return []tgbotapi.Chattable{msg}, nil
}
