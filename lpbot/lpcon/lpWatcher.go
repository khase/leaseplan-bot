package lpcon

import (
	"log"
	"time"

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
)

type LpWatcher struct {
	user *config.User

	isActive bool
}

func NewLpWatcher(user *config.User) *LpWatcher {
	watcher := new(LpWatcher)
	watcher.user = user

	watcher.isActive = false

	return watcher
}

func (watcher *LpWatcher) Stop() {
	watcher.isActive = false
}

func (watcher *LpWatcher) Watch(itemChannel chan []dto.Item) {
	log.Printf("Leaseplanwatcher for %s(%d): starting\n", watcher.user.FriendlyName, watcher.user.UserId)
	watcher.isActive = true
	defer func() {
		log.Printf("Leaseplanwatcher for %s(%d): shutdown\n", watcher.user.FriendlyName, watcher.user.UserId)
		watcher.isActive = false
		close(itemChannel)
	}()

	for watcher.isActive {
		totalRequestsStarted.WithLabelValues(watcher.user.FriendlyName).Inc()
		requestStart := time.Now()
		carList, err := pkg.GetAllCars(watcher.user.LeaseplanToken, 0, 90)
		requestDuration := time.Since(requestStart)
		requestTime.WithLabelValues(watcher.user.FriendlyName).Set(float64(requestDuration.Milliseconds()))
		if err != nil {
			totalRequestErrors.WithLabelValues(watcher.user.FriendlyName).Inc()
			log.Printf("Leaseplanwatcher for %s(%d): could not get car list %s\n", watcher.user.FriendlyName, watcher.user.UserId, err)
			continue
		}

		itemChannel <- carList

		log.Printf("Leaseplanwatcher for %s(%d): sleeping for %d minutes\n", watcher.user.FriendlyName, watcher.user.UserId, watcher.user.WatcherDelay)
		for minutesToSleep := watcher.user.WatcherDelay; minutesToSleep > 0; minutesToSleep-- {
			for secondsToSleep := 60; secondsToSleep > 0; secondsToSleep -= 5 {
				if !watcher.isActive {
					return
				}
				time.Sleep(5 * time.Second)
			}
		}
	}
}
