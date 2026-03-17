package bithumb

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/exchange"
	"github.com/neosouler7/bookstore-go/websocketmanager"
)

const name = "bmb"

type Bithumb struct{}

func (b *Bithumb) Ping() {
	websocketmanager.Ping(name)
}

func (b *Bithumb) Subscribe(pairs []string, wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(time.Second * 1)

	var streamSlice []string
	for _, pair := range pairs {
		var pairInfo = strings.Split(pair, ":")
		market, symbol := strings.ToUpper(pairInfo[0]), strings.ToUpper(pairInfo[1])
		streamSlice = append(streamSlice, fmt.Sprintf("\"%s-%s\"", market, symbol))
	}

	uuid := uuid.NewString()
	streams := strings.Join(streamSlice, ",")
	msg := fmt.Sprintf("[{\"ticket\": \"%s\"}, {\"type\": \"orderbook\", \"isOnlyRealtime\": \"True\", \"level\": \"1\", \"codes\": [%s]}]", uuid, streams)

	websocketmanager.SendMsg(name, msg)
	fmt.Printf(websocketmanager.SubscribeMsg, name)
}

func (b *Bithumb) HandleWsMessage(msgBytes []byte) {
	var rJson map[string]interface{}
	commons.Bytes2Json(msgBytes, &rJson)
	if _, ok := rJson["status"]; ok {
		fmt.Println("PONG") // {"status":"UP"}
		return
	}
	SetOrderbook("W", name, rJson)
}

func (b *Bithumb) HandleRestResponse(rJson map[string]interface{}) {
	SetOrderbook("R", name, rJson)
}

func Run(e string) {
	exchange.Run(e, &Bithumb{})
}
