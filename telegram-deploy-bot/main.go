package main

import (
	"log"
	"os"

	"gopkg.in/telegram-bot-api.v4"
)

func main() {

	me := os.Args[0]

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Printf("%s: missing env var BOT_TOKEN", me)
		os.Exit(2)
	}
	log.Printf("%s: found env var BOT_TOKEN", me)

	bot, errBot := tgbotapi.NewBotAPI(token)
	if errBot != nil {
		log.Printf("%s: failure creating bot client: %v", me, errBot)
		os.Exit(4)
	}

	log.Printf("%s: bot authorized on account: %s", me, bot.Self.UserName)
}
