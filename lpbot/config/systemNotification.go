package config

import "time"

var (
	SystemNotifications = [1]SystemNotification{
		{
			time.Date(2023, 04, 03, 23, 20, 0, 0, time.Now().Local().Location()),
			"Sicher ist dir schon aufgefallen, dass ich die letzte Zeit offline war 😔\n\nLeider will Leaseplan keine Bots gegen ihre API laufen haben und geht dagegen auch immer stärker aktiv vor. Durch ein angezogenes Rate Limit bin ich ins rampenlicht geraten und einige der User wurden entsprechend auch auf Leaseplan gesperrt. Das ist an der Stelle nicht weiter schlimm, die entsprechenden User können einfach über den Support wieder aktiviert werden. Als reaktion darauf haben wir ein paar änderungen in der Abruflogik vorgenommen um wieder etwas mehr unter dem Radar zu fliegen.\n\nWenn du mit so einem Vorfall keine Probleme hast akzeptiere bitte die /eula true und starte deine Notifications wieder mit /resume.\n\nÜbrigens gibt es jetzt auch einen Community Channel in dem ihr mir Probleme oder Wünsche mitteilen könnt https://t.me/+jCYLQd2KkOdiNmRi 😉",
			func(u *User) bool {
				return u.WatcherActive
			},
			func(u *User) {
				u.WatcherActive = false
			},
		},
	}
)

type SystemNotification struct {
	Publish       time.Time
	Message       string
	UserCondition func(*User) bool
	UserAction    func(*User)
}

func NewSystemNotification(publish time.Time, message string) *SystemNotification {
	notification := new(SystemNotification)

	notification.Publish = publish
	notification.Message = message

	return notification
}
