package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"w40k-minigame/game"
)

var units []game.Unit
var battleCounter int

func main() {
	var err error
	units, err = game.LoadUnitsFromFile("static/wh40k_10th.json")
	if err != nil {
		log.Fatalf("failed to load data: %v", err)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/", http.StatusFound)
	})

	http.HandleFunc("/factions", handleFactions)
	http.HandleFunc("/units", handleUnits)
	http.HandleFunc("/battle", handleBattle)
	http.HandleFunc("/version", handleVersion)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	port := "8080"
	fmt.Printf("\n🌐 Server listening on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleFactions(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(game.GetFactions(units))
}

func handleUnits(w http.ResponseWriter, r *http.Request) {
	faction := r.URL.Query().Get("faction")
	if faction == "" {
		http.Error(w, "Missing 'faction' query param", http.StatusBadRequest)
		return
	}
	filtered := []game.Unit{}
	for _, u := range units {
		if strings.EqualFold(normalize(u.FactionName), normalize(faction)) {
			filtered = append(filtered, u)
		}
	}

	for i := range filtered {
		filtered[i].UnitName = strings.Title(strings.ToLower(filtered[i].UnitName))
	}

	sort.Slice(filtered, func(i, j int) bool {
		return strings.ToLower(filtered[i].UnitName) < strings.ToLower(filtered[j].UnitName)
	})

	// Define a preview struct that includes full weapon objects
	type UnitPreview struct {
		Name      string        `json:"name"`
		Wounds    string        `json:"wounds"`
		Toughness string        `json:"toughness"`
		Weapons   []game.Weapon `json:"weapons"`
	}

	previews := []UnitPreview{}
	for _, u := range filtered {
		previews = append(previews, UnitPreview{
			Name:      u.UnitName,
			Wounds:    u.Stats[0].W,
			Toughness: u.Stats[0].T,
			Weapons:   u.Weapons, // Return full weapons here
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(previews)
}

func handleBattle(w http.ResponseWriter, r *http.Request) {
	type BattleRequest struct {
		Attacker        string   `json:"attacker"`
		AttackerWeapons []string `json:"attackerWeapons"`
		Defender        string   `json:"defender"`
		DefenderWeapons []string `json:"defenderWeapons"`
	}
	var req BattleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var atk, def game.Unit
	for _, u := range units {
		if strings.EqualFold(u.UnitName, req.Attacker) {
			atk = u
		}
		if strings.EqualFold(u.UnitName, req.Defender) {
			def = u
		}
	}
	if atk.UnitName == "" || def.UnitName == "" {
		http.Error(w, "Attacker or Defender not found", http.StatusBadRequest)
		return
	}

	// Find attacker's chosen weapons
	atkWeapons := []game.Weapon{}
	for _, wname := range req.AttackerWeapons {
		for _, w := range atk.Weapons {
			if strings.EqualFold(w.Name, wname) {
				atkWeapons = append(atkWeapons, w)
				break
			}
		}
	}

	// Find defender's chosen weapons
	defWeapons := []game.Weapon{}
	for _, wname := range req.DefenderWeapons {
		for _, w := range def.Weapons {
			if strings.EqualFold(w.Name, wname) {
				defWeapons = append(defWeapons, w)
				break
			}
		}
	}

	result := game.SimulateBattleMultipleWeapons(atk, atkWeapons, def, defWeapons)
	battleCounter++

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"result": result.Log,
	})
}

func handleVersion(w http.ResponseWriter, r *http.Request) {
	version := "dev" // or set via env or ldflags
	json.NewEncoder(w).Encode(map[string]string{"version": version})
}

func normalize(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "\"", "")
	s = strings.ReplaceAll(s, "/", "-")
	return s
}
