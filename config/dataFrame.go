package config

import "github.com/khase/leaseplanabocarexporter/dto"

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
