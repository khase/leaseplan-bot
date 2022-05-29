package config

import (
	"log"
	"time"

	"github.com/khase/leaseplanabocarexporter/dto"
	"github.com/khase/leaseplanabocarexporter/pkg"
	"gopkg.in/yaml.v2"
)

type User struct {
	UserMap *UserMap `yaml:"-"`

	UserId         int64  `yaml:"UserId,omitempty"`
	FriendlyName   string `yaml:"FriendlyName,omitempty"`
	LeaseplanToken string `yaml:"LeaseplanToken,omitempty"`

	WatcherActive  bool  `yaml:"WatcherActive"`
	WatcherRunning bool  `yaml:"-"`
	WatcherDelay   int32 `yaml:"WatcherDelay,omitempty"`
}

func NewUser(userMap *UserMap, userId int64, friendlyName string) *User {
	user := new(User)
	user.UserMap = userMap
	user.UserId = userId
	user.FriendlyName = friendlyName
	user.LeaseplanToken = ""
	user.WatcherActive = false
	user.WatcherRunning = false
	user.WatcherDelay = 15
	return user
}

func (user *User) GetHumanReadableUserInfo() (string, error) {
	data, err := yaml.Marshal(user)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (user *User) Save() {
	user.UserMap.Save()
}

func (user *User) StartWatcher() {
	user.WatcherActive = true
	if !user.WatcherRunning {
		go user.watch()
	}
}

func (user *User) StopWatcher() {
	user.WatcherActive = false
}

func (user *User) watch() {
	log.Printf("Watcher for %s(%d): started.", user.FriendlyName, user.UserId)
	user.WatcherRunning = true
	defer func() {
		user.WatcherRunning = false
		log.Printf("Watcher for %s(%d): shutdown", user.FriendlyName, user.UserId)
	}()

	lastCarList := []dto.Item{}
	for user.WatcherActive {
		currentCarList, err := pkg.GetAllCars(user.LeaseplanToken, 0, 20)
		if err != nil {
			log.Printf("Watcher for %s(%d): got an error: %s", user.FriendlyName, user.UserId, err)
		}

		log.Printf("Watcher for %s(%d): got %d car items", user.FriendlyName, user.UserId, len(currentCarList))
		added, removed := getItemDiff(lastCarList, currentCarList)
		log.Printf("Watcher for %s(%d): found differences: +%d, -%d", user.FriendlyName, user.UserId, len(added), len(removed))

		lastCarList = currentCarList

		log.Printf("Watcher for %s(%d): sleeps for %d minutes", user.FriendlyName, user.UserId, user.WatcherDelay)
		for i := user.WatcherDelay; i > 0; i-- {
			if !user.WatcherActive {
				break
			}
			time.Sleep(time.Minute)
		}
	}

	log.Printf("Watcher for %s(%d): has been disabled", user.FriendlyName, user.UserId)
}

func getItemDiff(previous []dto.Item, current []dto.Item) (added []dto.Item, removed []dto.Item) {
	previousMap := make(map[string]dto.Item)
	for _, element := range previous {
		previousMap[element.RentalObject.Ident] = element
	}
	currentMap := make(map[string]dto.Item)
	for _, element := range current {
		currentMap[element.RentalObject.Ident] = element
	}

	added = []dto.Item{}
	for key, element := range currentMap {
		_, exists := previousMap[key]
		if !exists {
			added = append(added, element)
		}
	}

	removed = []dto.Item{}
	for key, element := range previousMap {
		_, exists := currentMap[key]
		if !exists {
			removed = append(removed, element)
		}
	}

	return added, removed
}
