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

func receiveWs(pairs []string) {
	for {
		_, msgBytes, err := websocketmanager.Conn(exchange).ReadMessage()
		tgmanager.HandleErr(exchange, err)

		if strings.Contains(string(msgBytes), "connected") {
			subscribeWs(pairs) // just once
		} else if strings.Contains(string(msgBytes), "subscribe") {
			continue
		} else if strings.Contains(string(msgBytes), "push-orderbook") {
			var rJson interface{}
			commons.Bytes2Json(msgBytes, &rJson)
			go SetOrderbook("W", exchange, rJson.(map[string]interface{}))
		} else {
			tgmanager.HandleErr(exchange, websocketmanager.ErrReadMsg)
		}
	}
}

func rest(pairs []string) {
	c := make(chan map[string]interface{}, len(pairs)) // make buffered
	buffer, rateLimit := config.GetRateLimit(exchange)

	for {
		for _, pair := range pairs {
			go restmanager.FastHttpRequest(c, exchange, "GET", pair)

			// to avoid 429
			time.Sleep(time.Millisecond * time.Duration(int(1/rateLimit*10*100*buffer)))

			rJson := <-c
			go SetOrderbook("R", exchange, rJson)
		}
	}
}

func Run(e string) {
	exchange = e
	pairs := config.GetPairs(exchange)
	var wg sync.WaitGroup

	// receive websocket msg
	wg.Add(1)
	go receiveWs(pairs)

	// rest
	wg.Add(1)
	go rest(pairs)

	wg.Wait()
}
