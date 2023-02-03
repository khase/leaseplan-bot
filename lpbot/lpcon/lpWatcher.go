package lpcon

import (
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/khase/leaseplan-bot/lpbot/config"
	"github.com/khase/leaseplanabocarexporter/dto"
	"github.com/khase/leaseplanabocarexporter/pkg"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	totalRequestsStarted = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lpcon_total_requests",
			Help: "The total number of requests sent to leaseplan",
		},
		[]string{
			"username",
		})
	totalRequestErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lpcon_total_request_errors",
			Help: "The total number of requests returned an error",
		},
		[]string{
			"username",
		})
	requestTime = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lpcon_request_time",
			Help: "The duration in ms the leaseplan request took to finish",
		},
		[]string{
			"username",
		})

	watcherList        map[string]*LpWatcher = make(map[string]*LpWatcher)
	tgBot              *tgbotapi.BotAPI
	globalWatcherDelay int
)

type LpWatcher struct {
	levelKey string
	userlist map[string]*config.User

	isActive bool
}

func NewLpWatcher(levelKey string) *LpWatcher {
	watcher := new(LpWatcher)
	watcher.levelKey = levelKey
	watcher.userlist = make(map[string]*config.User)
	watcher.isActive = false

	return watcher
}

func SetTgBotForWatcher(bot *tgbotapi.BotAPI) {
	tgBot = bot
}

func SetWatcherDelay(delay int) {
	globalWatcherDelay = delay
}

func RegisterUserWatcher(user *config.User) {
	if updateUserInfo(user) != nil {
		return
	}

	watcher, exists := watcherList[user.LeaseplanLevelKey]

	if !exists {
		watcher = NewLpWatcher(user.LeaseplanLevelKey)
		watcherList[user.LeaseplanLevelKey] = watcher

		go watcher.Start()
	}

	watcher.registerUser(user)
}

func UnregisterUserWatcher(user *config.User) {
	watcher, exists := watcherList[user.LeaseplanLevelKey]

	if !exists {
		return
	}

	watcher.unregisterUser(user)
}

func (watcher *LpWatcher) registerUser(user *config.User) {
	log.Printf("Leaseplanwatcher for %s: adding user %s(%d)\n", watcher.levelKey, user.FriendlyName, user.UserId)
	watcher.userlist[strconv.FormatInt(user.UserId, 10)] = user
}

func (watcher *LpWatcher) unregisterUser(user *config.User) {
	log.Printf("Leaseplanwatcher for %s: removing user %s(%d)\n", watcher.levelKey, user.FriendlyName, user.UserId)
	delete(watcher.userlist, strconv.FormatInt(user.UserId, 10))
}

func (watcher *LpWatcher) Stop() {
	watcher.isActive = false
}

func (watcher *LpWatcher) Start() {
	watcher.isActive = true

	updateChannel := make(chan []dto.Item)
	go watcher.watch(updateChannel)

	for update := range updateChannel {
		for _, user := range watcher.userlist {
			if user.LeaseplanLevelKey != watcher.levelKey {
				watcher.reallocateUser(user)
				continue
			}
			if user.WatcherActive {
				user.Update(update, tgBot)
			}
		}
	}
}

func (watcher *LpWatcher) watch(itemChannel chan []dto.Item) {
	log.Printf("Leaseplanwatcher for %s: starting\n", watcher.levelKey)
	watcher.isActive = true
	defer func() {
		log.Printf("Leaseplanwatcher for %s: shutdown\n", watcher.levelKey)
		watcher.isActive = false
		close(itemChannel)
	}()

	for watcher.isActive {
		if len(watcher.userlist) == 0 {
			log.Printf("Leaseplanwatcher for %s: has no users in pool -> suspending for 30 sec\n", watcher.levelKey)
			time.Sleep(30 * time.Second)
			continue
		}

		var donorUser *config.User
		for _, user := range watcher.userlist {
			donorUser = user
			if updateUserInfo(user) != nil {
				continue
			}
			if donorUser.LeaseplanLevelKey != watcher.levelKey {
				watcher.reallocateUser(donorUser)
				continue
			}
			break
		}

		if donorUser == nil {
			log.Printf("Leaseplanwatcher for %s: could not select donor user (pool size: %d)\n", watcher.levelKey, len(watcher.userlist))
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("Leaseplanwatcher for %s: using donor token from %s(%d)\n", watcher.levelKey, donorUser.FriendlyName, donorUser.UserId)
		totalRequestsStarted.WithLabelValues(donorUser.FriendlyName).Inc()
		requestStart := time.Now()
		carList, err := pkg.GetAllCars(donorUser.LeaseplanToken, 0, 100)
		requestDuration := time.Since(requestStart)
		requestTime.WithLabelValues(donorUser.FriendlyName).Set(float64(requestDuration.Milliseconds()))
		if err != nil {
			totalRequestErrors.WithLabelValues(donorUser.FriendlyName).Inc()
			log.Printf("Leaseplanwatcher for %s with donor %s(%d): could not get car list %s\n", watcher.levelKey, donorUser.FriendlyName, donorUser.UserId, err)
			continue
		}

		itemChannel <- carList

		log.Printf("Leaseplanwatcher for %s: sleeping for %d minutes\n", watcher.levelKey, globalWatcherDelay)
		for minutesToSleep := globalWatcherDelay; minutesToSleep > 0; minutesToSleep-- {
			for secondsToSleep := 60; secondsToSleep > 0; secondsToSleep -= 5 {
				if !watcher.isActive {
					return
				}
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func updateUserInfo(user *config.User) error {
	lpUserInfo, err := pkg.GetUserInfo(user.LeaseplanToken)
	if err != nil {
		totalRequestErrors.WithLabelValues(user.FriendlyName).Inc()
		log.Printf("Leaseplanwatcher %s(%d): could not get userInfo: %s\n", user.FriendlyName, user.UserId, err)
		return err
	}
	user.LeaseplanLevelKey = lpUserInfo.AddressRole.RoleName
	user.Save()

	return nil
}

func (watcher *LpWatcher) reallocateUser(user *config.User) {
	log.Printf("Leaseplanwatcher for %s is reallocating user %s(%d): level changed to %s\n", watcher.levelKey, user.FriendlyName, user.UserId, user.LeaseplanLevelKey)
	watcher.unregisterUser(user)
	RegisterUserWatcher(user)
}
