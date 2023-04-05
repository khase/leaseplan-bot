package lpcon

import (
	"log"
	"math/rand"
	"reflect"
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
			"key",
		})
	totalRequestErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lpcon_total_request_errors",
			Help: "The total number of requests returned an error",
		},
		[]string{
			"username",
			"key",
		})
	requestTime = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lpcon_request_time",
			Help: "The duration in ms the leaseplan request took to finish",
		},
		[]string{
			"username",
			"key",
		})
	watcherLeaseplanCarsVisible = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lpcon_cars_visible_to_group",
			Help: "Number of cars visible to the group",
		},
		[]string{
			"key",
		})

	watcherList        map[string]*LpWatcher = make(map[string]*LpWatcher)
	tgBot              *tgbotapi.BotAPI
	globalWatcherDelay int
	watcherPageSize    int
)

type LpWatcher struct {
	levelKey string

	userlist       map[string]*config.User
	currentCarList []dto.Item

	state *LpWatcherState
}

type LpWatcherState struct {
	UserCount       int `json:"UserCount,omitempty"`
	CurrentCarCount int `json:"CurrentCarCount,omitempty"`
	Poll            struct {
		StartTime string `json:"StartTime,omitempty"`
		IsActive  bool   `json:"IsActive,omitempty"`
		Duration  string `json:"Duration,omitempty"`
	} `json:"Poll,omitempty"`
	IsActive bool `json:"IsActive,omitempty"`
}

func NewLpWatcher(levelKey string) *LpWatcher {
	watcher := new(LpWatcher)
	watcher.levelKey = levelKey
	watcher.userlist = make(map[string]*config.User)

	watcher.state = &LpWatcherState{
		UserCount:       0,
		CurrentCarCount: 0,

		IsActive: false,
	}

	return watcher
}

func SetTgBotForWatcher(bot *tgbotapi.BotAPI) {
	tgBot = bot
}

func SetWatcherDelay(delay int) {
	globalWatcherDelay = delay
}

func SetWatcherPageSize(pageZize int) {
	watcherPageSize = pageZize
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

func GetStates() map[string]*LpWatcherState {
	result := make(map[string]*LpWatcherState)

	for key, element := range watcherList {
		result[key] = element.state
	}

	return result
}

func GetCars() map[string][]dto.Item {
	result := make(map[string][]dto.Item)

	for key, element := range watcherList {
		result[key] = element.currentCarList
	}

	return result
}

func GetWatcherKeys() []string {
	result := make([]string, 0, len(watcherList))
	for key := range watcherList {
		result = append(result, key)
	}

	return result
}

func (watcher *LpWatcher) registerUser(user *config.User) {
	log.Printf("Leaseplanwatcher for %s: adding user %s(%d)\n", watcher.levelKey, user.FriendlyName, user.UserId)
	watcher.userlist[strconv.FormatInt(user.UserId, 10)] = user

	watcher.state.UserCount = len(watcher.userlist)
}

func (watcher *LpWatcher) unregisterUser(user *config.User) {
	log.Printf("Leaseplanwatcher for %s: removing user %s(%d)\n", watcher.levelKey, user.FriendlyName, user.UserId)
	delete(watcher.userlist, strconv.FormatInt(user.UserId, 10))

	watcher.state.UserCount = len(watcher.userlist)
}

func (watcher *LpWatcher) Stop() {
	watcher.state.IsActive = false
}

func (watcher *LpWatcher) Start() {
	watcher.state.IsActive = true

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
	watcher.state.IsActive = true
	defer func() {
		log.Printf("Leaseplanwatcher for %s: shutdown\n", watcher.levelKey)
		watcher.state.IsActive = false
		close(itemChannel)
	}()

	for watcher.state.IsActive {
		if len(watcher.userlist) == 0 {
			log.Printf("Leaseplanwatcher for %s: has no users in pool -> suspending for 30 sec\n", watcher.levelKey)
			time.Sleep(30 * time.Second)
			continue
		}

		var donorUser *config.User
		// select donor token
		rand.Seed(time.Now().Unix())
		userIds := reflect.ValueOf(watcher.userlist).MapKeys()
		for len(userIds) > 0 {
			idx := rand.Intn(len(userIds))
			userId := userIds[idx].String()
			user := watcher.userlist[userId]
			if !user.WatcherActive {
				userIds[idx] = userIds[len(userIds)-1]
				userIds = userIds[:len(userIds)-1]
				continue
			}
			if !user.EULA {
				user.WatcherError = "EULA not accepted. Accept with /eula true"
				userIds[idx] = userIds[len(userIds)-1]
				userIds = userIds[:len(userIds)-1]
				continue
			}
			if updateUserInfo(user) != nil {
				userIds[idx] = userIds[len(userIds)-1]
				userIds = userIds[:len(userIds)-1]
				continue
			}
			if user.LeaseplanLevelKey != watcher.levelKey {
				watcher.reallocateUser(user)
				userIds[idx] = userIds[len(userIds)-1]
				userIds = userIds[:len(userIds)-1]
				continue
			}
			donorUser = user
			break
		}

		if donorUser == nil {
			log.Printf("Leaseplanwatcher for %s: could not select donor user (pool size: %d)\n", watcher.levelKey, len(watcher.userlist))
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("Leaseplanwatcher for %s: using donor token from %s(%d)\n", watcher.levelKey, donorUser.FriendlyName, donorUser.UserId)
		totalRequestsStarted.WithLabelValues(donorUser.FriendlyName, watcher.levelKey).Inc()

		requestStart := time.Now()
		watcher.state.Poll.StartTime = requestStart.UTC().String()
		watcher.state.Poll.Duration = ""
		watcher.state.Poll.IsActive = true

		carList, err := pkg.GetAllCars(donorUser.LeaseplanToken, 0, watcherPageSize)

		requestDuration := time.Since(requestStart)
		watcher.state.Poll.Duration = requestDuration.String()
		watcher.state.Poll.IsActive = false

		log.Printf("Update for %s: got %d car items", watcher.levelKey, len(carList))
		watcherLeaseplanCarsVisible.WithLabelValues(watcher.levelKey).Set(float64(len(carList)))
		requestTime.WithLabelValues(donorUser.FriendlyName, watcher.levelKey).Set(float64(requestDuration.Milliseconds()))

		if err != nil {
			totalRequestErrors.WithLabelValues(donorUser.FriendlyName, watcher.levelKey).Inc()
			log.Printf("Leaseplanwatcher for %s with donor %s(%d): could not get car list %s\n", watcher.levelKey, donorUser.FriendlyName, donorUser.UserId, err)
		} else {
			watcher.currentCarList = carList
			watcher.state.CurrentCarCount = len(carList)

			itemChannel <- carList
		}

		log.Printf("Leaseplanwatcher for %s: sleeping for %d minutes\n", watcher.levelKey, globalWatcherDelay)
		for minutesToSleep := globalWatcherDelay; minutesToSleep > 0; minutesToSleep-- {
			for secondsToSleep := 60; secondsToSleep > 0; secondsToSleep -= 5 {
				if !watcher.state.IsActive {
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
		totalRequestErrors.WithLabelValues(user.FriendlyName, user.LeaseplanLevelKey).Inc()
		log.Printf("Leaseplanwatcher %s(%d): could not get userInfo: %s\n", user.FriendlyName, user.UserId, err)
		user.WatcherError = err.Error()
		user.WatcherActive = false
		user.Save()
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
