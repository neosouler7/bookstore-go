package korbit

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/restmanager"
	"github.com/neosouler7/bookstore-go/tgmanager"
	"github.com/neosouler7/bookstore-go/websocketmanager"
)

var (
	exchange string
)

// func pingWs(wsConn *websocket.Conn) {
// 	msg := "PING"
// 	for {
// 		err := websocketmanager.SendMsg(wsConn, msg)
// 		if err != nil {
// 			log.Fatalln(err)
// 		}
// 		time.Sleep(time.Second * 5)
// 	}
// }

func subscribeWs(pairs interface{}) {
	time.Sleep(time.Second * 1)
	var streamSlice []string
	for _, pair := range pairs.([]interface{}) {
		var pairInfo = strings.Split(pair.(string), ":")
		var market = strings.ToLower(pairInfo[0])
		var symbol = strings.ToLower(pairInfo[1])

		streamSlice = append(streamSlice, fmt.Sprintf("%s_%s", symbol, market))
	}

	ts := time.Now().UnixNano() / 100000 / 10
	streams := fmt.Sprintf("\"orderbook:%s\"", strings.Join(streamSlice, ","))
	msg := fmt.Sprintf("{\"accessToken\": \"null\", \"timestamp\": \"%d\", \"event\": \"korbit:subscribe\", \"data\": {\"channels\": [%s]}}", ts, streams)

	websocketmanager.SendMsg(exchange, msg)
	fmt.Println("KBT websocket subscribe msg sent!")
}

func receiveWs(pairs interface{}) {
	for {
		_, msgBytes, err := websocketmanager.Conn(exchange).ReadMessage()
		tgmanager.HandleErr(err, websocketmanager.ErrReadMsg)

		if strings.Contains(string(msgBytes), "connected") {
			subscribeWs(pairs) // just once
		} else if strings.Contains(string(msgBytes), "subscribe") {
			continue
		} else if strings.Contains(string(msgBytes), "push-orderbook") {
			var rJson interface{}
			commons.Bytes2Json(msgBytes, &rJson)
			SetOrderbook("W", exchange, rJson.(map[string]interface{}))
		} else {
			log.Fatalln(string(msgBytes))
		}
	}
}

func rest(pairs interface{}) {
	c := make(chan map[string]interface{})
	var rateLimit = commons.ReadConfig("RateLimit").(map[string]interface{})[exchange].(float64)
	var buffer = commons.ReadConfig("RateLimit").(map[string]interface{})["buffer"].(float64)

	for {
		for _, pair := range pairs.([]interface{}) {
			go restmanager.FastHttpRequest(c, exchange, "GET", pair.(string))
		}

		for i := 0; i < len(pairs.([]interface{})); i++ {
			rJson := <-c
			SetOrderbook("R", exchange, rJson)
		}

		// 1번에 (1s / rateLimit)s 만큼 쉬어야 하고, 동시에 pair 만큼 api hit 하니, 그만큼 쉬어야함.
		// ex) 1 / 10 s * 2 = 0.2s => 200ms
		pairsLength := float64(len(pairs.([]interface{}))) * buffer
		time.Sleep(time.Millisecond * time.Duration(int(1/rateLimit*pairsLength*10*100)))
	}
}

func Run(e string) {
	exchange = e
	var pairs = commons.ReadConfig("Pairs").(map[string]interface{})[exchange]

	var wg sync.WaitGroup

	// [ping]
	// wg.Add(1)
	// go pingWs(wsConn)

	// [subscribe websocket stream]
	// send subscribe msg on receiveWs
	// wg.Add(1)
	// go func() {
	// 	subscribeWs(wsConn, pairs)
	// 	wg.Done()
	// }()

	// receive websocket msg
	wg.Add(1)
	go receiveWs(pairs)

	// rest
	wg.Add(1)
	go rest(pairs)

	wg.Wait()
}
