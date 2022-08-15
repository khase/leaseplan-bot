package tgcon

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	totalMessagesRecieved = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tgcon_total_messages_recieved",
			Help: "The total number of messages recieved",
		},
		[]string{
			"username",
		})
	totalCommandMessagesRecieved = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tgcon_total_command_messages_recieved",
			Help: "The total number of command messages recieved",
		},
		[]string{
			"username",
			"command",
		})
	totalNonCommandMessagesRecieved = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tgcon_total_noncommand_messages_recieved",
			Help: "The total number of non command messages recieved",
		},
		[]string{
			"username",
		})
	totalErrorsOccured = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tgcon_total_errors_occured",
			Help: "The total number of errors occured",
		},
		[]string{
			"username",
		})
	totalPanicsCatched = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tgcon_total_panics_catched",
			Help: "The total number of panics catched",
		},
		[]string{
			"username",
		})

	ErrTelegramTokenUnser = errors.New("no telegram token has been provided")

	ErrUnknown                        = errors.New("internal error")
	ErrCommandNotImplemented          = errors.New("command not implemended")
	ErrCommandPermittedForUnknownUser = errors.New("this command can not be executed by unknown users")
	ErrCommandPermitted               = errors.New("this command can not be executed by this users")
)

type TgConnector struct {
	token string
	debug bool

	telegram *tgbotapi.BotAPI

	receiverRunning bool

	commands []*MessageCommand
}

func NewTgConnector(token string, debug bool) *TgConnector {
	tgCon := new(TgConnector)
	tgCon.token = token
	tgCon.debug = debug
	tgCon.receiverRunning = false
	tgCon.commands = []*MessageCommand{}

	return tgCon
}

func (bot *TgConnector) AddCommand(cmd *MessageCommand) {
	bot.commands = append(bot.commands, cmd)
}

func (bot *TgConnector) GetTgBotApi() *tgbotapi.BotAPI {
	return bot.telegram
}

func (bot *TgConnector) Init() error {
	if bot.token == "" {
		return ErrTelegramTokenUnser
	}

	telegram, err := tgbotapi.NewBotAPI(bot.token)
	if err != nil {
		return err
	}

	telegram.Debug = bot.debug
	log.Printf("Telegram: Authorized on account %s", telegram.Self.UserName)
	bot.telegram = telegram

	return nil
}

func (bot *TgConnector) ReceiveMessages() {
	log.Printf("Telegram receiver (%s): started.", bot.telegram.Self.UserName)
	bot.receiverRunning = true
	defer func() {
		bot.receiverRunning = false
		log.Printf("Telegram receiver (%s): shutdown", bot.telegram.Self.UserName)
	}()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.telegram.GetUpdatesChan(u)

	for bot.receiverRunning {
		for update := range updates {
			if !bot.receiverRunning {
				return
			}
			if update.Message != nil { // If we got a message
				if err := bot.handleMessage(update.Message); err != nil {
					totalNonCommandMessagesRecieved.WithLabelValues(update.Message.From.UserName).Inc()

					fmt.Printf("Error occured for user \"%s\": %s", update.Message.From.UserName, err)
				}
			}
		}
	}
}

func (bot *TgConnector) Shutdown() {
	bot.receiverRunning = false
}

func (bot *TgConnector) handleMessage(message *tgbotapi.Message) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("catched Panic: %+v\n", r)
			totalPanicsCatched.WithLabelValues(message.From.FirstName).Inc()

			msg := tgbotapi.NewMessage(
				message.Chat.ID,
				fmt.Sprintf("Ouch, da ist aber etwas richtig schief gelaufen 🤯\nWende dich bitte an den Bot-Admin."))
			msg.ReplyToMessageID = message.MessageID

			bot.telegram.Send(msg)
		}
	}()

	totalMessagesRecieved.WithLabelValues(message.From.FirstName).Inc()

	if !message.IsCommand() {
		totalNonCommandMessagesRecieved.WithLabelValues(message.From.FirstName).Inc()
		log.Printf("Got non command Message from %s: %s", message.From.FirstName, message.Text)

		msg := tgbotapi.NewMessage(message.Chat.ID, strconv.FormatInt(int64(math.Pow(2, 16))+rand.Int63n(int64(math.Pow(2, 32))), 2))
		msg.ReplyToMessageID = message.MessageID

		bot.telegram.Send(msg)
		return nil
	} else {
		err := bot.handleCommand(message)

		if errors.Is(err, ErrCommandNotImplemented) {
			msg := tgbotapi.NewMessage(
				message.Chat.ID,
				"Tut mir leid, aber das kann ich leider noch nicht 😣")
			msg.ReplyToMessageID = message.MessageID
			bot.telegram.Send(msg)

			return nil
		} else if errors.Is(err, ErrCommandPermittedForUnknownUser) {
			msg := tgbotapi.NewMessage(
				message.Chat.ID,
				fmt.Sprintf("Hallo %s,\ndu hast noch gar kein Profil bei mir. Du kannst jeder Zeit mit dem Kommando /start ein Profil bei mir erstellen 😉", message.From.FirstName))
			msg.ReplyToMessageID = message.MessageID

			bot.telegram.Send(msg)
			return nil
		} else if err != nil {
			msg := tgbotapi.NewMessage(
				message.Chat.ID,
				fmt.Sprintf("Ouch, da ist irgendetwas schief gelaufen 😵"))
			msg.ReplyToMessageID = message.MessageID

			bot.telegram.Send(msg)
			return err
		}

		return nil
	}
}

func (bot *TgConnector) GetCommandDescriptions() string {
	buf := new(bytes.Buffer)
	for _, cmd := range bot.commands {
		buf.WriteString(fmt.Sprintf("%s - %s\n", strings.ToLower(cmd.CommandTrigger), cmd.ShortDescription))
	}

	return buf.String()
}

func (bot *TgConnector) handleCommand(message *tgbotapi.Message) error {
	log.Printf("Handle command Message from %s: %s", message.From.FirstName, message.Text)
	for _, cmd := range bot.commands {
		if strings.ToLower(cmd.CommandTrigger) == strings.ToLower(message.Command()) {
			totalCommandMessagesRecieved.WithLabelValues(message.From.FirstName, cmd.CommandTrigger).Inc()
			// log.Printf("Executing command: %s", cmd.CommandTrigger)
			resultMessages, err := cmd.Execute(message)
			// data, err := yaml.Marshal(resultMessages)
			// log.Printf("%s resolved in %d messages for %s:\n%s", cmd.CommandTrigger, len(resultMessages), message.From.FirstName, data)
			for _, resultMessage := range resultMessages {
				if resultMessage == nil {
					continue
				}
				bot.telegram.Send(resultMessage)
			}
			return err
		}
	}

	return ErrCommandNotImplemented
}
