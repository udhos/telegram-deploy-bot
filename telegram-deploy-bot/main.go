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
		os.Exit(1)
	}
	log.Printf("%s: found env var BOT_TOKEN", me)

	bot, errBot := tgbotapi.NewBotAPI(token)
	if errBot != nil {
		log.Printf("%s: failure creating bot client: %v", me, errBot)
		os.Exit(2)
	}

	debug := os.Getenv("BOT_DEBUG") != ""
	log.Printf("%s: BOT_DEBUG=%v", me, debug)
	bot.Debug = debug

	log.Printf("%s: bot authorized on account: %s", me, bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, errChan := bot.GetUpdatesChan(u)
	if errChan != nil {
		log.Printf("%s: could not get update channel: %v", me, errChan)
		os.Exit(3)
	}

	log.Printf("%s: entering service loop", me)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("%s: [%s] %s", me, update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}
}
