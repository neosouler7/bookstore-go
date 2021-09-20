package binance

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
	pairs    []string
)

func pongWs() {
	for {
		websocketmanager.Pong(exchange)
		time.Sleep(time.Second * 5)
	}
}

func subscribeWs() {
	time.Sleep(time.Second * 1)
	var streamSlice []string
	for _, pair := range pairs {
		var pairInfo = strings.Split(pair, ":")
		market, symbol := pairInfo[0], pairInfo[1]

		streamSlice = append(streamSlice, fmt.Sprintf("\"%s%s@depth20\"", symbol, market))
	}
	streams := strings.Join(streamSlice, ",")
	msg := fmt.Sprintf("{\"method\": \"SUBSCRIBE\",\"params\": [%s],\"id\": %d}", streams, time.Now().UnixNano()/100000)

	websocketmanager.SendMsg(exchange, msg)
	fmt.Printf(websocketmanager.SubscribeMsg, exchange)
}

func receiveWs() {
	for {
		_, msgBytes, err := websocketmanager.Conn(exchange).ReadMessage()
		tgmanager.HandleErr(exchange, err)

		if strings.Contains(string(msgBytes), "result") {
			fmt.Printf(websocketmanager.FilteredMsg, exchange, string(msgBytes))
		} else {
			var rJson interface{}
			commons.Bytes2Json(msgBytes, &rJson)
			SetOrderbook("W", exchange, rJson.(map[string]interface{}))
		}

	}
}

func rest() {
	c := make(chan map[string]interface{})
	buffer, rateLimit := config.GetRateLimit(exchange)

	for {
		for _, pair := range pairs {
			go restmanager.FastHttpRequest(c, exchange, "GET", pair)
		}

		for i := 0; i < len(pairs); i++ {
			rJson := <-c
			SetOrderbook("R", exchange, rJson)
		}

		// 1번에 (1s / rateLimit)s 만큼 쉬어야 하고, 동시에 pair 만큼 api hit 하니, 그만큼 쉬어야함.
		// ex) 1 / 10 s * 2 = 0.2s => 200ms
		pairsLength := float64(len(pairs)) * buffer
		time.Sleep(time.Millisecond * time.Duration(int(1/rateLimit*pairsLength*10*100)))
	}
}

func Run(e string) {
	exchange, pairs = e, config.GetPairs(exchange)
	var wg sync.WaitGroup

	// pong
	wg.Add(1)
	go pongWs()

	// subscribe websocket stream
	wg.Add(1)
	go func() {
		subscribeWs()
		wg.Done()
	}()

	// receive websocket msg
	wg.Add(1)
	go receiveWs()

	// rest
	wg.Add(1)
	go rest()

	wg.Wait()
}
