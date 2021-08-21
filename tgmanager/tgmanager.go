package tgmanager

import (
	"errors"
	"fmt"
	"os"
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
	tz := os.Getenv("TZ")
	if tz == "" {
		fmt.Println("tg follows default timezone")
		tz = "Asia/Seoul"
	} else {
		fmt.Println("tg follows SERVER timezone")
	}
	location, _ = time.LoadLocation(tz)
}

func Bot() *tgbotapi.BotAPI {
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
	now := time.Now().In(location).Format(StampMicro)
	tgMsg = fmt.Sprintf("%s \n%s", tgMsg, now)

	for _, chat_id := range chat_ids {
		msg := tgbotapi.NewMessage(int64(chat_id.(float64)), tgMsg)
		_, err := Bot().Send(msg)
		commons.HandleErr(err, errSendMsg)
	}
}

// when need to get chatId
func GetUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updateChannel, err := Bot().GetUpdatesChan(u)
	commons.HandleErr(err, errGetUpdates)

	for update := range updateChannel {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		fmt.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	}
}
