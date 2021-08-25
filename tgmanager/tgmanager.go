package tgmanager

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	token         string
	chat_ids      []interface{}
	location      *time.Location
	StampMicro    = "Jan _2 15:04:05.000000"
	errGetBot     = errors.New("[ERROR] error on init tgBot")
	errSendMsg    = errors.New("[ERROR] error on sendMsg")
	errGetUpdates = errors.New("[ERROR] error on get updates")
)

var t *tgbotapi.BotAPI
var once sync.Once

func InitBot(t string, c_ids []interface{}, l *time.Location) {
	token = t
	chat_ids = c_ids
	location = l
	Bot()
}

func Bot() *tgbotapi.BotAPI {
	if t == nil {
		once.Do(func() {
			tgPointer, err := tgbotapi.NewBotAPI(token)
			if err != nil {
				log.Fatalln(errGetBot)
			}
			t = tgPointer
			t.Debug = true
		})
	}
	return t
}

func SendMsg(tgMsg string) {
	now := time.Now().In(location).Format(StampMicro)
	tgMsg = fmt.Sprintf("%s \n%s", tgMsg, now)

	for _, chat_id := range chat_ids {
		msg := tgbotapi.NewMessage(int64(chat_id.(float64)), tgMsg)
		_, err := Bot().Send(msg)
		if err != nil {
			log.Fatalln(errSendMsg)
		}
	}
}

// when need to get chatId
func GetUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updateChannel, err := Bot().GetUpdatesChan(u)
	if err != nil {
		log.Fatalln(errGetUpdates)
	}

	for update := range updateChannel {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		fmt.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	}
}

func HandleErr(msg string, err error) {
	if err != nil {
		tgMsg := fmt.Sprintf("[error] %s : %s", msg, err.Error())
		SendMsg(tgMsg)
		log.Fatalln(err)
	}
}
