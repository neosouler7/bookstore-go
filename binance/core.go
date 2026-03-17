package binance

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/exchange"
	"github.com/neosouler7/bookstore-go/websocketmanager"
)

const name = "bin"

type Binance struct{}

func (b *Binance) Ping() {
	websocketmanager.Ping(name)
}

func (b *Binance) Subscribe(pairs []string, wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(time.Second * 1)

	var streamSlice []string
	for _, pair := range pairs {
		pairInfo := strings.Split(pair, ":")
		market, symbol := strings.ToLower(pairInfo[0]), strings.ToLower(pairInfo[1])
		streamSlice = append(streamSlice, fmt.Sprintf("\"%s%s@depth20@100ms\"", symbol, market))
	}

	streams := strings.Join(streamSlice, ",")
	msg := fmt.Sprintf("{\"method\": \"SUBSCRIBE\",\"params\": [%s],\"id\": %d}", streams, time.Now().UnixNano()/100000)

	websocketmanager.SendMsg(name, msg)
	fmt.Printf(websocketmanager.SubscribeMsg, name)
}

func (b *Binance) HandleWsMessage(msgBytes []byte) {
	if strings.Contains(string(msgBytes), "result") {
		fmt.Printf(websocketmanager.FilteredMsg, name, string(msgBytes))
		return
	}
	var rJson interface{}
	commons.Bytes2Json(msgBytes, &rJson)
	SetOrderbook("W", name, rJson.(map[string]interface{}))
}

func (b *Binance) HandleRestResponse(rJson map[string]interface{}) {
	SetOrderbook("R", name, rJson)
}

func Run(e string) {
	exchange.Run(e, &Binance{})
}
