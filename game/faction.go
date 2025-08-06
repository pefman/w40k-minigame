package game

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func GetFactions(units []Unit) []string {
	unique := map[string]bool{}
	var factions []string
	for _, u := range units {
		if !unique[u.FactionName] {
			unique[u.FactionName] = true
			factions = append(factions, u.FactionName)
		}
	}
	return factions
}

func PromptSelect(prompt string, options []string) string {
	fmt.Println("\n" + prompt)
	for i, opt := range options {
		fmt.Printf("  [%d] %s\n", i+1, opt)
	}
	fmt.Print("Select: ")
	choice := ReadInt()
	if choice < 1 || choice > len(options) {
		fmt.Println("Invalid choice.")
		return PromptSelect(prompt, options)
	}
	return options[choice-1]
}

func PromptUnit(units []Unit, faction string, prompt string) Unit {
	filtered := []Unit{}
	for _, u := range units {
		if u.FactionName == faction {
			filtered = append(filtered, u)
		}
	}
	fmt.Println("\n" + prompt)
	for i, u := range filtered {
		fmt.Printf("  [%d] %s\n", i+1, u.UnitName)
	}
	fmt.Print("Select: ")
	choice := ReadInt()
	if choice < 1 || choice > len(filtered) {
		fmt.Println("Invalid choice.")
		return PromptUnit(units, faction, prompt)
	}
	return filtered[choice-1]
}

func ReadInt() int {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	n, err := strconv.Atoi(text)
	if err != nil {
		return 0
	}
	return n
}
