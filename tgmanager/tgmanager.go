package tgmanager

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/neosouler7/bookstore-go/config"
	"github.com/neosouler7/bookstore-go/redismanager"
)

var (
	StampMicro    = "Jan _2 15:04:05.000000"
	errGetBot     = errors.New("tg init failed")
	errSendMsg    = errors.New("tg sendMsg failed")
	errGetUpdates = errors.New("tg getUpdated failed")

	t    *tgbotapi.BotAPI
	once sync.Once
	b    bot
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
	// go processErrors()
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

// func getLastSentFilePath(exchange string) string {
// 	execPath, err := os.Executable()
// 	if err != nil {
// 		log.Println("Failed to get executable path:", err)
// 		return "last_sent.txt" // fallback
// 	}
// 	dir := filepath.Dir(execPath)
// 	return filepath.Join(dir, fmt.Sprintf("last_sent_%s.txt", exchange))
// }

// func readLastSentTime(exchange string) (time.Time, error) {
// 	path := getLastSentFilePath(exchange)
// 	data, err := os.ReadFile(path)
// 	if err != nil {
// 		if os.IsNotExist(err) {
// 			return time.Time{}, nil // 파일 없으면 0으로 간주
// 		}
// 		return time.Time{}, err
// 	}

// 	timestampStr := strings.TrimSpace(string(data))
// 	unixSec, err := strconv.ParseInt(timestampStr, 10, 64)
// 	if err != nil {
// 		return time.Time{}, err
// 	}

// 	return time.Unix(unixSec, 0), nil
// }

// func writeLastSentTime(exchange string, t time.Time) error {
// 	path := getLastSentFilePath(exchange)
// 	return os.WriteFile(path, []byte(fmt.Sprintf("%d", t.Unix())), 0644)
// }

func HandleErr(exchange string, err error) {
	if err == nil {
		return
	}

	lastSent, readErr := redismanager.ReadLastSentTime(exchange)
	if readErr != nil {
		log.Println("Failed to read last sent time from Redis:", readErr)
	}

	if time.Since(lastSent) >= 1*time.Second {
		go SendMsg(fmt.Sprintf("## ERROR %s %s\n%v", config.GetName(), exchange, err))
		if err := redismanager.WriteLastSentTime(exchange, time.Now()); err != nil {
			log.Println("Failed to write last sent time to Redis:", err)
		}
	} else {
		log.Println("Cooldown active, skipping Telegram message.")
	}

	log.Fatalln(err)

	// if err == nil {
	// 	return
	// }

	// now := time.Now()
	// lastSent, readErr := readLastSentTime(exchange)
	// if readErr != nil {
	// 	log.Println("Failed to read last sent time:", readErr) // 메시지 전송 여부 판단이 불확실하므로 보내고 기록하는 걸로 가정
	// }

	// if now.Sub(lastSent) >= 1*time.Second {
	// 	SendMsg(fmt.Sprintf("## ERROR %s %s\n%v", config.GetName(), exchange, err))

	// 	if err := writeLastSentTime(exchange, now); err != nil {
	// 		log.Println("Failed to write last sent time:", err)
	// 	}
	// } else {
	// 	log.Println("Cooldown active, skipping Telegram message.")
	// }

	// log.Fatalln(err)
}
