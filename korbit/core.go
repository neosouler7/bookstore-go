package korbit

import (
	"encoding/json"
	"fmt"
	"log"
	"neosouler7/bookstore-go/commons"
	"neosouler7/bookstore-go/restmanager"
	"neosouler7/bookstore-go/websocketmanager"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const latencyAllowed float64 = 12.0 // per 1 second

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

func subscribeWs(wsConn *websocket.Conn, pairs interface{}) {
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

	err := websocketmanager.SendMsg(wsConn, msg)
	fmt.Println("KBT websocket subscribe msg sent!")
	if err != nil {
		log.Fatalln(err)
	}
}

func receiveWs(wsConn *websocket.Conn, exchange string, pairs interface{}) {
	for {
		_, message, err := wsConn.ReadMessage()
		if err != nil {
			log.Fatalln(err)
		}

		if strings.Contains(string(message), "connected") {
			subscribeWs(wsConn, pairs) // just once
		} else if strings.Contains(string(message), "subscribe") {
			continue
		} else if strings.Contains(string(message), "push-orderbook") {
			var rJson interface{}
			err = json.Unmarshal(message, &rJson)
			if err != nil {
				log.Fatalln(err)
			}

			err := SetOrderbook("W", exchange, rJson.(map[string]interface{}))
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			log.Fatalln(string(message))
		}
	}
}

func rest(exchange string, pairs interface{}) {
	c := make(chan map[string]interface{})

	for {
		for _, pair := range pairs.([]interface{}) {
			go restmanager.FastHttpRequest(c, exchange, "GET", pair.(string))
		}

		for i := 0; i < len(pairs.([]interface{})); i++ {
			rJson := <-c

			err := SetOrderbook("R", exchange, rJson)
			if err != nil {
				log.Fatalln(err)
			}
		}

		// 1번에 (1s / LATENCY_ALLOWD) = 0.1s 쉬어야 하고, 동시에 pair 만큼 api hit 하니, 그만큼 쉬어야함.
		// ex) 0.1s * 2 = 0.2s => 200ms
		buffer := 1.0
		pairsLength := float64(len(pairs.([]interface{}))) * buffer
		time.Sleep(time.Millisecond * time.Duration(int(1/latencyAllowed*pairsLength*10*100)))
	}
}

func Run(exchange string) {
	var pairs = commons.ReadConfig("Pairs").(map[string]interface{})[exchange]

	// [get websocket connection]
	wsConn, err := websocketmanager.GetConn(exchange)
	if err != nil {
		log.Fatalln(err)
	}

	// // [execute pipeline]
	// // not sure of using pipeline ...
	// // go upbExecPipeline(exchange)

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

	// [receive websocket msg]
	wg.Add(1)
	go receiveWs(wsConn, exchange, pairs)

	// [rest]
	wg.Add(1)
	go rest(exchange, pairs)

	wg.Wait()
}