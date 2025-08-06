package game

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// BattleResult holds battle summary and log
type BattleResult struct {
	Log    string
	Winner string
	Loser  string
	Draw   bool
}

// SimulateBattleGroups simulates battle between groups of units with multiple weapons
func SimulateBattleGroups(
	attacker Unit, attackerWeapons []Weapon, attackerCount int,
	defender Unit, defenderWeapons []Weapon, defenderCount int,
) BattleResult {
	rand.Seed(time.Now().UnixNano())

	attackerWounds := attackerCount * atoi(attacker.Stats[0].W)
	defenderWounds := defenderCount * atoi(defender.Stats[0].W)

	log := &strings.Builder{}
	fmt.Fprintf(log, "💥 Battle starts: %d x %s vs %d x %s\n\n",
		attackerCount, attacker.UnitName, defenderCount, defender.UnitName)

	defenderTough := defender.Stats[0].T
	attackerTough := attacker.Stats[0].T

	round := 1
	for attackerWounds > 0 && defenderWounds > 0 {
		fmt.Fprintf(log, "🔁 Round %d\n\n", round)

		attackWithWeapons := func(weapons []Weapon, defenderWounds *int, defenderTough string, attackerName, defenderName string, unitCount int) {
			for _, weapon := range weapons {
				if *defenderWounds <= 0 {
					break
				}

				attacks := parseDice(weapon.Attacks) * unitCount

				if hasAbility(weapon.Abilities, "HEAVY") && strings.EqualFold(weapon.Range, "melee") {
					if attacks > unitCount {
						attacks -= unitCount
						fmt.Fprintf(log, "  (HEAVY ability active: attacks reduced by %d to %d)\n", unitCount, attacks)
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

				antiVehicleThreshold, antiVehicleBonus := parseAntiVehicleBonus(weapon.Abilities)

				fmt.Fprintf(log, "💥 %s attacks %s with %s\n", attackerName, defenderName, weapon.Name)
				fmt.Fprintf(log, "🎲 Rolling %d attacks. Need %d+ to hit\n", attacks, requiredHit)

				sustainedHits := parseSustainedHits(weapon.Abilities)
				rerollOnes := hasAbility(weapon.Abilities, "REROLL ONES")
				mortalWoundsPerHit := parseMortalWounds(weapon.Abilities)
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

							diceRoll, modifier := rollDamageDice(weapon.Damage, log)

							totalDamage := diceRoll + modifier

							defIsVehicle := isUnitVehicle(defender)

							if defIsVehicle && antiVehicleThreshold > 0 {
								weaponStrength, _ := strconv.Atoi(weapon.Strength)
								if weaponStrength >= antiVehicleThreshold {
									totalDamage += antiVehicleBonus
									fmt.Fprintf(log, "      (ANTI-VEHICLE active: +%d damage)\n", antiVehicleBonus)
								}
							}

							damage += totalDamage
							fmt.Fprintf(log, "      💥 Total damage dealt: %d (Dice total %d + Modifier %d)\n", totalDamage, diceRoll, modifier)

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

		attackWithWeapons(attackerWeapons, &defenderWounds, defenderTough, attacker.UnitName, defender.UnitName, attackerCount)
		if defenderWounds <= 0 {
			break
		}

		attackWithWeapons(defenderWeapons, &attackerWounds, attackerTough, defender.UnitName, attacker.UnitName, defenderCount)
		if attackerWounds <= 0 {
			break
		}

		round++
		fmt.Fprintf(log, "\n--------------------------\n\n")
	}

	var winner, loser string
	var draw bool
	if attackerWounds <= 0 && defenderWounds <= 0 {
		draw = true
		fmt.Fprintf(log, "☠️ Both sides wiped out in mutual destruction!\n")
	} else if attackerWounds <= 0 {
		winner = defender.UnitName
		loser = attacker.UnitName
		fmt.Fprintf(log, "🏁 %s wiped out! %s wins!\n", attacker.UnitName, defender.UnitName)
	} else if defenderWounds <= 0 {
		winner = attacker.UnitName
		loser = defender.UnitName
		fmt.Fprintf(log, "🏁 %s wiped out! %s wins!\n", defender.UnitName, attacker.UnitName)
	} else {
		fmt.Fprintf(log, "⚔️ Both sides still have units alive after the battle.\n")
	}

	return BattleResult{
		Log:    log.String(),
		Winner: winner,
		Loser:  loser,
		Draw:   draw,
	}
}

// rollDamageDice parses damage string and logs each damage dice roll
func rollDamageDice(damageStr string, log *strings.Builder) (int, int) {
	diceRoll := 0
	modifier := 0

	s := strings.ToUpper(strings.ReplaceAll(damageStr, " ", ""))
	re := regexp.MustCompile(`(?i)(\d*)D(\d+)([+\-]\d+)?`)
	matches := re.FindStringSubmatch(s)

	if len(matches) > 0 {
		countStr := matches[1]
		if countStr == "" {
			countStr = "1"
		}
		count, _ := strconv.Atoi(countStr)
		dieSize, _ := strconv.Atoi(matches[2])
		if matches[3] != "" {
			modifier, _ = strconv.Atoi(matches[3])
		}

		for di := 0; di < count; di++ {
			roll := rand.Intn(dieSize) + 1
			diceRoll += roll
			fmt.Fprintf(log, "      🎲 Damage dice roll %d: %d\n", di+1, roll)
		}
	} else {
		n, err := strconv.Atoi(s)
		if err != nil {
			diceRoll = 1
		} else {
			diceRoll = n
		}
	}

	return diceRoll, modifier
}

// isUnitVehicle checks if unit has "Vehicle" keyword
func isUnitVehicle(u Unit) bool {
	for _, kw := range u.Keywords {
		for _, w := range kw.Words {
			if strings.EqualFold(w, "Vehicle") {
				return true
			}
		}
	}
	return false
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// parseDice parses dice notation like "D6", "2D6+1", or just integer strings
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
	// fallback to int
	n, err := strconv.Atoi(s)
	if err != nil {
		return 1
	}
	return n
}

// hasAbility checks if ability string slice contains the specified ability (case-insensitive)
func hasAbility(abilities []string, ability string) bool {
	ability = strings.ToUpper(ability)
	for _, ab := range abilities {
		if strings.Contains(strings.ToUpper(ab), ability) {
			return true
		}
	}
	return false
}

// parseSkill converts skill string like "2+" into int 2
func parseSkill(s string) int {
	s = strings.Trim(s, "+")
	n, _ := strconv.Atoi(s)
	return n
}

// parseAntiVehicleBonus parses "ANTI-VEHICLE X+" returns threshold and bonus damage (usually X)
func parseAntiVehicleBonus(abilities []string) (threshold int, bonus int) {
	for _, ab := range abilities {
		if strings.HasPrefix(strings.ToUpper(ab), "ANTI-VEHICLE") {
			parts := strings.Fields(ab)
			if len(parts) == 2 {
				thresholdStr := strings.TrimSuffix(parts[1], "+")
				n, err := strconv.Atoi(thresholdStr)
				if err == nil {
					return n, n
				}
			}
		}
	}
	return 0, 0
}

// parseSustainedHits parses "SUSTAINED HITS X"
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

// parseMortalWounds parses "MORTAL WOUNDS X"
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

// rollDiceWithReroll rolls a d6, rerolls 1s if rerollOnes true
func rollDiceWithReroll(needed int, rerollOnes bool) int {
	roll := rand.Intn(6) + 1
	if rerollOnes && roll == 1 {
		roll = rand.Intn(6) + 1
	}
	return roll
}

// rollWoundVerbose rolls wound dice and logs detailed info, returns success/failure
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
