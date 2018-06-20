package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

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

	jenkinsURL := os.Getenv("BOT_JENKINS_URL")
	if jenkinsURL == "" {
		log.Printf("%s: missing env var BOT_JENKINS_URL", me)
		os.Exit(2)
	}
	log.Printf("%s: found env var BOT_JENKINS_URL=%s", me, jenkinsURL)

	jenkinsAuthUser := os.Getenv("BOT_JENKINS_USER")
	if jenkinsAuthUser == "" {
		log.Printf("%s: missing env var BOT_JENKINS_USER", me)
		os.Exit(3)
	}
	log.Printf("%s: found env var BOT_JENKINS_USER=%s", me, jenkinsAuthUser)

	jenkinsAuthPass := os.Getenv("BOT_JENKINS_PASS")
	if jenkinsAuthPass == "" {
		log.Printf("%s: missing env var BOT_JENKINS_PASS", me)
		os.Exit(4)
	}
	log.Printf("%s: found env var BOT_JENKINS_PASS", me)

	jenkinsInputID := os.Getenv("BOT_JENKINS_INPUT_ID")
	if jenkinsInputID == "" {
		log.Printf("%s: missing env var BOT_JENKINS_INPUT_ID", me)
		os.Exit(5)
	}
	log.Printf("%s: found env var BOT_JENKINS_INPUT_ID=%s", me, jenkinsInputID)

	bot, errBot := tgbotapi.NewBotAPI(token)
	if errBot != nil {
		log.Printf("%s: failure creating bot client: %v", me, errBot)
		os.Exit(6)
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
		os.Exit(7)
	}

	// Optional: wait for updates and clear them if you don't want to handle
	// a large backlog of old messages
	log.Printf("%s: clearing backlog of old messages", me)
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	log.Printf("%s: entering service loop", me)

	for update := range updates {

		log.Printf("%s: message=%v callbackQuery=%v", me, update.Message, update.CallbackQuery)

		if update.CallbackQuery != nil {
			log.Printf("handling callback")

			log.Printf("calling jenkins api")
			var action string
			if strings.HasPrefix(update.CallbackQuery.Data, "aprovar") {
				action = "proceedEmpty"
			} else {
				action = "abort"
			}

			var errApprove error

			parameters := strings.Fields(update.CallbackQuery.Data)
			if len(parameters) < 5 {
				errApprove = fmt.Errorf("bad short reponse: [%s]", update.CallbackQuery.Data)
			} else {
				jobName := parameters[2]
				buildID := parameters[4]
				errApprove = buildApprove(jenkinsURL, jenkinsAuthUser, jenkinsAuthPass, jobName, buildID, jenkinsInputID, action)
			}

			log.Printf("answering callback (ack button)")
			text := fmt.Sprintf("resposta: %s", update.CallbackQuery.Data)
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, text)
			bot.AnswerCallbackQuery(callback)

			log.Printf("removing keyboard")
			keyboard := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{})
			edit := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, keyboard)
			bot.Send(edit)

			log.Printf("sending feedback")
			feedback := fmt.Sprintf("%s respondeu: %s", update.CallbackQuery.From.FirstName, update.CallbackQuery.Data)
			if errApprove != nil {
				feedback += fmt.Sprintf("- erro api jenkins: %v", errApprove)
			}
			log.Printf("feedback: %s", feedback)
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, feedback)
			bot.Send(msg)

			continue
		}

		if update.Message == nil {
			continue
		}

		log.Printf("%s: user=%s text=%s", me, update.Message.From.UserName, update.Message.Text)

		// echo back text
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		bot.Send(msg)
	}
}

func buildApprove(jenkins, user, pass, jobName, buildID, inputID, action string) error {

	jenkinsURL := fmt.Sprintf("%s/job/%s/%s/input/%s/%s", jenkins, jobName, buildID, inputID, action)

	v := url.Values{}
	v.Set("name", "name")

	req, errNew := http.NewRequest("POST", jenkinsURL, strings.NewReader(v.Encode()))
	if errNew != nil {
		return errNew
	}

	req.SetBasicAuth(user, pass)

	client := http.Client{}

	resp, errDel := client.Do(req)
	if errDel != nil {
		return errDel
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad http post status: %d", resp.StatusCode)
	}

	_, errRead := ioutil.ReadAll(resp.Body)

	return errRead
}
