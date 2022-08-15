package config

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"os"

	"github.com/Masterminds/sprig"
	"github.com/khase/leaseplanabocarexporter/dto"
	"gopkg.in/yaml.v2"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type DataFrame struct {
	Previous []dto.Item `yaml:"Previous,omitempty"`
	Current  []dto.Item `yaml:"Current,omitempty"`
	Added    []dto.Item `yaml:"Added,omitempty"`
	Removed  []dto.Item `yaml:"Removed,omitempty"`

	HasChanges bool `yaml:"HasChanges,omitempty"`
}

func NewEmptyDataFrame() *DataFrame {
	frame := new(DataFrame)
	frame.Previous = []dto.Item{}
	frame.Current = []dto.Item{}
	frame.Added = []dto.Item{}
	frame.Removed = []dto.Item{}
	frame.HasChanges = false

	return frame
}

func NewDataFrame(previous []dto.Item, current []dto.Item) *DataFrame {
	frame := NewEmptyDataFrame()
	frame.Previous = previous
	frame.Current = current

	added, removed := getItemDiff(previous, current)

	frame.Added = added
	frame.Removed = removed

	frame.HasChanges = len(added) > 0 || len(removed) > 0

	return frame
}

func LoadDataFrameFile(path string) (*DataFrame, error) {
	frame := NewEmptyDataFrame()

	strData, err := os.ReadFile(path)
	if err != nil {
		return frame, err
	}

	err = yaml.Unmarshal(strData, frame)
	if err != nil {
		return frame, err
	}

	if err != nil {
		return frame, err
	}

	return frame, nil
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

func (dataFrame *DataFrame) SaveToFile(path string) error {
	data, err := yaml.Marshal(dataFrame)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data, 0600)
	if err != nil {
		return err
	}

	return nil
}

func (dataFrame *DataFrame) GetMessages(user *User) ([]tgbotapi.Chattable, error) {
	return dataFrame.getMessagesInternal(user, 0)
}

func (dataFrame *DataFrame) GetTestMessages(user *User, testLength int) ([]tgbotapi.Chattable, error) {
	return dataFrame.getMessagesInternal(user, testLength)
}

func (dataFrame *DataFrame) getMessagesInternal(user *User, testLength int) ([]tgbotapi.Chattable, error) {
	messages := make([]tgbotapi.Chattable, 0)
	summaryMessage, err := dataFrame.getSummaryMessage(user)
	if err != nil {
		return nil, err
	}
	messages = append(messages, summaryMessage)

	detailMessages, err := dataFrame.getDetailMessages(user, testLength)
	if err != nil {
		return nil, err
	}
	for _, msg := range detailMessages {
		messages = append(messages, msg)
	}

	return messages, nil
}

func (dataFrame *DataFrame) getSummaryMessage(user *User) (tgbotapi.Chattable, error) {
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

func (dataFrame *DataFrame) getDetailMessages(user *User, testLength int) ([]tgbotapi.Chattable, error) {
	added, err := getCarsDetailsTexts(dataFrame.Added, user.DetailMessageTemplate)
	if err != nil {
		return nil, err
	}

	if len(added) < testLength {
		if len(dataFrame.Current) <= 0 {
			return nil, errors.New("Test impossible, no data yet")
		}
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
		if len(dataFrame.Current) <= 0 {
			return nil, errors.New("Test impossible, no data yet")
		}
		for i := len(removed); i < testLength; i++ {
			line, err := getCarDetails(&dataFrame.Current[rand.Intn(len(dataFrame.Current))], user.DetailMessageTemplate)
			if err != nil {
				return nil, err
			}
			removed = append(removed, line)
		}
	}

	messages := make([]tgbotapi.Chattable, 0)
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

	if buf.Len() > 0 {
		messages = append(messages, createMessageAndResetBuffer(buf, user.UserId))
	}
	return messages, nil
}

func addMessageLine(buffer *bytes.Buffer, line string, userId int64, messages []tgbotapi.Chattable) []tgbotapi.Chattable {
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
	tmpl, err := template.
		New("Template").
		Funcs(sprig.FuncMap()).
		Funcs(template.FuncMap{
			"portalUrl": portalUrl,
			"taxPrice":  taxPrice,
			"netCost":   netCost,
			"italic":    italic,
			"bold":      bold,
		}).
		Parse(templateString)

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

func taxPrice(car dto.Item) float64 {
	taxRate := 0.01 // -> Diesel / Benzin
	if car.RentalObject.KindOfFuel == "Plug-in-Hybrid" {
		taxRate = 0.005
	}
	if car.RentalObject.KindOfFuel == "Elektro" {
		if car.RentalObject.PriceProducer1 <= 60000 {
			taxRate = 0.0025
		} else {
			taxRate = 0.005
		}
	}
	return car.RentalObject.PriceProducer1 * taxRate
}

func netCost(car dto.Item) float64 {
	taxFactor := 0.42
	return (taxPrice(car) * taxFactor) + (car.SalaryWaiver * (1 - taxFactor))
}

func italic(text string) string {
	return fmt.Sprintf("_%s_", text)
}

func bold(text string) string {
	return fmt.Sprintf("*%s*", text)
}

func portalUrl(car dto.Item) string {
	return fmt.Sprintf("[%s](https://www.leaseplan-abocar.de/offer-details/%s/%s)", car.OfferTypeName, car.Ident, car.RentalObject.Ident)
}
