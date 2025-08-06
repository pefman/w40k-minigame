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
// can use multiple weapons. It supports abilities like BLAST and SUSTAINED HITS.
func SimulateBattleMultipleWeapons(attacker Unit, attackerWeapons []Weapon, defender Unit, defenderWeapons []Weapon) BattleResult {
	rand.Seed(time.Now().UnixNano())

	// Extract wounds from units
	wA, _ := strconv.Atoi(attacker.Stats[0].W)
	wD, _ := strconv.Atoi(defender.Stats[0].W)

	log := &strings.Builder{}
	fmt.Fprintf(log, "💥 %s vs %s\n\n", attacker.UnitName, defender.UnitName)

	attackWithWeapons := func(weapons []Weapon, defenderWounds *int, defenderTough string, attackerName, defenderName string) {
		for _, weapon := range weapons {
			attacks := parseDice(weapon.Attacks)
			requiredHit := parseSkill(weapon.Skill)
			hits := 0
			wounds := 0
			damage := 0

			fmt.Fprintf(log, "💥 %s attacks %s with %s\n", attackerName, defenderName, weapon.Name)
			fmt.Fprintf(log, "🎲 Rolling %d attacks. Need %d+ to hit\n", attacks, requiredHit)

			sustainedHits := parseSustainedHits(weapon.Abilities)
			isBlast := hasBlast(weapon.Abilities)

			effectiveAttacks := attacks
			if isBlast {
				effectiveAttacks *= 3 // example: triple attacks for BLAST
				fmt.Fprintf(log, "  (BLAST ability active: attacks tripled to %d)\n", effectiveAttacks)
			}

			for i := 0; i < effectiveAttacks; i++ {
				hitRoll := rand.Intn(6) + 1
				fmt.Fprintf(log, "  🎯 Attack %d rolled a %d", i+1, hitRoll)
				if hitRoll >= requiredHit {
					hits++
					fmt.Fprintf(log, " — HIT!\n")

					if rollWoundVerbose(weapon.Strength, defenderTough, log, i+1) {
						wounds++
						dmg := parseDice(weapon.Damage)
						damage += dmg
						fmt.Fprintf(log, "      💥 Dealt %d damage\n", dmg)
					} else {
						fmt.Fprintf(log, "      ❌ Failed to wound\n")
					}

					// Handle sustained hits
					for sh := 0; sh < sustainedHits; sh++ {
						extraHit := rand.Intn(6) + 1
						fmt.Fprintf(log, "    🔄 Sustained Hit %d rolled a %d", sh+1, extraHit)
						if extraHit >= requiredHit {
							hits++
							fmt.Fprintf(log, " — HIT!\n")
							if rollWoundVerbose(weapon.Strength, defenderTough, log, i+1) {
								wounds++
								dmg := parseDice(weapon.Damage)
								damage += dmg
								fmt.Fprintf(log, "      💥 Dealt %d damage\n", dmg)
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

			*defenderWounds -= damage
			if *defenderWounds < 0 {
				*defenderWounds = 0
			}

			fmt.Fprintf(log, "\n📊 Summary: Hits: %d | Wounds: %d | Damage: %d | %s's Remaining Wounds: %d\n\n",
				hits, wounds, damage, defenderName, *defenderWounds)
		}
	}

	attackerWounds := wA
	defenderWounds := wD
	defenderTough := defender.Stats[0].T
	attackerTough := attacker.Stats[0].T

	// Attacker attacks defender
	attackWithWeapons(attackerWeapons, &defenderWounds, defenderTough, attacker.UnitName, defender.UnitName)

	// Defender attacks attacker
	attackWithWeapons(defenderWeapons, &attackerWounds, attackerTough, defender.UnitName, attacker.UnitName)

	// Determine winner
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

// Helper to parse SUSTAINED HITS ability
func parseSustainedHits(abilities []string) int {
	for _, ab := range abilities {
		if strings.HasPrefix(ab, "SUSTAINED HITS") {
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

// Helper to check for BLAST ability
func hasBlast(abilities []string) bool {
	for _, ab := range abilities {
		if ab == "BLAST" {
			return true
		}
	}
	return false
}

// rollWoundVerbose prints verbose wound roll info and returns success/failure
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

	// Regex to parse dice expressions like 4D6+2, 2D6, D6+1
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

	// Fallback: try to parse as int
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
