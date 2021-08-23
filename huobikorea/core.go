package huobikorea

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
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

func pingWs(ts interface{}) {
	msg := fmt.Sprintf("{\"pong\":%d}", int(ts.(float64)))
	websocketmanager.SendMsg(exchange, msg)
	fmt.Printf("HBK PONG %s\n", msg)
}

func subscribeWs(pairs interface{}) {
	time.Sleep(time.Second * 1)
	for _, pair := range pairs.([]interface{}) {
		var pairInfo = strings.Split(pair.(string), ":")
		var market = strings.ToLower(pairInfo[0])
		var symbol = strings.ToLower(pairInfo[1])

		msg := fmt.Sprintf("{\"sub\": \"market.%s%s.depth.step0\"}", symbol, market)
		websocketmanager.SendMsg(exchange, msg)
		fmt.Println("HBK websocket subscribe msg sent!")
	}

}

func receiveWs() {
	for {
		_, msgBytes, err := websocketmanager.Conn(exchange).ReadMessage()
		tgmanager.HandleErr(err, websocketmanager.ErrReadMsg)

		gzip, err := gzip.NewReader(bytes.NewReader(msgBytes))
		if err != nil {
			log.Fatalln(err)
		}

		gzipMsg := new(strings.Builder)
		_, err = io.Copy(gzipMsg, gzip)
		if err != nil {
			log.Fatalln(err)
		}

		if strings.Contains(gzipMsg.String(), "subbed") {
			continue
		} else if strings.Contains(gzipMsg.String(), "ping") {
			var rJson interface{}
			commons.Bytes2Json([]byte(gzipMsg.String()), &rJson)

			pingTs := rJson.(map[string]interface{})["ping"]
			pingWs(pingTs)
		} else if strings.Contains(gzipMsg.String(), "tick") {
			var rJson interface{}
			commons.Bytes2Json([]byte(gzipMsg.String()), &rJson)

			SetOrderbook("W", exchange, rJson.(map[string]interface{}))
		} else {
			log.Fatalln(gzipMsg.String())
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
