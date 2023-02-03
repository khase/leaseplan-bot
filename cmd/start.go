package cmd

import (
	"log"

	"github.com/khase/leaseplan-bot/lpbot"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	token        string
	watcherDelay int
	debug        bool

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
)

func init() {
	startCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "token to be used for telegram auth")
	startCmd.PersistentFlags().IntVarP(&watcherDelay, "watcherDelay", "w", 15, "polling delay for watchers in minutes")
	startCmd.PersistentFlags().StringVarP(&userDataFile, "userDataFile", "u", "./leaseplan-bot.userdata", "path to file containing all user data")
	startCmd.PersistentFlags().BoolVar(&createNew, "new", false, "if the userDataFile does not exist the bot will create a new database")
	startCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "weather or not the bot should be started in debug mode")
	viper.BindPFlag("telegramApiToken", startCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("watcherDelay", startCmd.PersistentFlags().Lookup("watcherDelay"))
	viper.BindPFlag("userDataFile", startCmd.PersistentFlags().Lookup("userDataFile"))
	viper.BindPFlag("new", startCmd.PersistentFlags().Lookup("new"))
	viper.BindPFlag("debug", startCmd.PersistentFlags().Lookup("debug"))
}

func startBot(apiToken string, userDataFile string, createNew bool, debug bool) error {
	return lpbot.StartBot(apiToken, debug, userDataFile, createNew, watcherDelay)
}
