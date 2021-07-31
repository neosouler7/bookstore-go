package coinone

import (
	"encoding/json"
	"errors"
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

const LATENCY_ALLOWED float64 = 5.0 // per 1 second

var (
	errResponseEncoding = errors.New("[ERROR] response encoding")
	errWSRequest        = errors.New("[ERROR] ws request")
)

// set redis pipeline global
// var pipe = redisManager.GetRedisPipeline()

// func upbExecPipeline(exchange string) {
// 	for {
// 		redisManager.ExecPipeline(exchange, &pipe)
// 		time.Sleep(time.Second * 1)
// 		fmt.Println("UPB execute pipeline")
// 	}
// }

func conPingWs(wsConn *websocket.Conn) {
	msg := "{\"requestType\": \"PING\"}"
	for {
		err := websocketmanager.SendWsMsg(wsConn, msg)
		if err != nil {
			log.Fatalln(err)
		}
		time.Sleep(time.Second * 5)
	}
}

func conSubscribeWs(wsConn *websocket.Conn, pairs interface{}) {
	time.Sleep(time.Second * 1)
	for _, pair := range pairs.([]string) {
		var pairInfo = strings.Split(pair, ":")
		var market = strings.ToUpper(pairInfo[0])
		var symbol = strings.ToUpper(pairInfo[1])

		msg := "{\"requestType\": \"SUBSCRIBE\", \"body\": {\"channel\": \"ORDERBOOK\", \"topic\": {\"priceCurrency\": \"" + strings.ToUpper(market) + "\", \"productCurrency\": \"" + strings.ToUpper(symbol) + "\", \"group\": \"EXPANDED\", \"size\": 30}}}"
		err := websocketmanager.SendWsMsg(wsConn, msg)
		if err != nil {
			log.Fatalln(err)
		}
	}
	log.Println("CON subscribe msg sent!")
}

func conReceiveWs(wsConn *websocket.Conn, exchange string) {
	for {
		_, message, err := wsConn.ReadMessage()
		if err != nil {
			log.Fatalln(err)
		}

		var data interface{}
		err = json.Unmarshal(message, &data)
		if err != nil {
			log.Fatalln(errResponseEncoding)
		}

		rJson := data.(map[string]interface{})
		responseType := rJson["responseType"]
		switch responseType {
		default:
			fmt.Printf("coinone unknown %s\n", responseType)
			log.Fatalln(errWSRequest)
		case "ERROR":
			responseErrCode := rJson["errorCode"]
			responseErrMsg := rJson["message"]
			fmt.Printf("coinone ws ERROR: %f %s\n", responseErrCode, responseErrMsg)
			log.Fatalln(errWSRequest)
		case "PONG":
			fmt.Println("coinone ws PONG")
		case "CONNECTED":
			fmt.Println("coinone ws CONNECTED")
		case "SUBSCRIBED":
			fmt.Println("coinone ws SUBSCRIBED")
		case "DATA":
			err := ConSetOrderbook("W", exchange, rJson)
			if err != nil {
				log.Fatalln(errSetOrderbook)
			}
		}
	}
}

func conRest(exchange string, pairs interface{}) {
	c := make(chan map[string]interface{})

	for {
		for _, pair := range pairs.([]string) {
			go restmanager.FastHttpRequest(c, exchange, "GET", pair)
		}

		for i := 0; i < len(pairs.([]string)); i++ {
			rJson := <-c

			err := ConSetOrderbook("R", exchange, rJson)
			if err != nil {
				log.Fatalln(err)
			}
		}

		// 동시에 pair 만큼 api hit 하니, 그만큼 쉬어야함
		// 1번에 (1s / LATENCY_ALLOWD) = 0.1s 쉬어야 하고, 동시에 pair 만큼 api hit 하니, 그만큼 쉬어야함.
		// ex) 0.1s * 2 = 0.2s => 200ms
		buffer := 1.0
		pairsLength := float64(len(pairs.([]string))) * buffer
		time.Sleep(time.Millisecond * time.Duration(int(1/LATENCY_ALLOWED*pairsLength*10*100)))
	}
}

func Run(exchange string) {
	var pairs = commons.ReadConfig("Pairs")

	// [get websocket connection]
	wsConn, err := websocketmanager.GetWsConn(exchange)
	if err != nil {
		log.Fatalln(err)
	}

	// [execute pipeline]
	// not sure of using pipeline ...
	// go upbExecPipeline(exchange)

	var wg sync.WaitGroup

	// [ping]
	wg.Add(1)
	go conPingWs(wsConn)

	// [subscribe websocket stream]
	wg.Add(1)
	go func() {
		conSubscribeWs(wsConn, pairs)
		wg.Done()
	}()

	// [receive websocket msg]
	wg.Add(1)
	go conReceiveWs(wsConn, exchange)

	// [rest]
	wg.Add(1)
	go conRest(exchange, pairs)

	wg.Wait()
}
