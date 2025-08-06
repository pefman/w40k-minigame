package game

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// BattleResult holds the result and log of a battle simulation
type BattleResult struct {
	Log    string
	Winner string
	Loser  string
	Draw   bool
}

// SimulateBattleMultipleWeapons simulates a battle where both attacker and defender
// can use multiple weapons. It supports abilities like BLAST, SUSTAINED HITS, DEVASTATING WOUNDS, etc.
func SimulateBattleMultipleWeapons(attacker Unit, attackerWeapons []Weapon, defender Unit, defenderWeapons []Weapon) BattleResult {
	rand.Seed(time.Now().UnixNano())

	wA, _ := strconv.Atoi(attacker.Stats[0].W)
	wD, _ := strconv.Atoi(defender.Stats[0].W)

	log := &strings.Builder{}
	fmt.Fprintf(log, "💥 %s vs %s\n\n", attacker.UnitName, defender.UnitName)

	defIsVehicle := false
	for _, kw := range defender.Keywords {
		for _, w := range kw.Words {
			if strings.EqualFold(w, "Vehicle") {
				defIsVehicle = true
				break
			}
		}
		if defIsVehicle {
			break
		}
	}

	attackWithWeapons := func(weapons []Weapon, defenderWounds *int, defenderTough string, attackerName, defenderName string) {
		for _, weapon := range weapons {
			attacks := parseDice(weapon.Attacks)

			if hasAbility(weapon.Abilities, "HEAVY") && strings.EqualFold(weapon.Range, "melee") {
				if attacks > 1 {
					attacks--
					fmt.Fprintf(log, "  (HEAVY ability active: attacks reduced by 1 to %d)\n", attacks)
				}
			}

			if hasAbility(weapon.Abilities, "BLAST") {
				attacks *= 3
				fmt.Fprintf(log, "  (BLAST ability active: attacks tripled to %d)\n", attacks)
			}

			requiredHit := parseSkill(weapon.Skill)
			hits := 0
			wounds := 0
			damage := 0
			mortalWounds := 0

			fmt.Fprintf(log, "💥 %s attacks %s with %s\n", attackerName, defenderName, weapon.Name)
			fmt.Fprintf(log, "🎲 Rolling %d attacks. Need %d+ to hit\n", attacks, requiredHit)

			sustainedHits := parseSustainedHits(weapon.Abilities)
			rerollOnes := hasAbility(weapon.Abilities, "REROLL ONES")
			mortalWoundsPerHit := parseMortalWounds(weapon.Abilities)
			antiVehicleThreshold := parseAntiVehicleThreshold(weapon.Abilities)
			hasDevastating := hasAbility(weapon.Abilities, "DEVASTATING WOUNDS")

			for i := 0; i < attacks; i++ {
				hitRoll := rollDiceWithReroll(requiredHit, rerollOnes)
				fmt.Fprintf(log, "  🎯 Attack %d rolled a %d", i+1, hitRoll)
				if hitRoll >= requiredHit {
					hits++
					fmt.Fprintf(log, " — HIT!\n")

					isCrit := hitRoll == 6
					if isCrit && hasDevastating {
						mortalWounds++
						fmt.Fprintf(log, "      💥 Critical Hit! Devastating Wounds inflicted (+1 mortal wound)\n")
					}

					if rollWoundVerbose(weapon.Strength, defenderTough, log, i+1) {
						wounds++
						dmg := parseDice(weapon.Damage)

						if defIsVehicle && antiVehicleThreshold > 0 {
							weaponStrength, _ := strconv.Atoi(weapon.Strength)
							if weaponStrength >= antiVehicleThreshold {
								dmg++
								fmt.Fprintf(log, "      (ANTI-VEHICLE active: +1 damage)\n")
							}
						}

						damage += dmg
						fmt.Fprintf(log, "      💥 Dealt %d damage\n", dmg)

						if mortalWoundsPerHit > 0 {
							mortalWounds += mortalWoundsPerHit
							fmt.Fprintf(log, "      💀 Inflicted %d mortal wounds (ignores armor)\n", mortalWoundsPerHit)
						}
					} else {
						fmt.Fprintf(log, "      ❌ Failed to wound\n")
					}

					for sh := 0; sh < sustainedHits; sh++ {
						extraHit := rollDiceWithReroll(requiredHit, rerollOnes)
						fmt.Fprintf(log, "    🔄 Sustained Hit %d rolled a %d", sh+1, extraHit)
						if extraHit >= requiredHit {
							hits++
							fmt.Fprintf(log, " — HIT!\n")

							if rollWoundVerbose(weapon.Strength, defenderTough, log, i+1) {
								wounds++
								dmg := parseDice(weapon.Damage)
								damage += dmg
								fmt.Fprintf(log, "      💥 Dealt %d damage\n", dmg)

								if mortalWoundsPerHit > 0 {
									mortalWounds += mortalWoundsPerHit
									fmt.Fprintf(log, "      💀 Inflicted %d mortal wounds (ignores armor)\n", mortalWoundsPerHit)
								}
							} else {
								fmt.Fprintf(log, "      ❌ Failed to wound\n")
							}
						} else {
							fmt.Fprintf(log, " — MISS\n")
						}
					}
				} else {
					fmt.Fprintf(log, " — MISS\n")
				}
			}

			totalDamage := damage + mortalWounds
			*defenderWounds -= totalDamage
			if *defenderWounds < 0 {
				*defenderWounds = 0
			}

			fmt.Fprintf(log, "\n📊 Summary: Hits: %d | Wounds: %d | Damage: %d (+%d mortal wounds) | %s's Remaining Wounds: %d\n\n",
				hits, wounds, damage, mortalWounds, defenderName, *defenderWounds)
		}
	}

	attackerWounds := wA
	defenderWounds := wD
	defenderTough := defender.Stats[0].T
	attackerTough := attacker.Stats[0].T

	attackWithWeapons(attackerWeapons, &defenderWounds, defenderTough, attacker.UnitName, defender.UnitName)
	attackWithWeapons(defenderWeapons, &attackerWounds, attackerTough, defender.UnitName, attacker.UnitName)

	var winner, loser string
	var draw bool
	if attackerWounds <= 0 && defenderWounds <= 0 {
		draw = true
		fmt.Fprintf(log, "☠️ Both units are slain in mutual destruction!\n")
	} else if attackerWounds <= 0 {
		winner = defender.UnitName
		loser = attacker.UnitName
		fmt.Fprintf(log, "🏁 %s is slain! %s wins!\n", attacker.UnitName, defender.UnitName)
	} else if defenderWounds <= 0 {
		winner = attacker.UnitName
		loser = defender.UnitName
		fmt.Fprintf(log, "🏁 %s is slain! %s wins!\n", defender.UnitName, attacker.UnitName)
	} else {
		fmt.Fprintf(log, "⚔️ Both units survive the exchange.\n")
	}

	return BattleResult{
		Log:    log.String(),
		Winner: winner,
		Loser:  loser,
		Draw:   draw,
	}
}

// Helper functions below...

func hasAbility(abilities []string, ability string) bool {
	for _, ab := range abilities {
		if strings.Contains(strings.ToUpper(ab), ability) {
			return true
		}
	}
	return false
}

func parseSustainedHits(abilities []string) int {
	for _, ab := range abilities {
		if strings.HasPrefix(strings.ToUpper(ab), "SUSTAINED HITS") {
			parts := strings.Fields(ab)
			if len(parts) == 3 {
				n, err := strconv.Atoi(parts[2])
				if err == nil {
					return n
				}
			}
		}
	}
	return 0
}

func parseMortalWounds(abilities []string) int {
	for _, ab := range abilities {
		if strings.HasPrefix(strings.ToUpper(ab), "MORTAL WOUNDS") {
			parts := strings.Fields(ab)
			if len(parts) == 3 {
				n, err := strconv.Atoi(parts[2])
				if err == nil {
					return n
				}
			}
		}
	}
	return 0
}

func parseAntiVehicleThreshold(abilities []string) int {
	for _, ab := range abilities {
		if strings.HasPrefix(strings.ToUpper(ab), "ANTI-VEHICLE") {
			parts := strings.Fields(ab)
			if len(parts) == 2 {
				thresholdStr := strings.TrimSuffix(parts[1], "+")
				n, err := strconv.Atoi(thresholdStr)
				if err == nil {
					return n
				}
			}
		}
	}
	return 0
}

func rollDiceWithReroll(needed int, rerollOnes bool) int {
	roll := rand.Intn(6) + 1
	if rerollOnes && roll == 1 {
		roll = rand.Intn(6) + 1
	}
	return roll
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

func parseDice(s string) int {
	s = strings.ToUpper(strings.ReplaceAll(s, " ", ""))

	re := regexp.MustCompile(`(?i)(\d*)D(\d+)([+\-]\d+)?`)
	matches := re.FindStringSubmatch(s)
	if len(matches) > 0 {
		countStr := matches[1]
		if countStr == "" {
			countStr = "1"
		}
		count, _ := strconv.Atoi(countStr)
		dieSize, _ := strconv.Atoi(matches[2])
		modifier := 0
		if matches[3] != "" {
			modifier, _ = strconv.Atoi(matches[3])
		}

		total := 0
		for i := 0; i < count; i++ {
			total += rand.Intn(dieSize) + 1
		}
		total += modifier
		if total < 1 {
			total = 1
		}
		return total
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		return 1
	}
	return n
}

func parseSkill(s string) int {
	s = strings.Trim(s, "+")
	n, _ := strconv.Atoi(s)
	return n
}
