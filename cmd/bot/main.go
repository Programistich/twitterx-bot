package main

import (
	"log"

	"twitterx-bot/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
