package cmd

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/khase/leaseplan-bot/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	token string
	debug bool

	userDataFile string
	createNew    bool

	startCmd = &cobra.Command{
		Use:   "start",
		Short: "start the leaseplan bot",
		Long:  `start the leaseplan bot`,
		Run: func(cmd *cobra.Command, args []string) {
			err := startBot(token, userDataFile, createNew, debug)
			if err != nil {
				log.Fatal("Bot loop reportet an fatal error: ", err)
			}
		},
	}

	ErrUnknown                        = errors.New("internal error")
	ErrCommandNotImplemented          = errors.New("command not implemended")
	ErrCommandPermittedForUnknownUser = errors.New("this command can not be executed by unknown users")
	ErrCommandPermitted               = errors.New("this command can not be executed by this users")
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "token to be used for telegram auth")
	rootCmd.PersistentFlags().StringVarP(&userDataFile, "userDataFile", "u", "./leaseplan-bot.userdata", "path to file containing all user data")
	rootCmd.PersistentFlags().BoolVar(&createNew, "new", false, "if the userDataFile does not exist the bot will create a new database")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "weather or not the bot should be started in debug mode")
	viper.BindPFlag("telegramApiToken", rootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("userDataFile", rootCmd.PersistentFlags().Lookup("userDataFile"))
	viper.BindPFlag("new", rootCmd.PersistentFlags().Lookup("new"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

func startBot(apiToken string, userDataFile string, createNew bool, debug bool) error {
	if apiToken == "" {
		return errors.New("No bot token set. Use flag `-t` to provide a telegram bot api token")
	}

	userMap, err := config.LoadUserMap(userDataFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if !createNew {
			log.Printf("User userDateFile \"%s\" does not exist. To create a new one start the bot with the \"--new\" flag", userDataFile)
			return err
		}

		userMap.SaveToFile(userDataFile)
	}
	userMap.StartActiveWatchers()

	bot, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		return err
	}

	bot.Debug = debug
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	err = handleMessages(updates, bot, userMap)
	if err != nil {
		return err
	}

	return nil
}

func handleMessages(updates tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI, userMap *config.UserMap) error {
	for update := range updates {
		if update.Message != nil { // If we got a message
			handleMessage(update.Message, bot, userMap)
		}
	}

	return nil
}

func handleMessage(message *tgbotapi.Message, bot *tgbotapi.BotAPI, userMap *config.UserMap) error {
	// userId := message.From.ID
	// user, exists := userMap.Users[userId]

	if !strings.HasPrefix(message.Text, "/") {
		log.Printf("Got non command Message from %s: %s", message.From.FirstName, message.Text)

		msg := tgbotapi.NewMessage(message.Chat.ID, strconv.FormatInt(int64(math.Pow(2, 16))+rand.Int63n(int64(math.Pow(2, 32))), 2))
		msg.ReplyToMessageID = message.MessageID

		bot.Send(msg)
		return nil
	} else {
		user := userMap.Users[message.From.ID]
		err := handleCommand(message, bot, user, userMap)
		if errors.Is(err, ErrCommandNotImplemented) {
			msg := tgbotapi.NewMessage(
				message.Chat.ID,
				"Tut mir leid, aber das kann ich leider noch nicht 😣")
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)

			return nil
		} else if errors.Is(err, ErrCommandPermittedForUnknownUser) {
			msg := tgbotapi.NewMessage(
				message.Chat.ID,
				fmt.Sprintf("Hallo %s,\ndu hast noch gar kein Profil bei mir. Du kannst jeder Zeit mit dem Kommando /start ein Profil bei mir erstellen 😉", message.From.FirstName))
			msg.ReplyToMessageID = message.MessageID

			bot.Send(msg)
			return nil
		} else if err != nil {
			msg := tgbotapi.NewMessage(
				message.Chat.ID,
				fmt.Sprintf("Ouch, da ist irgendetwas schief gelaufen 😵"))
			msg.ReplyToMessageID = message.MessageID

			bot.Send(msg)
			return err
		}

		return nil
	}
}

func handleCommand(message *tgbotapi.Message, bot *tgbotapi.BotAPI, user *config.User, userMap *config.UserMap) error {
	command := strings.Split(message.Text, " ")
	switch command[0] {
	case "/start":
		return handleStartCommand(message, bot, userMap)
	case "/setToken":
		return handleSetTokenCommand(message, bot, user)
	case "/login":
		return handleLoginCommand(message, bot, user)
	case "/connect":
		return handleConnectCommand(message, bot, user)
	case "/messageFormat":
		return handleMessageFormatCommand(message, bot, user)
	case "/filter":
		return handleFilterCommand(message, bot, user)
	case "/pause":
		return handlePauseCommand(message, bot, user)
	case "/resume":
		return handleResumeCommand(message, bot, user)
	case "/whoami":
		return handleWhoamiCommand(message, bot, user)
	default:
		return nil
	}
}

func handleWhoamiCommand(message *tgbotapi.Message, bot *tgbotapi.BotAPI, user *config.User) error {
	if user == nil {
		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			"Das weiß ich leider auch nicht 🤷‍♂️")
		msg.ReplyToMessageID = message.MessageID

		bot.Send(msg)

		return nil
	}

	infos, err := user.GetHumanReadableUserInfo()
	if err != nil {
		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			fmt.Sprintf("Hallo %s,\nich kenne deinen Namen aber irgendwas stimmt mit meinen Schaltkreisen nicht und ich kann dir mehr leider auch nicht sagen 😨", user.FriendlyName))
		msg.ReplyToMessageID = message.MessageID

		bot.Send(msg)
		return nil
	}

	msg := tgbotapi.NewMessage(
		message.Chat.ID,
		fmt.Sprintf("Hallo %s 🙂,\nfolgende Infos habe ich über dich:\n%s\n\nAm besten löschst du die Nachricht wieder, damit dein Token hier nicht im Verlauf stehen bleibt 😉.", user.FriendlyName, infos))
	msg.ReplyToMessageID = message.MessageID

	bot.Send(msg)
	return nil
}

func handleResumeCommand(message *tgbotapi.Message, bot *tgbotapi.BotAPI, user *config.User) error {
	if user == nil {
		return ErrCommandPermittedForUnknownUser
	}

	user.StartWatcher()
	user.Save()

	return nil
}

func handlePauseCommand(message *tgbotapi.Message, bot *tgbotapi.BotAPI, user *config.User) error {
	if user == nil {
		return ErrCommandPermittedForUnknownUser
	}

	user.StopWatcher()
	user.Save()

	return nil
}

func handleFilterCommand(message *tgbotapi.Message, bot *tgbotapi.BotAPI, user *config.User) error {
	return ErrCommandNotImplemented
}

func handleMessageFormatCommand(message *tgbotapi.Message, bot *tgbotapi.BotAPI, user *config.User) error {
	return ErrCommandNotImplemented
}

func handleConnectCommand(message *tgbotapi.Message, bot *tgbotapi.BotAPI, user *config.User) error {
	return ErrCommandNotImplemented
}

func handleLoginCommand(message *tgbotapi.Message, bot *tgbotapi.BotAPI, user *config.User) error {
	if user == nil {
		return ErrCommandPermittedForUnknownUser
	}

	command := strings.Split(message.Text, " ")

	if len(command) != 3 {
		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			"Bitte sende mir dein Leaseplan login in folgendem Format \"/login <dein username> <dein passwort>\" (du kannst mir alternativ auch gleich ein leaseplan token zusenden \"/setToken\").")
		msg.ReplyToMessageID = message.MessageID

		bot.Send(msg)
		return nil
	}

	return setToken(command[1], message, bot, user)
}

func handleSetTokenCommand(message *tgbotapi.Message, bot *tgbotapi.BotAPI, user *config.User) error {
	if user == nil {
		return ErrCommandPermittedForUnknownUser
	}

	command := strings.Split(message.Text, " ")

	if len(command) != 2 {
		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			"Bitte sende mir dein Leaseplan token in folgendem Format \"/setToken <dein Token>\".")
		msg.ReplyToMessageID = message.MessageID

		bot.Send(msg)
		return nil
	}

	return setToken(command[1], message, bot, user)
}

func setToken(token string, message *tgbotapi.Message, bot *tgbotapi.BotAPI, user *config.User) error {
	user.LeaseplanToken = token
	user.Save()

	bot.Send(
		tgbotapi.NewDeleteMessage(
			message.Chat.ID,
			message.MessageID))
	msg := tgbotapi.NewMessage(
		message.Chat.ID,
		"Perfekt 🎉, ich wede dich benachrichtigen sobald sich dein angebot bei Leaseplan geändert hat.\nSicherheitshalber habe ich das Token aus unserem Verlauf gelöscht.\nDu kannst natürlich noch das Nachrichtenformat (/messageFormat) sowie eigene Filter (/filter) einstellen.")
	bot.Send(msg)

	return nil
}

func handleStartCommand(message *tgbotapi.Message, bot *tgbotapi.BotAPI, userMap *config.UserMap) error {
	user, exists := userMap.Users[message.From.ID]
	if exists {
		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			fmt.Sprintf(
				"Hallo %s,\nwir kennen uns bereits 🤗.",
				user.FriendlyName))
		bot.Send(msg)

		return nil
	}

	user, err := userMap.CreateNewUser(message.From.ID, message.From.FirstName)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(
		message.Chat.ID,
		fmt.Sprintf(
			"Hallo %s,\nich kenne dich jetzt und wir können beginnen 🎉🎊\nTeile mir am besten deinen Leaseplan Token (/setToken, /login) oder connecte dich mit einem deiner Kollegen (/connect).",
			user.FriendlyName))
	bot.Send(msg)

	return nil
}
