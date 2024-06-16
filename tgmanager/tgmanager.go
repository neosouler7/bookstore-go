package tgmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	StampMicro    = "Jan _2 15:04:05.000000"
	errGetBot     = errors.New("tg init failed")
	errSendMsg    = errors.New("tg sendMsg failed")
	errGetUpdates = errors.New("tg getUpdated failed")

	t           *tgbotapi.BotAPI
	once        sync.Once
	b           bot
	mu          sync.Mutex
	errorFile   = "errors.json"
	minInterval = 3 * time.Second // 최소 메시지 전송 간격
)

type bot struct {
	token    string
	chat_ids []int
	location *time.Location
}

type ErrorLog struct {
	LatestSend int64    `json:"latest_send"`
	Errors     []string `json:"errors"`
}

func InitBot(t string, c_ids []int, l *time.Location) {
	b = bot{t, c_ids, l}
	go processErrors()
}

func (b *bot) Bot() *tgbotapi.BotAPI {
	once.Do(func() {
		if t == nil {
			tgPointer, err := tgbotapi.NewBotAPI(b.token)
			if err != nil {
				log.Fatalln(errGetBot)
			}
			t = tgPointer
			t.Debug = true
		}
	})
	return t
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

// func HandleErr(exchange string, err error) {
// 	if err != nil {
// 		tgMsg := fmt.Sprintf("## ERROR %s %s\n%s", config.GetName(), exchange, err.Error())
// 		SendMsg(tgMsg)
// 		log.Fatalln(err)
// 	}
// }

func HandleErr(exchange string, err error) {
	if err != nil {
		tgMsg := fmt.Sprintf("%s:%s", exchange, err.Error())
		appendError(tgMsg)
		log.Fatalln(err)
	}
}

func appendError(errorMsg string) {
	mu.Lock()
	defer mu.Unlock()

	errorLog := readErrorLog()
	errorLog.Errors = append(errorLog.Errors, errorMsg)
	saveErrorLog(errorLog)
}

func readErrorLog() ErrorLog {
	var errorLog ErrorLog

	data, err := os.ReadFile(errorFile)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrorLog{}
		}
		log.Println("Error reading file:", err)
		return ErrorLog{}
	}

	err = json.Unmarshal(data, &errorLog)
	if err != nil {
		log.Println("Error unmarshaling JSON:", err)
		return ErrorLog{}
	}

	return errorLog
}

func saveErrorLog(errorLog ErrorLog) {
	data, err := json.Marshal(errorLog)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return
	}

	err = os.WriteFile(errorFile, data, 0644)
	if err != nil {
		log.Println("Error writing file:", err)
	}
}

func processErrors() {
	ticker := time.NewTicker(minInterval)
	defer ticker.Stop()

	for {
		<-ticker.C
		mu.Lock()
		errorLog := readErrorLog()
		if len(errorLog.Errors) > 0 {
			if time.Since(time.Unix(0, errorLog.LatestSend)) >= minInterval {
				log.Println("Sending errors to Telegram:", errorLog.Errors)
				sendErrorsToTelegram(errorLog.Errors)
				errorLog.Errors = []string{}
				errorLog.LatestSend = time.Now().UnixNano()
				saveErrorLog(errorLog)
			}
		}
		mu.Unlock()
	}
}

func sendErrorsToTelegram(errors []string) {
	if len(errors) == 0 {
		return
	}

	message := fmt.Sprintf("## ERRORS ##\n\n%s\n\n%s", joinErrors(errors), time.Now().Format(StampMicro))
	for _, chat_id := range b.chat_ids {
		msg := tgbotapi.NewMessage(int64(chat_id), message)
		log.Println("Sending message to Telegram:", message)
		_, err := b.Bot().Send(msg)
		if err != nil {
			log.Println(errSendMsg, err)
		}
	}
}

func joinErrors(errors []string) string {
	result := ""
	for _, err := range errors {
		result += err + "\n"
	}
	return result
}
