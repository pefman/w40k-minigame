package game

import (
	"encoding/json"
	"os"
)

type Unit struct {
	FactionName string `json:"factionname"`
	UnitName    string `json:"unitname"`
}

func LoadUnitsFromFile(path string) ([]Unit, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var units []Unit
	err = json.Unmarshal(data, &units)
	if err != nil {
		return nil, err
	}
	return units, nil
}
