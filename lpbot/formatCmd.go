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
		ShortDescription: "setzt deine persönliche summaryMessage",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handleSummaryMessageFormatCommand(message, UserMap.Users[message.From.ID])
		},
	}
	DetailFormatCmd = &tgcon.MessageCommand{
		CommandTrigger:   "setdetailmessageformat",
		ShortDescription: "setzt deine persönliche detailMessage",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handleDetailMessageFormatCommand(message, UserMap.Users[message.From.ID])
		},
	}
	TestFormatCmd = &tgcon.MessageCommand{
		CommandTrigger:   "test",
		ShortDescription: "gibt die aktuellen Daten als Testnachricht zurück",
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

	oldTemplate := user.SummaryMessageTemplate
	user.SummaryMessageTemplate = message.CommandArguments()

	_, err := user.LastFrame.GetTestMessages(user, 1)
	if err != nil {
		fmt.Printf("Could not evaluate template \"%s\": %s", user.SummaryMessageTemplate, err)
		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			fmt.Sprintf("Dein Format \"%s\" kann leider nicht übernommen werden: %s", user.SummaryMessageTemplate, err))
		msg.ReplyToMessageID = message.MessageID

		user.SummaryMessageTemplate = oldTemplate
		return []tgbotapi.Chattable{msg}, nil
	}
	user.Save()

	msg := tgbotapi.NewMessage(
		message.Chat.ID,
		fmt.Sprintf("Ich habe dein Format \"%s\" übernommen", user.SummaryMessageTemplate))
	msg.ReplyToMessageID = message.MessageID

	return []tgbotapi.Chattable{msg}, nil
}

func handleDetailMessageFormatCommand(message *tgbotapi.Message, user *config.User) ([]tgbotapi.Chattable, error) {
	if user == nil {
		return nil, tgcon.ErrCommandPermittedForUnknownUser
	}

	oldTemplate := user.DetailMessageTemplate
	user.DetailMessageTemplate = message.CommandArguments()

	_, err := user.LastFrame.GetTestMessages(user, 1)
	if err != nil {
		fmt.Printf("Could not evaluate template \"%s\": %s", user.DetailMessageTemplate, err)
		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			fmt.Sprintf("Dein Format \"%s\" kann leider nicht übernommen werden: %s", user.DetailMessageTemplate, err))
		msg.ReplyToMessageID = message.MessageID

		user.DetailMessageTemplate = oldTemplate
		return []tgbotapi.Chattable{msg}, nil
	}
	user.Save()

	msg := tgbotapi.NewMessage(
		message.Chat.ID,
		fmt.Sprintf("Ich habe dein Format \"%s\" übernommen", user.DetailMessageTemplate))
	msg.ReplyToMessageID = message.MessageID

	return []tgbotapi.Chattable{msg}, nil
}
