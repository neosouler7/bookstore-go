package coinone

import (
	"errors"
	"fmt"
	"log"
	"os"
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
	errWSRequest = errors.New("websocket request")
	exchange     string
	pingMsg      string = "{\"requestType\": \"PING\"}"
)

func pongWs() {
	for {
		websocketmanager.SendMsg(exchange, pingMsg)
		time.Sleep(time.Second * 5)
	}
}

func subscribeWs(pairs []string) {
	time.Sleep(time.Second * 1)
	for _, pair := range pairs {
		var pairInfo = strings.Split(pair, ":")
		var market = strings.ToUpper(pairInfo[0])
		var symbol = strings.ToUpper(pairInfo[1])

		msg := "{\"requestType\": \"SUBSCRIBE\", \"body\": {\"channel\": \"ORDERBOOK\", \"topic\": {\"priceCurrency\": \"" + strings.ToUpper(market) + "\", \"productCurrency\": \"" + strings.ToUpper(symbol) + "\", \"group\": \"EXPANDED\", \"size\": 30}}}"
		websocketmanager.SendMsg(exchange, msg)
	}
	fmt.Printf(websocketmanager.SubscribeMsg, exchange)
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
			os.Exit(0)
		case "ERROR":
			responseErrCode := rJson["errorCode"]
			responseErrMsg := rJson["message"]
			tgmanager.HandleErr(exchange, errWSRequest)

			fmt.Printf("coinone ws ERROR: %f %s\n", responseErrCode, responseErrMsg)
			log.Fatalln(errWSRequest)
		case "PONG":
			fmt.Println("coinone ws PONG")
		case "CONNECTED":
			fmt.Println("coinone ws CONNECTED")
		case "SUBSCRIBED":
			fmt.Println("coinone ws SUBSCRIBED")
		case "DATA":
			go SetOrderbook("W", exchange, rJson)
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
