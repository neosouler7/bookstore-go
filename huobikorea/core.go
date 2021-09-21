package huobikorea

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
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

func pongWs(msg string) {
	websocketmanager.SendMsg(exchange, msg)
	fmt.Printf("%s PONG %s\n", exchange, msg)
}

func subscribeWs(pairs []string) {
	time.Sleep(time.Second * 1)
	for _, pair := range pairs {
		var pairInfo = strings.Split(pair, ":")
		market, symbol := strings.ToLower(pairInfo[0]), strings.ToLower(pairInfo[1])

		msg := fmt.Sprintf("{\"sub\": \"market.%s%s.depth.step0\"}", symbol, market)
		websocketmanager.SendMsg(exchange, msg)
		fmt.Printf(websocketmanager.SubscribeMsg, exchange)
	}

}

func receiveWs() {
	for {
		_, msgBytes, err := websocketmanager.Conn(exchange).ReadMessage()
		tgmanager.HandleErr(exchange, err)

		gzip, err := gzip.NewReader(bytes.NewReader(msgBytes))
		tgmanager.HandleErr(exchange, err)

		gzipMsg := new(strings.Builder)
		_, err = io.Copy(gzipMsg, gzip)
		tgmanager.HandleErr(exchange, err)

		if strings.Contains(gzipMsg.String(), "subbed") {
			continue
		} else if strings.Contains(gzipMsg.String(), "ping") {
			pongWs(strings.Replace(gzipMsg.String(), "ping", "pong", -1))
		} else if strings.Contains(gzipMsg.String(), "tick") {
			var rJson interface{}
			commons.Bytes2Json([]byte(gzipMsg.String()), &rJson)

			SetOrderbook("W", exchange, rJson.(map[string]interface{}))
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
	exchange = e
	pairs := config.GetPairs(exchange)
	var wg sync.WaitGroup

	// [subscribe websocket stream]
	wg.Add(1)
	go func() {
		subscribeWs(pairs)
		wg.Done()
	}()

	// [receive websocket msg]
	wg.Add(1)
	go receiveWs()

	// [rest]
	wg.Add(1)
	go rest(pairs)

	wg.Wait()
}
