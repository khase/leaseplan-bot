package lpbot

import (
	"log"

	"github.com/khase/leaseplan-bot/lpbot/config"
	"github.com/khase/leaseplan-bot/lpbot/lpcon"
	"github.com/khase/leaseplanabocarexporter/dto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	userLeaseplanCarsVisible = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lpcon_cars_visible",
			Help: "Number of cars visible to the user",
		},
		[]string{
			"username",
		})
	totalMessagesSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tgcon_total_messages_sent",
			Help: "The total number of messages sent",
		},
		[]string{
			"username",
		})
)

type UserHandler struct {
	user *config.User

	WatcherRunning bool
}

func NewUserHandler(user *config.User) *UserHandler {
	handler := new(UserHandler)
	handler.user = user

	return handler
}

func (handler *UserHandler) StartWatcher(bot *tgbotapi.BotAPI) {
	if !handler.WatcherRunning {
		go handler.watch(bot)
	}
}

func (handler *UserHandler) watch(bot *tgbotapi.BotAPI) {
	log.Printf("Handler for %s(%d): started.", handler.user.FriendlyName, handler.user.UserId)
	handler.WatcherRunning = true
	watcher := lpcon.NewLpWatcher(handler.user)
	defer func() {
		handler.WatcherRunning = false
		log.Printf("Handler for %s(%d): shutdown", handler.user.FriendlyName, handler.user.UserId)
	}()

	updateChannel := make(chan []dto.Item)
	go watcher.Watch(updateChannel)

	for update := range updateChannel {
		userLeaseplanCarsVisible.WithLabelValues(handler.user.FriendlyName).Set(float64(len(update)))
		log.Printf("Handler for %s(%d): got %d car items", handler.user.FriendlyName, handler.user.UserId, len(update))
		frame := config.NewDataFrame(handler.user.LastFrame.Current, update)
		log.Printf("Handler for %s(%d): found differences: +%d, -%d", handler.user.FriendlyName, handler.user.UserId, len(frame.Added), len(frame.Removed))

		if frame.HasChanges {
			messages, err := frame.GetMessages(handler.user)
			if err != nil {
				log.Printf("Handler for %s(%d): got an error: %s", handler.user.FriendlyName, handler.user.UserId, err)
			}

			totalMessagesSent.WithLabelValues(handler.user.FriendlyName).Add(float64(len(messages)))
			for _, message := range messages {
				bot.Send(message)
			}
		}

		handler.user.LastFrame = frame
		handler.user.SaveUserCache()
	}

	log.Printf("Handler for %s(%d): has been disabled", handler.user.FriendlyName, handler.user.UserId)
}
