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

func pongWs() {
	for {
		websocketmanager.SendMsg(exchange, pingMsg)
		time.Sleep(time.Second * 5)
	}
}

func subscribeWs(pairs []string) {
	time.Sleep(time.Second * 1)
	var streamSlice []string
	for _, pair := range pairs {
		var pairInfo = strings.Split(pair, ":")
		market, symbol := strings.ToUpper(pairInfo[0]), strings.ToUpper(pairInfo[1])

		streamSlice = append(streamSlice, fmt.Sprintf("'%s-%s'", market, symbol))
	}
	uuid := uuid.NewString()
	streams := strings.Join(streamSlice, ",")
	msg := fmt.Sprintf("[{'ticket':'%s'}, {'type': 'orderbook', 'codes': [%s]}]", uuid, streams)

	websocketmanager.SendMsg(exchange, msg)
	fmt.Printf(websocketmanager.SubscribeMsg, exchange)
}

func receiveWs() {
	for {
		_, msgBytes, err := websocketmanager.Conn(exchange).ReadMessage()
		tgmanager.HandleErr(exchange, err)

		if strings.Contains(string(msgBytes), "status") {
			fmt.Println("PONG") // {"status":"UP"}
		} else {
			var rJson interface{}
			commons.Bytes2Json(msgBytes, &rJson)
			go SetOrderbook("W", exchange, rJson.(map[string]interface{}))
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
			pairsLength := float64(len(pairs)) * buffer
			time.Sleep(time.Millisecond * time.Duration(int(1/rateLimit*pairsLength*10*100)))
		}

		for i := 0; i < len(pairs); i++ {
			rJson := <-c
			go SetOrderbook("R", exchange, rJson)
		}
	}
}

func Run(e string) {
	exchange = e
	pairs := config.GetPairs(exchange)
	var wg sync.WaitGroup

	// ping
	wg.Add(1)
	go pongWs()

	// subscribe websocket stream
	wg.Add(1)
	go func() {
		subscribeWs(pairs)
		wg.Done()
	}()

	// receive websocket msg
	wg.Add(1)
	go receiveWs()

	// rest
	wg.Add(1)
	go rest(pairs)

	wg.Wait()
}
