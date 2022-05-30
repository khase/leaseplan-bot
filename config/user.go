package config

import (
	"bytes"
	"html/template"
	"log"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/khase/leaseplanabocarexporter/dto"
	"github.com/khase/leaseplanabocarexporter/pkg"
	"gopkg.in/yaml.v2"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type User struct {
	UserMap *UserMap `yaml:"-"`

	UserId         int64  `yaml:"UserId,omitempty"`
	FriendlyName   string `yaml:"FriendlyName,omitempty"`
	LeaseplanToken string `yaml:"LeaseplanToken,omitempty"`

	WatcherActive  bool  `yaml:"WatcherActive"`
	WatcherRunning bool  `yaml:"-"`
	WatcherDelay   int32 `yaml:"WatcherDelay,omitempty"`

	SummaryMessageTemplate string `yaml:"SummaryMessageTemplate,omitempty"`
	DetailMessageTemplate  string `yaml:"DetailMessageTemplate,omitempty"`
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
	user.SummaryMessageTemplate = "{{ len(.Previous) }} -> {{ len(.Current) }} (+{{ len(.Added) }}, -{{ len(.Removed) }})"
	user.DetailMessageTemplate = "[{{ .OfferTypeName }}](https://www.leaseplan-abocar.de/offer-details/{{ .Ident }}/{{ .RentalObject.Ident }}) {{ .RentalObject.PowerHp }}PS ({{ .RentalObject.PriceProducer1 }}€)"
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

func (user *User) StartWatcher(bot *tgbotapi.BotAPI) {
	user.WatcherActive = true
	if !user.WatcherRunning {
		go user.watch(bot)
	}
}

func (user *User) StopWatcher() {
	user.WatcherActive = false
}

func (user *User) watch(bot *tgbotapi.BotAPI) {
	log.Printf("Watcher for %s(%d): started.", user.FriendlyName, user.UserId)
	user.WatcherRunning = true
	defer func() {
		user.WatcherRunning = false
		log.Printf("Watcher for %s(%d): shutdown", user.FriendlyName, user.UserId)
	}()

	lastCarList := []dto.Item{}
	isFirstRun := true
	for user.WatcherActive {
		currentCarList, err := pkg.GetAllCars(user.LeaseplanToken, 0, 50)
		if err != nil {
			log.Printf("Watcher for %s(%d): got an error: %s", user.FriendlyName, user.UserId, err)
		}

		log.Printf("Watcher for %s(%d): got %d car items", user.FriendlyName, user.UserId, len(currentCarList))
		frame := NewDataFrame(lastCarList, currentCarList)
		log.Printf("Watcher for %s(%d): found differences: +%d, -%d", user.FriendlyName, user.UserId, len(frame.Added), len(frame.Removed))

		if !isFirstRun && frame.HasChanges {
			summary, err := fillTemplate(user.SummaryMessageTemplate, frame)
			if err != nil {
				log.Printf("Watcher for %s(%d): got an error: %s", user.FriendlyName, user.UserId, err)
			}
			msg := tgbotapi.NewMessage(user.UserId, summary)
			msg.ParseMode = "Markdown"
			bot.Send(msg)

			buf := new(bytes.Buffer)
			if len(frame.Added) > 0 {
				buf.WriteString("Added:\n")
				for _, car := range frame.Added {
					carText, err := fillTemplate(user.DetailMessageTemplate, car)
					if err != nil {
						log.Printf("Watcher for %s(%d): got an error: %s", user.FriendlyName, user.UserId, err)
					}

					if buf.Len()+len(carText) > 3500 {
						msg = tgbotapi.NewMessage(user.UserId, buf.String())
						msg.ParseMode = "Markdown"
						bot.Send(msg)
						buf.Reset()
					}

					buf.WriteString(carText + "\n")
				}
			}
			if len(frame.Removed) > 0 {
				if len(frame.Added) > 0 {
					buf.WriteString("\n")
				}
				buf.WriteString("Removed:\n")
				for _, car := range frame.Removed {
					carText, err := fillTemplate(user.DetailMessageTemplate, car)
					if err != nil {
						log.Printf("Watcher for %s(%d): got an error: %s", user.FriendlyName, user.UserId, err)
					}
					buf.WriteString(carText + "\n")

					if buf.Len()+len(carText) > 3500 {
						msg = tgbotapi.NewMessage(user.UserId, buf.String())
						msg.ParseMode = "Markdown"
						bot.Send(msg)
						buf.Reset()
					}
				}
			}
			msg = tgbotapi.NewMessage(user.UserId, buf.String())
			msg.ParseMode = "Markdown"
			bot.Send(msg)
		}

		lastCarList = currentCarList
		isFirstRun = false

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

func fillTemplate(templateString string, input interface{}) (string, error) {
	tmpl, err := template.New("Template").Funcs(sprig.FuncMap()).Parse(templateString)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)

	err = tmpl.Execute(buf, input)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
