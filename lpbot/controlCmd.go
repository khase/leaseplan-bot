package lpbot

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/khase/leaseplan-bot/lpbot/config"
	"github.com/khase/leaseplan-bot/lpbot/lpcon"
	"github.com/khase/leaseplan-bot/lpbot/tgcon"
)

var (
	StartCmd = &tgcon.MessageCommand{
		CommandTrigger:   "start",
		ShortDescription: "erstellt einen internen Benutzer",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handleStartCommand(message, UserMap)
		},
	}
	ResumeCmd = &tgcon.MessageCommand{
		CommandTrigger:   "resume",
		ShortDescription: "aktiviert deine update Nachrichten",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handleResumeCommand(message, UserMap.Users[message.From.ID])
		},
	}
	PauseCmd = &tgcon.MessageCommand{
		CommandTrigger:   "pause",
		ShortDescription: "pausiert deine update Nachrichten",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handlePauseCommand(message, UserMap.Users[message.From.ID])
		},
	}
	ThrottleCmd = &tgcon.MessageCommand{
		CommandTrigger:   "throttle",
		ShortDescription: "drosselt deine nachrichten",
		Description:      "Drosselt deine Nachrichten sodass du nur noch maximal ein Update alle n Minuten bekommst. (Ein Update kann dennoch mehrere Nachrichten generieren)",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handleThrottleCommand(message, UserMap.Users[message.From.ID])
		},
	}
	IgnoreDetailsCmd = &tgcon.MessageCommand{
		CommandTrigger:   "ignoreDetails",
		ShortDescription: "sendet keine details mehr",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handleIgnoreDetailsCommand(message, UserMap.Users[message.From.ID])
		},
	}
	IgnoreRemovedCmd = &tgcon.MessageCommand{
		CommandTrigger:   "ignoreRemoved",
		ShortDescription: "sendet keine details für entfernte angebote",
		Description:      "",
		Execute: func(message *tgbotapi.Message) ([]tgbotapi.Chattable, error) {
			return handleIgnoreRemovedCommand(message, UserMap.Users[message.From.ID])
		},
	}
	WhoamiCmd = &tgcon.MessageCommand{
		CommandTrigger:   "whoami",
		ShortDescription: "gibt alle über dich bekannten Infos zurück",
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
	user.Save()
	lpcon.RegisterUserWatcher(user)

	return nil, nil
}

func handlePauseCommand(message *tgbotapi.Message, user *config.User) ([]tgbotapi.Chattable, error) {
	if user == nil {
		return nil, tgcon.ErrCommandPermittedForUnknownUser
	}

	user.StopWatcher()
	user.Save()
	lpcon.UnregisterUserWatcher(user)

	return nil, nil
}

func handleThrottleCommand(message *tgbotapi.Message, user *config.User) ([]tgbotapi.Chattable, error) {
	if user == nil {
		return nil, tgcon.ErrCommandPermittedForUnknownUser
	}

	command := strings.Split(message.Text, " ")

	var text string
	if len(command) == 1 {
		if user.WatcherDelay <= 5 {
			text = fmt.Sprintf("Du bekommst deine Updates so schnell es geht 👍")
		} else {
			text = fmt.Sprintf("Deine Updates sind gedrosselt auf maximal 1 Update alle %d Minuten", user.WatcherDelay)
		}
	} else if len(command) == 2 {
		throttle, err := strconv.Atoi(command[1])
		if err != nil {
			return nil, err
		}

		if !user.IsAdmin && (throttle < 15) {
			text = fmt.Sprintf("Sorry, das geht nicht. Ich will kein Ärger mit Leaseplan 😨🤷‍♂️.")
		} else {
			user.WatcherDelay = int32(throttle)
			user.Save()
			text = fmt.Sprintf("Deine Updates sind gedrosselt auf maximal 1 Update alle %d Minuten", user.WatcherDelay)
		}
	}

	msg := tgbotapi.NewMessage(
		message.Chat.ID,
		text)
	msg.ReplyToMessageID = message.MessageID

	return []tgbotapi.Chattable{msg}, nil
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

func handleIgnoreDetailsCommand(message *tgbotapi.Message, user *config.User) ([]tgbotapi.Chattable, error) {
	if user == nil {
		return nil, tgcon.ErrCommandPermittedForUnknownUser
	}

	if len(message.CommandArguments()) > 0 {
		arg, err := strconv.ParseBool(message.CommandArguments())
		if err != nil {
			msg := tgbotapi.NewMessage(
				message.Chat.ID,
				fmt.Sprintf("Was diese \"%s\"??? 0 oder 1?", user.FriendlyName))
			msg.ReplyToMessageID = message.MessageID

			return []tgbotapi.Chattable{msg}, err
		}

		user.IgnoreDetails = arg
	} else {
		user.IgnoreDetails = true
	}

	var msgTxt string
	if user.IgnoreDetails {
		msgTxt = fmt.Sprintf("Hallo %s,\ndu bekommst keine Detailnachrichten mehr.", user.FriendlyName)
	} else {
		msgTxt = fmt.Sprintf("Hallo %s,\ndu bekommst Detailnachrichten wieder.", user.FriendlyName)
	}

	msg := tgbotapi.NewMessage(
		message.Chat.ID,
		msgTxt)
	msg.ReplyToMessageID = message.MessageID

	return []tgbotapi.Chattable{msg}, nil
}

func handleIgnoreRemovedCommand(message *tgbotapi.Message, user *config.User) ([]tgbotapi.Chattable, error) {

	if user == nil {
		return nil, tgcon.ErrCommandPermittedForUnknownUser
	}

	if len(message.CommandArguments()) > 0 {
		arg, err := strconv.ParseBool(message.CommandArguments())
		if err != nil {
			msg := tgbotapi.NewMessage(
				message.Chat.ID,
				fmt.Sprintf("Was diese \"%s\"??? 0 oder 1?", user.FriendlyName))
			msg.ReplyToMessageID = message.MessageID

			return []tgbotapi.Chattable{msg}, err
		}

		user.IgnoreRemoved = arg
	} else {
		user.IgnoreRemoved = true
	}

	var msgTxt string
	if user.IgnoreRemoved {
		msgTxt = fmt.Sprintf("Hallo %s,\ndu bekommst keine Nachrichten für entfernte Angebote mehr.", user.FriendlyName)
	} else {
		msgTxt = fmt.Sprintf("Hallo %s,\ndu bekommst wieder Nachrichten für entfernte Angebote.", user.FriendlyName)
	}

	msg := tgbotapi.NewMessage(
		message.Chat.ID,
		msgTxt)
	msg.ReplyToMessageID = message.MessageID

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
