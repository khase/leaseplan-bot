package lpbot

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/khase/leaseplan-bot/lpbot/config"
	"github.com/khase/leaseplan-bot/lpbot/tgcon"
)

var (
	SummaryFormatCmd = &tgcon.MessageCommand{
		CommandTrigger:   "setsummarymessageformat",
		ShortDescription: "",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handleSummaryMessageFormatCommand(message, UserMap.Users[message.From.ID])
		},
	}
	DetailFormatCmd = &tgcon.MessageCommand{
		CommandTrigger:   "setdetailmessageformat",
		ShortDescription: "",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handleDetailMessageFormatCommand(message, UserMap.Users[message.From.ID])
		},
	}
	TestFormatCmd = &tgcon.MessageCommand{
		CommandTrigger:   "test",
		ShortDescription: "",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handleTestCommand(message, UserMap.Users[message.From.ID])
		},
	}
)

func handleTestCommand(message *tgbotapi.Message, user *config.User) ([]tgbotapi.Chattable, error) {
	if user == nil {
		return nil, tgcon.ErrCommandPermittedForUnknownUser
	}

	command := strings.Split(message.Text, " ")

	if len(command) == 1 {
		messages, err := user.LastFrame.GetTestMessages(user, 0)
		if err != nil {
			msg := tgbotapi.NewMessage(
				message.Chat.ID,
				fmt.Sprintf("Deine Formatierung schlägt leider fehl: %s", err))
			msg.ReplyToMessageID = message.MessageID

			return []tgbotapi.Chattable{msg}, err
		}

		return messages, nil
	} else if len(command) == 2 {
		testMessages, err := strconv.Atoi(command[1])
		if err != nil {
			return nil, err
		}

		messages, err := user.LastFrame.GetTestMessages(user, testMessages)
		if err != nil {
			msg := tgbotapi.NewMessage(
				message.Chat.ID,
				fmt.Sprintf("Deine Formatierung schlägt leider fehl: %s", err))
			msg.ReplyToMessageID = message.MessageID

			return []tgbotapi.Chattable{msg}, err
		}

		return messages, nil
	}

	return nil, nil
}

func handleSummaryMessageFormatCommand(message *tgbotapi.Message, user *config.User) ([]tgbotapi.Chattable, error) {
	if user == nil {
		return nil, tgcon.ErrCommandPermittedForUnknownUser
	}

	if _, after, found := strings.Cut(message.Text, "/summarymessageformat "); !found {
		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			"Bitte sende mir dein Format wie folgt\"/summarymessageformat <dein Format>\"")
		msg.ReplyToMessageID = message.MessageID

		return []tgbotapi.Chattable{msg}, nil
	} else {
		oldTemplate := user.SummaryMessageTemplate
		user.SummaryMessageTemplate = after

		_, err := user.LastFrame.GetTestMessages(user, 1)
		if err != nil {
			msg := tgbotapi.NewMessage(
				message.Chat.ID,
				fmt.Sprintf("Dein Format \"%s\" kann leider nicht übernommen werden: %s", after, err))
			msg.ReplyToMessageID = message.MessageID

			user.SummaryMessageTemplate = oldTemplate
			return []tgbotapi.Chattable{msg}, nil
		}
		user.Save()

		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			fmt.Sprintf("Ich habe dein Format \"%s\" übernommen", after))
		msg.ReplyToMessageID = message.MessageID

		return []tgbotapi.Chattable{msg}, nil
	}
}

func handleDetailMessageFormatCommand(message *tgbotapi.Message, user *config.User) ([]tgbotapi.Chattable, error) {
	if user == nil {
		return nil, tgcon.ErrCommandPermittedForUnknownUser
	}

	if _, after, found := strings.Cut(message.Text, "/detailmessageformat "); !found {
		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			"Bitte sende mir dein Format wie folgt\"/detailmessageformat <dein Format>\"")
		msg.ReplyToMessageID = message.MessageID

		return []tgbotapi.Chattable{msg}, nil
	} else {
		oldTemplate := user.DetailMessageTemplate
		user.DetailMessageTemplate = after

		_, err := user.LastFrame.GetTestMessages(user, 1)
		if err != nil {
			msg := tgbotapi.NewMessage(
				message.Chat.ID,
				fmt.Sprintf("Dein Format \"%s\" kann leider nicht übernommen werden: %s", after, err))
			msg.ReplyToMessageID = message.MessageID

			user.DetailMessageTemplate = oldTemplate
			return []tgbotapi.Chattable{msg}, nil
		}
		user.Save()

		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			fmt.Sprintf("Ich habe dein Format \"%s\" übernommen", after))
		msg.ReplyToMessageID = message.MessageID

		return []tgbotapi.Chattable{msg}, nil
	}
}
