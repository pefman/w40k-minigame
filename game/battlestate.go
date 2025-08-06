package game

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type BattleResult struct {
	Log    string
	Winner string
	Loser  string
	Draw   bool
}

func SimulateBattle(attacker, defender Unit, mode string) BattleResult {
	rand.Seed(time.Now().UnixNano())
	weaponA := chooseWeapon(attacker, mode)
	weaponD := chooseWeapon(defender, mode)

	wA, _ := strconv.Atoi(attacker.Stats[0].W)
	wD, _ := strconv.Atoi(defender.Stats[0].W)

	log := &strings.Builder{}
	fmt.Fprintf(log, "💥 %s vs %s (%s)\n\n", attacker.UnitName, defender.UnitName, mode)

	round := 1
	turn := "attacker"

	for wA > 0 && wD > 0 {
		fmt.Fprintf(log, "🔁 Round %d — %s's turn\n", round, strings.Title(turn))

		var atk Unit
		var def Unit
		var atkWpn Weapon
		var defTough string
		var defWounds *int

		if turn == "attacker" {
			atk = attacker
			def = defender
			atkWpn = weaponA
			defTough = def.Stats[0].T
			defWounds = &wD
		} else {
			atk = defender
			def = attacker
			atkWpn = weaponD
			defTough = def.Stats[0].T
			defWounds = &wA
		}

		fmt.Fprintf(log, "💥 %s attacks %s with %s\n", atk.UnitName, def.UnitName, atkWpn.Name)
		attacks := parseDice(atkWpn.Attacks)
		hits := 0
		wounds := 0
		damage := 0

		requiredHit := parseSkill(atkWpn.Skill)
		fmt.Fprintf(log, "🎲 Rolling %d attacks. Need %d+ to hit\n", attacks, requiredHit)

		for i := 0; i < attacks; i++ {
			hitRoll := rand.Intn(6) + 1
			fmt.Fprintf(log, "  🎯 Attack %d rolled a %d", i+1, hitRoll)
			if hitRoll >= requiredHit {
				hits++
				fmt.Fprintf(log, " — HIT!\n")
				if rollWoundVerbose(atkWpn.Strength, defTough, log, i+1) {
					wounds++
					dmg := parseDice(atkWpn.Damage)
					damage += dmg
					fmt.Fprintf(log, "      💥 Dealt %d damage\n", dmg)
				} else {
					fmt.Fprintf(log, "      ❌ Failed to wound\n")
				}
			} else {
				fmt.Fprintf(log, " — MISS\n")
			}
		}

		*defWounds -= damage
		if *defWounds < 0 {
			*defWounds = 0
		}

		fmt.Fprintf(log, "\n📊 Summary: Hits: %d | Wounds: %d | Damage Dealt: %d | %s's Remaining Wounds: %d\n\n",
			hits, wounds, damage, def.UnitName, *defWounds)

		if wA == 0 || wD == 0 {
			break
		}

		if turn == "attacker" {
			turn = "defender"
		} else {
			turn = "attacker"
			round++
		}
	}

	var winner, loser string
	var draw bool

	if wA <= 0 && wD <= 0 {
		draw = true
		fmt.Fprintf(log, "☠️ Both units are slain in mutual destruction!\n")
	} else if wA <= 0 {
		winner = defender.UnitName
		loser = attacker.UnitName
		fmt.Fprintf(log, "🏁 %s is slain! %s wins!\n", attacker.UnitName, defender.UnitName)
	} else {
		winner = attacker.UnitName
		loser = defender.UnitName
		fmt.Fprintf(log, "🏁 %s is slain! %s wins!\n", defender.UnitName, attacker.UnitName)
	}

	return BattleResult{
		Log:    log.String(),
		Winner: winner,
		Loser:  loser,
		Draw:   draw,
	}
}

func chooseWeapon(unit Unit, mode string) Weapon {
	for _, w := range unit.Weapons {
		if (mode == "melee" && w.Range == "Melee") || (mode == "ranged" && w.Range != "Melee") {
			return w
		}
	}
	return Weapon{}
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

func parseSkill(s string) int {
	s = strings.Trim(s, "+")
	n, _ := strconv.Atoi(s)
	return n
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
