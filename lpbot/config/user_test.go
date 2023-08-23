package config_test

import (
	"log"
	"testing"

	"github.com/khase/leaseplan-bot/lpbot/config"
	"github.com/khase/leaseplanabocarexporter/dto"
	"gopkg.in/yaml.v2"
)

func TestFilter(t *testing.T) {
	cars := []dto.Item{
		{
			RentalObject: dto.RentalObject{
				CarLabel: "Volvo",
			},
		},
		{
			RentalObject: dto.RentalObject{
				CarLabel: "BMW",
			},
		},
	}

	filteredList := config.FilterUpdateList(cars, []string{
		"ne (.RentalObject.CarLabel | lower) \"volvo\"",
	})

	if len(filteredList) > 1 {
		data, _ := yaml.Marshal(filteredList)
		log.Printf(string(data))
		t.Fatalf("expected 1 cars but got %d", len(filteredList))
	}

	if filteredList[0].RentalObject.CarLabel == "Volvo" {
		data, _ := yaml.Marshal(filteredList)
		log.Printf(string(data))
		t.Fatalf("expected list not to contain cars from Volvo")
	}
}
