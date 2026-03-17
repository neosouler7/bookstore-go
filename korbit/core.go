package korbit

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

const name = "kbt"

type Korbit struct{}

func (k *Korbit) Ping() {
	websocketmanager.Ping(name)
}

func (k *Korbit) Subscribe(pairs []string, wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(time.Second * 1)

	var streamSlice []string
	for _, pair := range pairs {
		var pairInfo = strings.Split(pair, ":")
		market, symbol := strings.ToLower(pairInfo[0]), strings.ToLower(pairInfo[1])
		streamSlice = append(streamSlice, fmt.Sprintf("\"%s_%s\"", symbol, market))
	}
	msg := fmt.Sprintf("[{\"method\": \"subscribe\", \"type\": \"orderbook\", \"symbols\": [%s]}]", strings.Join(streamSlice, ","))

	websocketmanager.SendMsg(name, msg)
	fmt.Printf(websocketmanager.SubscribeMsg, name)
}

func (k *Korbit) HandleWsMessage(msgBytes []byte) {
	var rJson map[string]interface{}
	commons.Bytes2Json(msgBytes, &rJson)
	if rType, ok := rJson["type"].(string); ok {
		switch rType {
		case "orderbook":
			SetOrderbook("W", name, rJson)
		case "pong":
			fmt.Println("PONG")
		default:
			tgmanager.HandleErr(name, fmt.Errorf("unknown message: %s", rType))
		}
		return
	}
	tgmanager.HandleErr(name, fmt.Errorf("unknown message: %s", string(msgBytes)))
}

func (k *Korbit) HandleRestResponse(rJson map[string]interface{}) {
	SetOrderbook("R", name, rJson)
}

func Run(e string) {
	exchange.Run(e, &Korbit{})
}
