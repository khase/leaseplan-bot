package config

import (
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/khase/leaseplanabocarexporter/dto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gopkg.in/yaml.v2"
)

var (
	cacheBasePath            = "cache"
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

type User struct {
	UserMap *UserMap `yaml:"-"`

	UserId            int64  `yaml:"UserId,omitempty"`
	FriendlyName      string `yaml:"FriendlyName,omitempty"`
	LeaseplanToken    string `yaml:"LeaseplanToken,omitempty"`
	LeaseplanLevelKey string `yaml:"LeaseplanLevelKey,omitempty"`

	IsAdmin bool `yaml:"IsAdmin,omitempty"`

	WatcherActive bool  `yaml:"WatcherActive"`
	WatcherDelay  int32 `yaml:"WatcherDelay,omitempty"`

	SummaryMessageTemplate string `yaml:"SummaryMessageTemplate,omitempty"`
	DetailMessageTemplate  string `yaml:"DetailMessageTemplate,omitempty"`

	IgnoreDetails bool `yaml:"IgnoreDetails,omitempty"`
	IgnoreRemoved bool `yaml:"IgnoreRemoved,omitempty"`

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
	user.IgnoreDetails = false
	user.IgnoreRemoved = false
	user.IsAdmin = false
	user.SummaryMessageTemplate = "{{ len .Previous }} -> {{ len .Current }} (+{{ len .Added }}, -{{ len .Removed }})"
	user.DetailMessageTemplate = "{{ portalUrl . }}\n  PS: {{ .RentalObject.PowerHp }}, Antrieb: {{ .RentalObject.KindOfFuel }}\n  BLP: {{ .RentalObject.PriceProducer1 }}€, BGV: {{.SalaryWaiver}}€, Netto: ~{{ round ( netCost . ) 2 }}€\n  Verfügbar: {{.RentalObject.DateRegistration.Format \"02.01.2006\"}}"
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
	user.LastFrame = frame
	if err != nil {
		fmt.Printf("Failed loading usercache for %s: %s\n", user.FriendlyName, err)
	} else {
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

func (user *User) Update(update []dto.Item, bot *tgbotapi.BotAPI) {
	elapsed := time.Since(user.LastFrame.Timestamp)
	if elapsed.Minutes() < float64(user.WatcherDelay) {
		log.Printf("Update for %s(%d): dropped (user throtteling, elapsed time %.2f / %d minutes)", user.FriendlyName, user.UserId, elapsed.Minutes(), user.WatcherDelay)
		return
	}

	userLeaseplanCarsVisible.WithLabelValues(user.FriendlyName).Set(float64(len(update)))
	frame := NewDataFrame(user.LastFrame.Current, update)
	log.Printf("Update for %s(%d): found differences: +%d, -%d", user.FriendlyName, user.UserId, len(frame.Added), len(frame.Removed))

	if frame.HasChanges {
		messages, err := frame.GetMessages(user)
		if err != nil {
			log.Printf("Update for %s(%d): got an error: %s", user.FriendlyName, user.UserId, err)
		}

		totalMessagesSent.WithLabelValues(user.FriendlyName).Add(float64(len(messages)))
		go func() {
			if !user.IsAdmin {
				time.Sleep(5 * time.Minute)
			}
			for _, message := range messages {
				bot.Send(message)
			}
		}()

		user.LastFrame = frame
		user.SaveUserCache()
	}
}
