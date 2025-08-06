// game/battle.go
package game

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func SimulateBattle(attacker, defender Unit, mode string) string {
	rand.Seed(time.Now().UnixNano())

	var weapon Weapon
	found := false
	for _, w := range attacker.Weapons {
		if (mode == "melee" && w.Range == "Melee") || (mode == "ranged" && w.Range != "Melee") {
			weapon = w
			found = true
			break
		}
	}
	if !found {
		return fmt.Sprintf("%s has no %s weapons!", attacker.UnitName, mode)
	}

	defWounds, _ := strconv.Atoi(defender.Stats[0].W)
	defToughness := defender.Stats[0].T
	battleLog := &strings.Builder{}
	fmt.Fprintf(battleLog, "💥 %s attacks %s with %s (%s)\n\n", attacker.UnitName, defender.UnitName, weapon.Name, mode)

	round := 1
	for defWounds > 0 {
		fmt.Fprintf(battleLog, "🔁 Round %d\n", round)

		attacks := parseDice(weapon.Attacks)
		hits := 0
		wounds := 0
		damage := 0

		requiredHit := parseSkill(weapon.Skill)
		fmt.Fprintf(battleLog, "🎲 Rolling %d attacks. Need %+d to hit based on skill %s\n", attacks, requiredHit, weapon.Skill)

		for i := 0; i < attacks; i++ {
			hitRoll := rand.Intn(6) + 1
			fmt.Fprintf(battleLog, "  🎯 Attack %d rolled a %d", i+1, hitRoll)
			if hitRoll >= requiredHit {
				hits++
				fmt.Fprintf(battleLog, " — HIT!\n")
				if rollWoundVerbose(weapon.Strength, defToughness, battleLog, i+1) {
					wounds++
					dmg := parseDice(weapon.Damage)
					damage += dmg
					fmt.Fprintf(battleLog, "      💥 Dealt %d damage\n", dmg)
				} else {
					fmt.Fprintf(battleLog, "      ❌ Failed to wound\n")
				}
			} else {
				fmt.Fprintf(battleLog, " — MISS\n")
			}
		}

		defWounds -= damage
		if defWounds < 0 {
			defWounds = 0
		}
		fmt.Fprintf(battleLog, "\n📊 Summary: Hits: %d | Wounds: %d | Damage Dealt: %d | %s's Remaining Wounds: %d\n\n", hits, wounds, damage, defender.UnitName, defWounds)

		if defWounds == 0 {
			break
		}
		round++
	}

	fmt.Fprintf(battleLog, "🏁 %s is slain after %d round(s)!\n", defender.UnitName, round)
	return battleLog.String()
}

func rollWoundVerbose(strStr string, tghStr string, log *strings.Builder, index int) bool {
	s, _ := strconv.Atoi(strStr)
	t, _ := strconv.Atoi(tghStr)
	roll := rand.Intn(6) + 1

	var needed int
	switch {
	case s >= t*2:
		needed = 2
	case s > t:
		needed = 3
	case s == t:
		needed = 4
	case s*2 <= t:
		needed = 6
	default:
		needed = 5
	}

	fmt.Fprintf(log, "      🎲 Wound Roll %d: %d vs Toughness %d (need %d+)\n", index, roll, t, needed)
	return roll >= needed
}

func parseSkill(s string) int {
	s = strings.Trim(s, "+")
	n, _ := strconv.Atoi(s)
	return n
}

func parseDice(s string) int {
	s = strings.TrimSpace(s)
	switch s {
	case "D6":
		return rand.Intn(6) + 1
	case "2D6":
		return rand.Intn(6) + 1 + rand.Intn(6) + 1
	default:
		n, err := strconv.Atoi(s)
		if err != nil {
			return 1
		}
		return n
	}
}
