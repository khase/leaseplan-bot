package config

import (
	"bytes"
	"fmt"
	"html/template"
	"math/rand"

	"github.com/Masterminds/sprig"
	"github.com/khase/leaseplanabocarexporter/dto"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type DataFrame struct {
	Previous []dto.Item `yaml:"Previous,omitempty"`
	Current  []dto.Item `yaml:"Current,omitempty"`
	Added    []dto.Item `yaml:"Added,omitempty"`
	Removed  []dto.Item `yaml:"Removed,omitempty"`

	HasChanges bool `yaml:"Removed,omitempty"`
}

func NewDataFrame(previous []dto.Item, current []dto.Item) *DataFrame {
	frame := new(DataFrame)
	frame.Previous = previous
	frame.Current = current

	added, removed := getItemDiff(previous, current)

	frame.Added = added
	frame.Removed = removed

	frame.HasChanges = len(added) > 0 || len(removed) > 0

	return frame
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

func (dataFrame *DataFrame) GetMessages(user *User) ([]tgbotapi.MessageConfig, error) {
	return dataFrame.getMessagesInternal(user, 0)
}

func (dataFrame *DataFrame) GetTestMessages(user *User, testLength int) ([]tgbotapi.MessageConfig, error) {
	return dataFrame.getMessagesInternal(user, testLength)
}

func (dataFrame *DataFrame) getMessagesInternal(user *User, testLength int) ([]tgbotapi.MessageConfig, error) {
	messages := make([]tgbotapi.MessageConfig, 1)
	summaryMessage, err := dataFrame.getSummaryMessage(user)
	if err != nil {
		return nil, err
	}
	messages = append(messages, summaryMessage)

	detailMessages, err := dataFrame.getDetailMessages(user, testLength)
	if err != nil {
		return nil, err
	}
	messages = append(messages, detailMessages...)

	return messages, nil
}

func (dataFrame *DataFrame) getSummaryMessage(user *User) (tgbotapi.MessageConfig, error) {
	summary, err := dataFrame.getSummaryText(user.SummaryMessageTemplate)
	if err != nil {
		return tgbotapi.MessageConfig{}, err
	}

	msg := tgbotapi.NewMessage(user.UserId, summary)
	return msg, nil
}

func (dataFrame *DataFrame) getSummaryText(template string) (string, error) {
	summaryString, err := fillTemplate(template, dataFrame)
	if err != nil {
		return "", err
	}

	return summaryString, nil
}

func (dataFrame *DataFrame) getDetailMessages(user *User, testLength int) ([]tgbotapi.MessageConfig, error) {
	added, err := getCarsDetailsTexts(dataFrame.Added, user.DetailMessageTemplate)
	if err != nil {
		return nil, err
	}

	if len(added) < testLength {
		for i := len(added); i < testLength; i++ {
			line, err := getCarDetails(&dataFrame.Current[rand.Intn(len(dataFrame.Current))], user.DetailMessageTemplate)
			if err != nil {
				return nil, err
			}
			added = append(added, line)
		}
	}

	removed, err := getCarsDetailsTexts(dataFrame.Removed, user.DetailMessageTemplate)
	if err != nil {
		return nil, err
	}

	if len(removed) < testLength {
		for i := len(removed); i < testLength; i++ {
			line, err := getCarDetails(&dataFrame.Current[rand.Intn(len(dataFrame.Current))], user.DetailMessageTemplate)
			if err != nil {
				return nil, err
			}
			removed = append(removed, line)
		}
	}

	messages := make([]tgbotapi.MessageConfig, 0)
	buf := new(bytes.Buffer)

	if len(added) > 0 {
		buf.WriteString("Added:\n")
		for _, line := range added {
			messages = addMessageLine(buf, line, user.UserId, messages)
		}
	}
	if len(removed) > 0 {
		if len(added) > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString("Removed:\n")
		for _, line := range removed {
			messages = addMessageLine(buf, line, user.UserId, messages)
		}
	}

	messages = append(messages, createMessageAndResetBuffer(buf, user.UserId))
	return messages, nil
}

func addMessageLine(buffer *bytes.Buffer, line string, userId int64, messages []tgbotapi.MessageConfig) []tgbotapi.MessageConfig {
	if buffer.Len()+len(line) > 3500 {
		messages = append(messages, createMessageAndResetBuffer(buffer, userId))
	}
	buffer.WriteString(fmt.Sprintf("%s\n", line))

	return messages
}

func createMessageAndResetBuffer(buffer *bytes.Buffer, userId int64) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(userId, buffer.String())
	msg.ParseMode = "Markdown"
	buffer.Reset()

	return msg
}

func getCarsDetailsTexts(cars []dto.Item, template string) ([]string, error) {
	result := make([]string, 0)
	for _, car := range cars {
		detailString, err := getCarDetails(&car, template)
		if err != nil {
			return nil, err
		}
		result = append(result, detailString)
	}

	return result, nil
}

func getCarDetails(car *dto.Item, template string) (string, error) {
	detailString, err := fillTemplate(template, car)
	if err != nil {
		return "", err
	}

	return detailString, nil
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
