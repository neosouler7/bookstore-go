package coinone

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
)

var (
	exchange string
	pingMsg  string = "{\"request_type\": \"PING\"}"
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

	for _, pair := range pairs {
		pairInfo := strings.Split(pair, ":")
		market, symbol := strings.ToUpper(pairInfo[0]), strings.ToUpper(pairInfo[1])

		msg := "{\"request_type\": \"SUBSCRIBE\", \"channel\": \"ORDERBOOK\", \"topic\": {\"quote_currency\": \"" + strings.ToUpper(market) + "\", \"target_currency\": \"" + strings.ToUpper(symbol) + "\"}}"
		websocketmanager.SendMsg(exchange, msg)
		fmt.Printf(websocketmanager.SubscribeMsg, exchange)
	}
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
			var rJson map[string]interface{}
			commons.Bytes2Json(msgBytes, &rJson)

			rType := rJson["response_type"].(string)
			switch rType {
			default:
				tgmanager.HandleErr(exchange, fmt.Errorf("unknown message: %s", rType))
			case "ERROR":
				errCode := fmt.Sprintf("%v", rJson["error_code"])
				errMsg := fmt.Sprintf("%v", rJson["message"])
				tgmanager.HandleErr(exchange, fmt.Errorf("code: %s, message: %s", errCode, errMsg))
			case "PONG":
				fmt.Println("con PONG")
			case "CONNECTED":
				fmt.Println("con CONNECTED")
			case "SUBSCRIBED":
				fmt.Println("con SUBSCRIBED")
			case "DATA":
				go SetOrderbook("W", exchange, rJson)
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
