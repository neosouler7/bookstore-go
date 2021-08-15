package tgmanager

import (
	"errors"
	"fmt"
	"neosouler7/bookstore-go/commons"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	token         = commons.ReadConfig("Tg").(map[string]interface{})["token"].(string)
	chat_ids      = commons.ReadConfig("Tg").(map[string]interface{})["chat_ids"].([]interface{})
	errGetBot     = errors.New("[ERROR] error on init tgBot")
	errSendMsg    = errors.New("[ERROR] error on sendMsg")
	errGetUpdates = errors.New("[ERROR] error on get updates")
)

var t *tgbotapi.BotAPI
var once sync.Once

func tgBot() *tgbotapi.BotAPI {
	if t == nil {
		once.Do(func() {
			tgPointer, err := tgbotapi.NewBotAPI(token)
			commons.HandleErr(err, errGetBot)
			t = tgPointer
			t.Debug = true
		})
	}
	return t
}

func SendMsg(tgMsg string) {
	currentTime := time.Now().Format("2006-01-02 15:04:05.123456")
	tgMsg = fmt.Sprintf("%s \n%s", tgMsg, currentTime)

	for _, chat_id := range chat_ids {
		msg := tgbotapi.NewMessage(int64(chat_id.(float64)), tgMsg)
		_, err := tgBot().Send(msg)
		commons.HandleErr(err, errSendMsg)
	}
}

// when need to get chatId
func GetUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updateChannel, err := tgBot().GetUpdatesChan(u)
	commons.HandleErr(err, errGetUpdates)

	for update := range updateChannel {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		fmt.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	}
}
