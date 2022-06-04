package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/khase/leaseplan-bot/lpbot/tgcon"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	ErrExternalInterrupt = errors.New("interrupted from external signal")
)

func init() {
	startCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "token to be used for telegram auth")
	startCmd.PersistentFlags().StringVarP(&userDataFile, "userDataFile", "u", "./leaseplan-bot.userdata", "path to file containing all user data")
	startCmd.PersistentFlags().BoolVar(&createNew, "new", false, "if the userDataFile does not exist the bot will create a new database")
	startCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "weather or not the bot should be started in debug mode")
	viper.BindPFlag("telegramApiToken", startCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("userDataFile", startCmd.PersistentFlags().Lookup("userDataFile"))
	viper.BindPFlag("new", startCmd.PersistentFlags().Lookup("new"))
	viper.BindPFlag("debug", startCmd.PersistentFlags().Lookup("debug"))
}

func startBot(apiToken string, userDataFile string, createNew bool, debug bool) error {

	tgBot := tgcon.NewTgConnector(apiToken, debug, userDataFile, createNew)
	err := tgBot.Init()

	if errors.Is(err, tgcon.ErrTelegramTokenUnser) {
		return errors.New("No bot token set. Use flag `-t` to provide a telegram bot api token")
	} else if errors.Is(err, tgcon.ErrUserDataFileNotExistant) {
		log.Printf("User userDateFile \"%s\" does not exist. To create a new one start the bot with the \"--new\" flag", userDataFile)
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
