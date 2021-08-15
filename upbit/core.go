package upbit

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

	"github.com/google/uuid"
)

const latencyAllowed float64 = 10.0 // per 1 second
var ex string

func pingWs() {
	msg := "PING"
	for {
		err := websocketmanager.SendMsg(ex, msg)
		if err != nil {
			log.Fatalln(err)
		}
		time.Sleep(time.Second * 5)
	}
}

func subscribeWs(pairs interface{}) {
	time.Sleep(time.Second * 1)
	var streamSlice []string
	for _, pair := range pairs.([]interface{}) {
		var pairInfo = strings.Split(pair.(string), ":")
		var market = strings.ToUpper(pairInfo[0])
		var symbol = strings.ToUpper(pairInfo[1])

		streamSlice = append(streamSlice, fmt.Sprintf("'%s-%s'", market, symbol))
	}
	uuid := uuid.NewString()
	streams := strings.Join(streamSlice, ",")
	msg := fmt.Sprintf("[{'ticket':'%s'}, {'type': 'orderbook', 'codes': [%s]}]", uuid, streams)

	err := websocketmanager.SendMsg(ex, msg)
	fmt.Println("UPB websocket subscribe msg sent!")
	if err != nil {
		log.Fatalln(err)
	}
}

func receiveWs() {
	for {
		_, message, err := websocketmanager.GetConn(ex).ReadMessage()
		if err != nil {
			log.Fatalln(err)
		}

		if strings.Contains(string(message), "status") {
			fmt.Println("PONG") // {"status":"UP"}
		} else {
			var rJson interface{}
			err = json.Unmarshal(message, &rJson)
			if err != nil {
				log.Fatalln(err)
			}

			err := SetOrderbook("W", ex, rJson.(map[string]interface{}))
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
}

func rest(pairs interface{}) {
	c := make(chan map[string]interface{})

	for {
		for _, pair := range pairs.([]interface{}) {
			go restmanager.FastHttpRequest(c, ex, "GET", pair.(string))
		}

		for i := 0; i < len(pairs.([]interface{})); i++ {
			rJson := <-c

			err := SetOrderbook("R", ex, rJson)
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
	ex = exchange
	var pairs = commons.ReadConfig("Pairs").(map[string]interface{})[exchange]

	var wg sync.WaitGroup

	// [ping]
	wg.Add(1)
	go pingWs()

	// [subscribe websocket stream]
	wg.Add(1)
	go func() {
		subscribeWs(pairs)
		wg.Done()
	}()

	// [receive websocket msg]
	wg.Add(1)
	go receiveWs()

	// [rest]
	wg.Add(1)
	go rest(pairs)

	wg.Wait()
}
