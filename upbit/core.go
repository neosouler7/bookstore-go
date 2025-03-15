package upbit

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/config"
	"github.com/neosouler7/bookstore-go/restmanager"
	"github.com/neosouler7/bookstore-go/tgmanager"
	"github.com/neosouler7/bookstore-go/websocketmanager"

	"github.com/google/uuid"
)

var (
	exchange string
	pingMsg  string = "PING"
)

func pongWs(done <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			websocketmanager.SendMsg(exchange, pingMsg)
		case <-done:
			return
		}
	}
}

func subscribeWs(pairs []string, wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(time.Second * 1)

	var streamSlice []string
	for _, pair := range pairs {
		pairInfo := strings.Split(pair, ":")
		market, symbol := strings.ToUpper(pairInfo[0]), strings.ToUpper(pairInfo[1])
		streamSlice = append(streamSlice, fmt.Sprintf("'%s-%s'", market, symbol))
	}

	uuid := uuid.NewString()
	streams := strings.Join(streamSlice, ",")
	msg := fmt.Sprintf("[{'ticket':'%s'}, {'type': 'orderbook', 'codes': [%s]}]", uuid, streams)

	websocketmanager.SendMsg(exchange, msg)
	fmt.Printf(websocketmanager.SubscribeMsg, exchange)
}

func receiveWs(done <-chan struct{}, msgQueue chan<- []byte) {
	for {
		select {
		case <-done:
			return
		default:
			_, msgBytes, err := websocketmanager.Conn(exchange).ReadMessage()
			if err != nil {
				tgmanager.HandleErr(exchange, err)
			}
			msgQueue <- msgBytes
		}
	}
}

func processWsMessages(done <-chan struct{}, msgQueue <-chan []byte) {
	for {
		select {
		case <-done:
			return
		case msgBytes := <-msgQueue:
			if strings.Contains(string(msgBytes), "status") {
				fmt.Println("PONG") // {"status":"UP"}
			} else {
				var rJson interface{}
				commons.Bytes2Json(msgBytes, &rJson)
				go SetOrderbook("W", exchange, rJson.(map[string]interface{}))
			}
		}
	}
}

func rest(pairs []string, done <-chan struct{}, restQueue chan<- map[string]interface{}) {
	// buffer, rateLimit := config.GetRateLimit(exchange)
	for {
		select {
		case <-done:
			return
		default:
			for _, pair := range pairs {
				go func(pair string) {
					rJson := restmanager.FastHttpRequest2(exchange, "GET", pair)
					restQueue <- rJson
				}(pair)
				// time.Sleep(time.Millisecond * time.Duration(int(1/rateLimit*10*100*buffer)))
				time.Sleep(time.Millisecond * 200)
			}
		}
	}
}

func processRestResponses(done <-chan struct{}, restQueue <-chan map[string]interface{}) {
	for {
		select {
		case <-done:
			return
		case rJson := <-restQueue:
			go SetOrderbook("R", exchange, rJson)
		}
	}
}

func Run(e string) {
	exchange = e
	pairs := config.GetPairs(exchange)
	var wg sync.WaitGroup
	done := make(chan struct{})
	wsQueue := make(chan []byte, 1)                            // WebSocket 메시지 큐
	restQueue := make(chan map[string]interface{}, len(pairs)) // REST 응답 큐

	// ping
	wg.Add(1)
	go func() {
		defer wg.Done()
		pongWs(done)
	}()

	// subscribe websocket stream
	wg.Add(1)
	go subscribeWs(pairs, &wg)

	// receive websocket msg
	wg.Add(1)
	go func() {
		defer wg.Done()
		receiveWs(done, wsQueue)
	}()

	// process websocket messages
	wg.Add(1)
	go func() {
		defer wg.Done()
		processWsMessages(done, wsQueue)
	}()

	// rest
	wg.Add(1)
	go func() {
		defer wg.Done()
		rest(pairs, done, restQueue)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		processRestResponses(done, restQueue)
	}()

	wg.Wait()
	close(done)
}
