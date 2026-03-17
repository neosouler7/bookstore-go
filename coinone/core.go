package coinone

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/exchange"
	"github.com/neosouler7/bookstore-go/tgmanager"
	"github.com/neosouler7/bookstore-go/websocketmanager"
)

const (
	name    = "con"
	pingMsg = "{\"request_type\": \"PING\"}"
)

type Coinone struct{}

func (c *Coinone) Ping() {
	websocketmanager.SendMsg(name, pingMsg)
}

func (c *Coinone) Subscribe(pairs []string, wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(time.Second * 1)

	for _, pair := range pairs {
		pairInfo := strings.Split(pair, ":")
		market, symbol := strings.ToUpper(pairInfo[0]), strings.ToUpper(pairInfo[1])
		msg := "{\"request_type\": \"SUBSCRIBE\", \"channel\": \"ORDERBOOK\", \"topic\": {\"quote_currency\": \"" + market + "\", \"target_currency\": \"" + symbol + "\"}}"
		websocketmanager.SendMsg(name, msg)
		fmt.Printf(websocketmanager.SubscribeMsg, name)
	}
}

func (c *Coinone) HandleWsMessage(msgBytes []byte) {
	var rJson map[string]interface{}
	commons.Bytes2Json(msgBytes, &rJson)

	rType := rJson["response_type"].(string)
	switch rType {
	case "PONG":
		fmt.Println("con PONG")
	case "CONNECTED":
		fmt.Println("con CONNECTED")
	case "SUBSCRIBED":
		fmt.Println("con SUBSCRIBED")
	case "DATA":
		SetOrderbook("W", name, rJson)
	case "ERROR":
		errCode := fmt.Sprintf("%v", rJson["error_code"])
		errMsg := fmt.Sprintf("%v", rJson["message"])
		tgmanager.HandleErr(name, fmt.Errorf("code: %s, message: %s", errCode, errMsg))
	default:
		tgmanager.HandleErr(name, fmt.Errorf("unknown message: %s", rType))
	}
}

func (c *Coinone) HandleRestResponse(rJson map[string]interface{}) {
	SetOrderbook("R", name, rJson)
}

func Run(e string) {
	exchange.Run(e, &Coinone{})
}
