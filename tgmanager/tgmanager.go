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
	StampMicro    = "Jan _2 15:04:05.000000"
	errGetBot     = errors.New("tg init failed")
	errSendMsg    = errors.New("tg sendMsg failed")
	errGetUpdates = errors.New("tg getUpdated failed")
	t             *tgbotapi.BotAPI
	once          sync.Once
	b             bot
)

type bot struct {
	token    string
	chat_ids []int
	location *time.Location
}

func InitBot(t string, c_ids []int, l *time.Location) {
	b = bot{t, c_ids, l}
}

func (b *bot) Bot() *tgbotapi.BotAPI {
	if t == nil {
		once.Do(func() {
			tgPointer, err := tgbotapi.NewBotAPI(b.token)
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
	now := time.Now().In(b.location).Format(StampMicro)
	tgMsg = fmt.Sprintf("%s \n\n%s", tgMsg, now)

	for _, chat_id := range b.chat_ids {
		msg := tgbotapi.NewMessage(int64(chat_id), tgMsg)
		_, err := b.Bot().Send(msg)
		if err != nil {
			log.Fatalln(errSendMsg)
		}
	}
}

// when need to get chatId
func GetUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updateChannel, err := b.Bot().GetUpdatesChan(u)
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
		tgMsg := fmt.Sprintf("[error]\n%s\n→%s", msg, err.Error())
		SendMsg(tgMsg)
		log.Fatalln(err)
	}
}
