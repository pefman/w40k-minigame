package main

import (
	"fmt"
	"w40k-minigame/game"
)

func main() {
	fmt.Println("Welcome to W40K Minigame!")
	units, err := game.LoadUnitsFromFile("static/wh40k_10th.json")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Loaded %d units!\n", len(units))
}
