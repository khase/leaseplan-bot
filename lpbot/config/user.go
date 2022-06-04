package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

var (
	cacheBasePath = "cache"
)

type User struct {
	UserMap *UserMap `yaml:"-"`

	UserId         int64  `yaml:"UserId,omitempty"`
	FriendlyName   string `yaml:"FriendlyName,omitempty"`
	LeaseplanToken string `yaml:"LeaseplanToken,omitempty"`

	WatcherActive bool  `yaml:"WatcherActive"`
	WatcherDelay  int32 `yaml:"WatcherDelay,omitempty"`

	SummaryMessageTemplate string `yaml:"SummaryMessageTemplate,omitempty"`
	DetailMessageTemplate  string `yaml:"DetailMessageTemplate,omitempty"`

	LastFrame *DataFrame `yaml:"-"`
}

func NewUser(userMap *UserMap, userId int64, friendlyName string) *User {
	user := new(User)
	user.UserMap = userMap
	user.UserId = userId
	user.FriendlyName = friendlyName
	user.LeaseplanToken = ""
	user.WatcherActive = false
	user.WatcherDelay = 15
	user.SummaryMessageTemplate = "{{ len .Previous }} -> {{ len .Current }} (+{{ len .Added }}, -{{ len .Removed }})"
	user.DetailMessageTemplate = "[{{ .OfferTypeName }}](https://www.leaseplan-abocar.de/offer-details/{{ .Ident }}/{{ .RentalObject.Ident }}) {{ .RentalObject.PowerHp }}PS ({{ .RentalObject.PriceProducer1 }}€)"
	user.LastFrame = NewDataFrame(nil, nil)
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

func (user *User) LoadUserCache() {
	frame, err := LoadDataFrameFile(fmt.Sprintf("%s/%d.lastframe", cacheBasePath, user.UserId))
	if err != nil {
		fmt.Printf("Failed loading usercache for %s: %s\n", user.FriendlyName, err)
	} else {
		user.LastFrame = frame
		fmt.Printf("Loaded usercache for %s: %d -> %d (+%d, -%d)\n", user.FriendlyName, len(user.LastFrame.Previous), len(user.LastFrame.Current), len(user.LastFrame.Added), len(user.LastFrame.Removed))
	}
}

func (user *User) SaveUserCache() {
	os.MkdirAll(cacheBasePath, os.ModePerm)
	user.LastFrame.SaveToFile(fmt.Sprintf("%s/%d.lastframe", cacheBasePath, user.UserId))
}

func (user *User) StartWatcher() {
	user.WatcherActive = true
}

func (user *User) StopWatcher() {
	user.WatcherActive = false
}
