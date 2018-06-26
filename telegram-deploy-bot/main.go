package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gopkg.in/telegram-bot-api.v4"
)

const version = "0.0"

var authorizedIDTable = map[int]struct{}{}

func main() {

	me := os.Args[0]

	log.Printf("%s: telegram bot version=%s runtime=%s GOMAXPROCS=%d", me, version, runtime.Version(), runtime.GOMAXPROCS(0))

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

	authorizedUserIDList := os.Getenv("BOT_AUTHORIZED_USER_ID_LIST")
	if authorizedUserIDList == "" {
		log.Printf("%s: missing env var BOT_AUTHORIZED_USER_ID_LIST", me)
		os.Exit(6)
	}
	log.Printf("%s: found env var BOT_AUTHORIZED_USER_ID_LIST=%s", me, authorizedUserIDList)

	for _, id := range strings.Split(authorizedUserIDList, ",") {
		value, err := strconv.Atoi(id)
		if err != nil {
			log.Printf("bad authorized user ID: [%s]: %v", id, err)
			os.Exit(7)
		}
		authorizedIDTable[value] = struct{}{}
	}

	bot, errBot := tgbotapi.NewBotAPI(token)
	if errBot != nil {
		log.Printf("%s: failure creating bot client: %v", me, errBot)
		os.Exit(8)
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
		os.Exit(9)
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
			log.Printf("handling callback: update=%v", update)
			log.Printf("handling callback: CallbackQuery=%v", *update.CallbackQuery)

			if !authorizedApprover(update.CallbackQuery.From.ID) {
				feedback := fmt.Sprintf("%s(id=%d) não tem permissão para autorizar", update.CallbackQuery.From.FirstName, update.CallbackQuery.From.ID)
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, feedback)
				bot.Send(msg)
				continue
			}

			log.Printf("calling jenkins api")
			var action string
			if strings.HasPrefix(update.CallbackQuery.Data, "aprovar") {
				action = "proceedEmpty"
			} else {
				action = "abort"
			}

			var strApprove string
			var errApprove error

			parameters := strings.Fields(update.CallbackQuery.Data)
			if len(parameters) < 7 {
				errApprove = fmt.Errorf("bad short jenkins response: [%s]", update.CallbackQuery.Data)
			} else {
				jobName := parameters[2]
				buildID := parameters[4]
				jenkinsInputID := parameters[6]
				strApprove, errApprove = buildApprove(jenkinsURL, jenkinsAuthUser, jenkinsAuthPass, jobName, buildID, jenkinsInputID, action)
				log.Printf("jenkins response: %v - %s", errApprove, strApprove)
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
				feedback += fmt.Sprintf(" (erro api jenkins: %v - %s)", errApprove, strApprove)
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

func authorizedApprover(ID int) bool {
	_, found := authorizedIDTable[ID]
	return found
}

func buildApprove(jenkins, user, pass, jobName, buildID, inputID, action string) (string, error) {

	jenkinsURL := fmt.Sprintf("%s/job/%s/%s/input/%s/%s", jenkins, jobName, buildID, inputID, action)

	log.Printf("jenkins: %s", jenkinsURL)

	v := url.Values{}
	v.Set("name", "name")

	req, errNew := http.NewRequest("POST", jenkinsURL, strings.NewReader(v.Encode()))
	if errNew != nil {
		return "", errNew
	}

	req.SetBasicAuth(user, pass)

	client := http.Client{}

	resp, errDel := client.Do(req)
	if errDel != nil {
		return "", errDel
	}

	defer resp.Body.Close()

	buf, errRead := ioutil.ReadAll(resp.Body)

	strBuf := string(buf)

	if resp.StatusCode != 200 {
		return strBuf, fmt.Errorf("bad http post status: %d", resp.StatusCode)
	}

	return strBuf, errRead
}
