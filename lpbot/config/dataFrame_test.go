package config_test

import (
	"log"
	"testing"

	"github.com/khase/leaseplan-bot/lpbot/config"
	"gopkg.in/yaml.v2"
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
