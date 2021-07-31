package tgmanager

import (
	"errors"
	"fmt"
	"log"
	"neosouler7/bookstore-go/commons"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	errorTgInit = errors.New("[ERROR] error on init telegramBot")
)

func getBot() (*tgbotapi.BotAPI, error) {
	var tgInfo = commons.ReadConfig("Tg").(map[string]interface{})
	bot, err := tgbotapi.NewBotAPI(tgInfo["token"].(string))
	if err != nil {
		log.Fatalln(errorTgInit)
	}
	bot.Debug = true

	return bot, err

	// [when need to get chatId]
	// u := tgbotapi.NewUpdate(0)
	// u.Timeout = 60
	// updateChannel, err := tgBot.GetUpdatesChan(u)
	// if err != nil {
	// 	panic(err)
	// }
	// for update := range updateChannel {
	// 	if update.Message == nil { // ignore any non-Message Updates
	// 		continue
	// 	}
	// 	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	// }
}

func SendTgMessage(tgMsg string) {
	bot, _ := getBot()
	var tgInfo = commons.ReadConfig("Tg").(map[string]interface{})
	chat_ids := tgInfo["chat_ids"].([]interface{})

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	tgMsg = fmt.Sprintf("%s \n%s", tgMsg, currentTime)

	for _, chat_id := range chat_ids {
		msg := tgbotapi.NewMessage(int64(chat_id.(float64)), tgMsg)
		_, err := bot.Send(msg)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
