package coinone

import (
	"errors"
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
	errResponseEncoding = errors.New("[ERROR] response encoding")
	errWSRequest        = errors.New("[ERROR] ws request")
	exchange            string
)

func pongWs() {
	msg := "{\"requestType\": \"PING\"}"
	for {
		websocketmanager.SendMsg(exchange, msg)
		time.Sleep(time.Second * 5)
	}
}

func subscribeWs(pairs interface{}) {
	time.Sleep(time.Second * 1)
	for _, pair := range pairs.([]interface{}) {
		var pairInfo = strings.Split(pair.(string), ":")
		var market = strings.ToUpper(pairInfo[0])
		var symbol = strings.ToUpper(pairInfo[1])

		msg := "{\"requestType\": \"SUBSCRIBE\", \"body\": {\"channel\": \"ORDERBOOK\", \"topic\": {\"priceCurrency\": \"" + strings.ToUpper(market) + "\", \"productCurrency\": \"" + strings.ToUpper(symbol) + "\", \"group\": \"EXPANDED\", \"size\": 30}}}"
		websocketmanager.SendMsg(exchange, msg)
	}
	fmt.Println("CON websocket subscribe msg sent!")
}

func receiveWs() {
	for {
		_, msgBytes, err := websocketmanager.Conn(exchange).ReadMessage()
		tgmanager.HandleErr(exchange, err)

		var data interface{}
		commons.Bytes2Json(msgBytes, &data)

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
			SetOrderbook("W", exchange, rJson)
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
