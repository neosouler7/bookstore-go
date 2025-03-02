package korbit

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
)

func subscribeWs(pairs []string) {
	time.Sleep(time.Second * 1)
	var streamSlice []string
	for _, pair := range pairs {
		var pairInfo = strings.Split(pair, ":")
		market, symbol := strings.ToLower(pairInfo[0]), strings.ToLower(pairInfo[1])

		streamSlice = append(streamSlice, fmt.Sprintf("%s_%s", symbol, market))
	}

	ts := time.Now().UnixNano() / 100000 / 10
	streams := fmt.Sprintf("\"orderbook:%s\"", strings.Join(streamSlice, ","))
	msg := fmt.Sprintf("{\"accessToken\": \"null\", \"timestamp\": \"%d\", \"event\": \"korbit:subscribe\", \"data\": {\"channels\": [%s]}}", ts, streams)

	websocketmanager.SendMsg(exchange, msg)
	fmt.Printf(websocketmanager.SubscribeMsg, exchange)
}

func receiveWs(pairs []string, done <-chan struct{}, msgQueue chan<- []byte) {
	for {
		select {
		case <-done:
			return
		default:
			_, msgBytes, err := websocketmanager.Conn(exchange).ReadMessage()
			tgmanager.HandleErr(exchange, err)

			if strings.Contains(string(msgBytes), "connected") {
				subscribeWs(pairs) // just once
			} else if strings.Contains(string(msgBytes), "subscribe") {
				continue
			} else if strings.Contains(string(msgBytes), "push-orderbook") {
				msgQueue <- msgBytes
			} else {
				tgmanager.HandleErr(exchange, websocketmanager.ErrReadMsg)
			}
		}
	}
}

func processWsMessages(done <-chan struct{}, msgQueue <-chan []byte) {
	for {
		select {
		case <-done:
			return
		case msgBytes := <-msgQueue:
			var rJson interface{}
			commons.Bytes2Json(msgBytes, &rJson)
			go SetOrderbook("W", exchange, rJson.(map[string]interface{}))
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
				time.Sleep(time.Millisecond * 500)
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

	// receive websocket msg
	wg.Add(1)
	go func() {
		defer wg.Done()
		receiveWs(pairs, done, wsQueue)
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
}
