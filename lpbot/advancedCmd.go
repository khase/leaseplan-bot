package lpbot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/khase/leaseplan-bot/lpbot/config"
	"github.com/khase/leaseplan-bot/lpbot/tgcon"
)

var (
	FilterCmd = &tgcon.MessageCommand{
		CommandTrigger:   "filter",
		ShortDescription: "setzt einen Filter für Benachrichtigungen",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handleFilterCommand(message, UserMap.Users[message.From.ID])
		},
	}
	ExcelCmd = &tgcon.MessageCommand{
		CommandTrigger:   "excel",
		ShortDescription: "erstellt eine excel liste aller verfügbaren Autos",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handleExcelCommand(message, UserMap.Users[message.From.ID])
		},
	}
)

func handleFilterCommand(message *tgbotapi.Message, user *config.User) ([]tgbotapi.Chattable, error) {
	if user == nil {
		return nil, tgcon.ErrCommandPermittedForUnknownUser
	}

	args := strings.Split(message.CommandArguments(), " ")

	switch args[0] {
	case "list":
		filterList, err := user.GetHumanReadableFilterList()
		if err != nil {
			msg := tgbotapi.NewMessage(
				message.Chat.ID,
				fmt.Sprintf("Hallo %s,\nirgendwas stimmt mit meinen Schaltkreisen nicht und ich kann dir mehr leider auch nicht sagen 😨", user.FriendlyName))
			msg.ReplyToMessageID = message.MessageID

			return []tgbotapi.Chattable{msg}, nil
		}

		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			fmt.Sprintf("Hallo %s 🙂,\nDu hast folgende Filter aktiv:\n%s", user.FriendlyName, filterList))
		msg.ReplyToMessageID = message.MessageID

		return []tgbotapi.Chattable{msg}, nil

	case "add":
		if len(args) < 2 {
			msg := tgbotapi.NewMessage(
				message.Chat.ID,
				fmt.Sprintf("Ich konnte keinen filter zum hinzufügen finden"))
			msg.ReplyToMessageID = message.MessageID

			return []tgbotapi.Chattable{msg}, nil
		}
		filter := strings.Join(args[1:], " ")
		user.AddFilter(filter)
		user.Save()

		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			fmt.Sprintf("Ich habe '%s' für dich als filter hinzugefügt 👍", filter))
		msg.ReplyToMessageID = message.MessageID

		return []tgbotapi.Chattable{msg}, nil

	case "remove":
		if len(args) < 2 {
			msg := tgbotapi.NewMessage(
				message.Chat.ID,
				fmt.Sprintf("Ich konnte keinen filter zum entfernen finden"))
			msg.ReplyToMessageID = message.MessageID

			return []tgbotapi.Chattable{msg}, nil
		}
		filter := strings.Join(args[1:], " ")
		user.RemoveFilter(filter)
		user.Save()

		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			fmt.Sprintf("Ich habe '%s' aus deinen filtern entfernt 👍", filter))
		msg.ReplyToMessageID = message.MessageID

		return []tgbotapi.Chattable{msg}, nil
	}

	msg := tgbotapi.NewMessage(
		message.Chat.ID,
		fmt.Sprintf("Ich kann mit '%s' leider nichts anfangen 😨", args[0]))
	msg.ReplyToMessageID = message.MessageID

	return []tgbotapi.Chattable{msg}, nil
}

func handleExcelCommand(message *tgbotapi.Message, user *config.User) ([]tgbotapi.Chattable, error) {
	return nil, tgcon.ErrCommandNotImplemented

	// msg := tgbotapi.NewDocument(
	// 	message.Chat.ID,
	// 	tgbotapi.FileBytes())
	// msg.ReplyToMessageID = message.MessageID

	// return []tgbotapi.Chattable{msg}, nil
}
