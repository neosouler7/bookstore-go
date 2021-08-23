package tgmanager

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/neosouler7/bookstore-go/commons"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	token         = commons.ReadConfig("Tg").(map[string]interface{})["token"].(string)
	chat_ids      = commons.ReadConfig("Tg").(map[string]interface{})["chat_ids"].([]interface{})
	location      *time.Location
	StampMicro    = "Jan _2 15:04:05.000000"
	errGetBot     = errors.New("[ERROR] error on init tgBot")
	errSendMsg    = errors.New("[ERROR] error on sendMsg")
	errGetUpdates = errors.New("[ERROR] error on get updates")
)

var t *tgbotapi.BotAPI
var once sync.Once

func init() {
	location = commons.SetTimeZone("tg")
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

// get err as returned error, and log with specific errMsg
func HandleErr(err error, errMsg error) {
	if err != nil {
		SendMsg(errMsg.Error())
		log.Fatalln(errMsg)
	}
}
