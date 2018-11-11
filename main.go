package main

import (
	"fmt"
	"time"
)

var currentGames map[string]string = map[string]string{}

func main() {

	bot := InitBot(false)

	fmt.Println("Bot ready")

	for {
		fmt.Println("Checking for new games")
		games := FetchGames()
		for _, game := range games {
			if currentGames[game.ID] != game.ChangeDate {
				currentGames[game.ID] = game.ChangeDate
				time.Sleep(5 * time.Second)
				fmt.Println("Sending message to user about", game.Title)

				PostMessage(bot, game)
			}
		}

		go cleanupCurrentGames()
		time.Sleep(1 * time.Hour)
	}
}

func cleanupCurrentGames() {
	layout := "2006-01-02T15:04:05.000Z"
	cutoff := time.Now().AddDate(0, 0, -7)
	for _, ID := range currentGames {
		t, _ := time.Parse(layout, currentGames[ID])
		if t.Before(cutoff) {
			delete(currentGames, ID)
		}
	}
}
