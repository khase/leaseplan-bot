package lpcon

import (
	"log"
	"time"

	"github.com/khase/leaseplan-bot/lpbot/config"
	"github.com/khase/leaseplanabocarexporter/dto"
	"github.com/khase/leaseplanabocarexporter/pkg"
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
	log.Printf("Leaseplanwatcher for %s: starting\n", watcher.user.FriendlyName)
	watcher.isActive = true
	defer func() {
		log.Printf("Leaseplanwatcher for %s: shutdown\n", watcher.user.FriendlyName)
		watcher.isActive = false
		close(itemChannel)
	}()

	for watcher.isActive {
		carList, err := pkg.GetAllCars(watcher.user.LeaseplanToken, 0, 50)
		if err != nil {
			log.Printf("Leaseplanwatcher for %s: could not get car list %s\n", watcher.user.FriendlyName, err)
		}

		itemChannel <- carList

		log.Printf("Leaseplanwatcher for %s: sleeping for %d minutes\n", watcher.user.FriendlyName, watcher.user.WatcherDelay)
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
