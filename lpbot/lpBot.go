package lpbot

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/khase/leaseplan-bot/lpbot/api"
	"github.com/khase/leaseplan-bot/lpbot/config"
	"github.com/khase/leaseplan-bot/lpbot/lpcon"
	"github.com/khase/leaseplan-bot/lpbot/tgcon"
)

var (
	UserMap *config.UserMap

	ErrUserDataFileNotExistant = errors.New("userdata file does not exist")
	ErrExternalInterrupt       = errors.New("interrupted from external signal")
)

func StartBot(token string, debug bool, userDataFile string, createNew bool, watcherDelay int, watcherPageSize int) error {
	go api.InitAndListen()

	userMap, err := config.LoadUserMap(userDataFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if !createNew {
			return ErrUserDataFileNotExistant
		}

		userMap.SaveToFile(userDataFile)
	}
	UserMap = userMap

	tgBot := tgcon.NewTgConnector(token, debug)
	tgBot.AddCommand(StartCmd)
	tgBot.AddCommand(WhoamiCmd)
	tgBot.AddCommand(ResumeCmd)
	tgBot.AddCommand(PauseCmd)
	tgBot.AddCommand(LoginCmd)
	tgBot.AddCommand(TokenCmd)
	tgBot.AddCommand(EulaCmd)
	tgBot.AddCommand(ConnectCmd)
	tgBot.AddCommand(ThrottleCmd)
	tgBot.AddCommand(IgnoreDetailsCmd)
	tgBot.AddCommand(IgnoreRemovedCmd)
	tgBot.AddCommand(SummaryFormatCmd)
	tgBot.AddCommand(DetailFormatCmd)
	tgBot.AddCommand(TestFormatCmd)
	tgBot.AddCommand(FilterCmd)

	log.Printf("Bot Command Descriptions:\n%s", tgBot.GetCommandDescriptions())
	err = tgBot.Init()

	if errors.Is(err, tgcon.ErrTelegramTokenUnser) {
		return errors.New("No bot token set. Use flag `-t` to provide a telegram bot api token")
	}

	commandChannel := make(chan error)

	go tgBot.ReceiveMessages()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			tgBot.Shutdown()
			commandChannel <- ErrExternalInterrupt
		}
	}()

	sendSystemNotifications(UserMap, tgBot.GetTgBotApi())
	startActiveHandlers(UserMap, tgBot.GetTgBotApi(), watcherDelay, watcherPageSize)

	for {
		for command := range commandChannel {
			if errors.Is(command, ErrExternalInterrupt) {
				fmt.Printf("\nReceived external interrupt.\n")
				fmt.Printf("Shutting down bot...\n")
				time.Sleep(5 * time.Second)
				return nil
			}
		}
	}
}

func sendSystemNotifications(userMap *config.UserMap, bot *tgbotapi.BotAPI) {
	for _, user := range userMap.Users {
		for _, notification := range config.SystemNotifications {
			if notification.Publish.After(user.LastSystemnotification) {
				if notification.UserCondition(user) {
					fmt.Printf("Sending System Notification to user %s(%d)\n", user.FriendlyName, user.UserId)
					bot.Send(
						tgbotapi.NewMessage(
							user.UserId,
							notification.Message,
						),
					)
					notification.UserAction(user)
				}
				user.LastSystemnotification = notification.Publish
				user.Save()
			}
		}
	}
}

func startActiveHandlers(userMap *config.UserMap, bot *tgbotapi.BotAPI, delay int, pageSize int) error {
	lpcon.SetTgBotForWatcher(bot)
	lpcon.SetWatcherDelay(delay)
	lpcon.SetWatcherPageSize(pageSize)

	for _, user := range userMap.Users {
		if user.WatcherActive {
			lpcon.RegisterUserWatcher(user)
		}
	}

	return nil
}
