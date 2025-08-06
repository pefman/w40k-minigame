package game

import (
	"encoding/json"
	"os"
)

type Weapon struct {
	Name     string `json:"name"`
	Range    string `json:"range"`
	Attacks  string `json:"attacks"`
	Skill    string `json:"skill"`
	Strength string `json:"strength"`
	AP       string `json:"ap"`
	Damage   string `json:"damage"`
}

type Stats struct {
	Unit string `json:"unit"`
	T    string `json:"t"`
	W    string `json:"w"`
	SV   string `json:"sv"`
}

type Unit struct {
	FactionName string   `json:"factionname"`
	UnitName    string   `json:"unitname"`
	Stats       []Stats  `json:"stats"`
	Weapons     []Weapon `json:"weapons"`
}

func LoadUnitsFromFile(path string) ([]Unit, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var units []Unit
	if err := json.Unmarshal(data, &units); err != nil {
		return nil, err
	}
	return units, nil
}
