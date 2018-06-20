package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gopkg.in/telegram-bot-api.v4"
)

func main() {

	me := os.Args[0]

	if len(os.Args) != 4 {
		log.Printf("usage: %s job build chat", me)
		os.Exit(1)
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Printf("%s: missing env var BOT_TOKEN", me)
		os.Exit(2)
	}
	log.Printf("%s: found env var BOT_TOKEN", me)

	job := os.Args[1]
	build := os.Args[2]
	chat := os.Args[3]

	chatID, errChat := strconv.ParseInt(chat, 10, 64)
	if errChat != nil {
		log.Printf("%s: bad chat id: %s: %v", me, chat, errChat)
		os.Exit(3)
	}

	bot, errBot := tgbotapi.NewBotAPI(token)
	if errBot != nil {
		log.Printf("%s: failure creating bot client: %v", me, errBot)
		os.Exit(4)
	}

	debug := os.Getenv("BOT_DEBUG") != ""
	log.Printf("%s: BOT_DEBUG=%v", me, debug)
	bot.Debug = debug

	log.Printf("%s: authorized on account: %s", me, bot.Self.UserName)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("aprovar", fmt.Sprintf("aprovar job %s build %s", job, build)),
			tgbotapi.NewInlineKeyboardButtonData("negar", fmt.Sprintf("negar job %s build %s", job, build)),
		},
	)

	text := fmt.Sprintf("Autorizar deploy job=%s build=%s ?", job, build)

	msg := tgbotapi.NewMessage(chatID, text)

	msg.ReplyMarkup = keyboard

	bot.Send(msg)
}
