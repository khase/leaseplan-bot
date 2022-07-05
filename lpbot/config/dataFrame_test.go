package config_test

import (
	"log"
	"strings"
	"testing"

	"github.com/khase/leaseplan-bot/lpbot/config"
	"github.com/khase/leaseplanabocarexporter/dto"
	"gopkg.in/yaml.v2"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestDataFrameLoading(t *testing.T) {
	frame, err := config.LoadDataFrameFile("../../testdata/dataframe.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if frame == nil {
		t.Fatalf("could not load dataframe")
	}
}

func TestMessage(t *testing.T) {
	frame, err := config.LoadDataFrameFile("../../testdata/nochange.dataframe.yaml")
	if err != nil {
		t.Fatal(err)
	}

	messages, err := frame.GetMessages(&config.User{
		UserId:                 123,
		SummaryMessageTemplate: "{{ len .Previous }} -> {{ len .Current }} (+{{ len .Added }}, -{{ len .Removed }})",
		DetailMessageTemplate:  "[{{ .OfferTypeName }}](https://www.leaseplan-abocar.de/offer-details/{{ .Ident }}/{{ .RentalObject.Ident }}) {{ .RentalObject.PowerHp }}PS ({{ .RentalObject.PriceProducer1 }}€)",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(messages) != 1 {
		data, _ := yaml.Marshal(messages)
		log.Printf(string(data))
		t.Fatalf("expected 1 message but got %d", len(messages))
	}
}

func TestTestMessage(t *testing.T) {
	frame, err := config.LoadDataFrameFile("../../testdata/nochange.dataframe.yaml")
	if err != nil {
		t.Fatal(err)
	}

	messages, err := frame.GetTestMessages(&config.User{
		UserId:                 123,
		SummaryMessageTemplate: "{{ len .Previous }} -> {{ len .Current }} (+{{ len .Added }}, -{{ len .Removed }})",
		DetailMessageTemplate:  "[{{ .OfferTypeName }}](https://www.leaseplan-abocar.de/offer-details/{{ .Ident }}/{{ .RentalObject.Ident }}) {{ .RentalObject.PowerHp }}PS ({{ .RentalObject.PriceProducer1 }}€)",
	}, 2)
	if err != nil {
		t.Fatal(err)
	}

	if len(messages) != 2 {
		data, _ := yaml.Marshal(messages)
		log.Printf(string(data))
		t.Fatalf("expected 2 message but got %d", len(messages))
	}
}

func TestTaxNetto(t *testing.T) {
	prev := []dto.Item{}
	cur := []dto.Item{
		{
			RentalObject: dto.RentalObject{
				Ident:          "1",
				KindOfFuel:     "Benzin",
				PriceProducer1: 90000,
			},
			SalaryWaiver: 500,
		},
		{
			RentalObject: dto.RentalObject{
				Ident:          "2",
				KindOfFuel:     "Diesel",
				PriceProducer1: 90000,
			},
			SalaryWaiver: 500,
		},
		{
			RentalObject: dto.RentalObject{
				Ident:          "3",
				KindOfFuel:     "Plug-in-Hybrid",
				PriceProducer1: 90000,
			},
			SalaryWaiver: 500,
		},
		{
			RentalObject: dto.RentalObject{
				Ident:          "4",
				KindOfFuel:     "Elektro",
				PriceProducer1: 90000,
			},
			SalaryWaiver: 500,
		},
		{
			RentalObject: dto.RentalObject{
				Ident:          "5",
				KindOfFuel:     "Elektro",
				PriceProducer1: 60000,
			},
			SalaryWaiver: 500,
		},
		{
			RentalObject: dto.RentalObject{
				Ident:          "6",
				KindOfFuel:     "Elektro",
				PriceProducer1: 60001,
			},
			SalaryWaiver: 500,
		},
	}
	frame := config.NewDataFrame(prev, cur)

	messages, err := frame.GetMessages(&config.User{
		UserId:                 123,
		SummaryMessageTemplate: "{{ len .Previous }} -> {{ len .Current }} (+{{ len .Added }}, -{{ len .Removed }})",
		DetailMessageTemplate:  "{{ .RentalObject.KindOfFuel | toString | italic }} {{ .RentalObject.PriceProducer1 | toString | bold }}€ -> {{ round ( taxPrice . ) 2 }}€ / {{ round ( netCost . ) 2 }}€",
	})

	if err != nil {
		t.Fatal(err)
	}

	if len(messages) != 2 {
		data, _ := yaml.Marshal(messages)
		log.Printf(string(data))
		t.Fatalf("expected 2 message but got %d", len(messages))
	}

	if messages[0].(tgbotapi.MessageConfig).Text != "0 -> 6 (+6, -0)" {
		t.Fatalf("message 1 should be \"0 -> 6 (+6, -0)\" but got \"%s\"", messages[0].(tgbotapi.MessageConfig).Text)
	}

	contents := []string{
		"_Benzin_ *90000*€ -> 900€ / 668€",
		"_Diesel_ *90000*€ -> 900€ / 668€",
		"_Plug-in-Hybrid_ *90000*€ -> 450€ / 479€",
		"_Elektro_ *90000*€ -> 450€ / 479€",
		"_Elektro_ *60000*€ -> 150€ / 353€",
		"_Elektro_ *60001*€ -> 300.01€ / 416€",
	}

	for _, val := range contents {
		if !strings.Contains(messages[1].(tgbotapi.MessageConfig).Text, val) {
			t.Fatalf("message 1 should contain \"%s\" but didn't. Text \"%s\"", val, messages[1].(tgbotapi.MessageConfig).Text)
		}
	}
}
