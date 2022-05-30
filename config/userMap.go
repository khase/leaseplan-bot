package config

import (
	"errors"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UserMap struct {
	Path  string          `yaml:"-"`
	Users map[int64]*User `yaml:"Users,omitempty"`
}

func NewUserMap(userDataFile string) *UserMap {
	userMap := new(UserMap)
	userMap.Path = userDataFile
	userMap.Users = make(map[int64]*User)
	return userMap
}

func LoadUserMap(userDataFile string) (*UserMap, error) {
	userMap := NewUserMap(userDataFile)
	err := userMap.LoadFromFile(userDataFile)
	if err != nil {
		return userMap, err
	}

	log.Printf("Loading user file from %s", userDataFile)

	return userMap, nil
}

func (userMap *UserMap) LoadFromFile(userDataFile string) error {
	strData, err := os.ReadFile(userDataFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(strData, userMap)
	if err != nil {
		return err
	}

	userMap.fixUserBackReference()
	if err != nil {
		return err
	}

	return nil
}

func (userMap *UserMap) Reload() error {
	if userMap.Path == "" {
		return errors.New("Path of userMap has not been set")
	}

	err := userMap.LoadFromFile(userMap.Path)
	if err != nil {
		return err
	}

	return nil
}

func (userMap *UserMap) SaveToFile(userDataFile string) error {
	data, err := yaml.Marshal(userMap)
	if err != nil {
		return err
	}

	err = os.WriteFile(userDataFile, data, 0600)
	if err != nil {
		return err
	}

	return nil
}

func (userMap *UserMap) Save() error {
	if userMap.Path == "" {
		return errors.New("Path of userMap has not been set")
	}

	err := userMap.SaveToFile(userMap.Path)
	if err != nil {
		return err
	}

	return nil
}

func (userMap *UserMap) CreateNewUser(userId int64, friendlyName string) (*User, error) {
	_, exists := userMap.Users[userId]
	if exists {
		return nil, errors.New(fmt.Sprintf("User with ID: %d does already exist", userId))
	}

	newUser := NewUser(userMap, userId, friendlyName)
	userMap.Users[userId] = newUser

	err := userMap.Save()
	if err != nil {
		return newUser, nil
	}

	return newUser, nil
}

func (userMap *UserMap) fixUserBackReference() error {
	for _, user := range userMap.Users {
		user.UserMap = userMap
	}

	return nil
}

func (userMap *UserMap) StartActiveWatchers(bot *tgbotapi.BotAPI) error {
	for _, user := range userMap.Users {
		if user.WatcherActive {
			user.StartWatcher(bot)
		}
	}

	return nil
}
