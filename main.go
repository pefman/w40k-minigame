package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"w40k-minigame/game"
)

var units []game.Unit

func main() {
	var err error
	units, err = game.LoadUnitsFromFile("static/wh40k_10th.json")
	if err != nil {
		log.Fatalf("failed to load data: %v", err)
	}

	http.HandleFunc("/factions", handleFactions)
	http.HandleFunc("/units", handleUnits)
	http.HandleFunc("/battle", handleBattle)
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

	// Normalize unit names and sort by wounds
	for i := range filtered {
		filtered[i].UnitName = strings.Title(strings.ToLower(filtered[i].UnitName))
	}
	sort.Slice(filtered, func(i, j int) bool {
		wi, _ := strconv.Atoi(filtered[i].Stats[0].W)
		wj, _ := strconv.Atoi(filtered[j].Stats[0].W)
		return wi > wj
	})

	// Return a simplified preview
	type UnitPreview struct {
		Name      string `json:"name"`
		Wounds    string `json:"wounds"`
		Toughness string `json:"toughness"`
	}
	previews := []UnitPreview{}
	for _, u := range filtered {
		previews = append(previews, UnitPreview{
			Name:      u.UnitName,
			Wounds:    u.Stats[0].W,
			Toughness: u.Stats[0].T,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(previews)
}

func handleBattle(w http.ResponseWriter, r *http.Request) {
	type BattleRequest struct {
		Attacker string `json:"attacker"`
		Defender string `json:"defender"`
		Type     string `json:"type"`
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

	result := game.SimulateBattle(atk, def, req.Type)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"result": result.Log,
	})
}

func normalize(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "\"", "")
	s = strings.ReplaceAll(s, "/", "-")
	return s
}
