package huobikorea

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"neosouler7/bookstore-go/commons"
	"neosouler7/bookstore-go/restmanager"
	"neosouler7/bookstore-go/websocketmanager"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const latencyAllowed float64 = 10.0 // per 1 second

// set redis pipeline global
// var pipe = redisManager.GetRedisPipeline()

// func upbExecPipeline(exchange string) {
// 	for {
// 		redisManager.ExecPipeline(exchange, &pipe)
// 		time.Sleep(time.Second * 1)
// 		fmt.Println("UPB execute pipeline")
// 	}
// }

func pingWs(wsConn *websocket.Conn, ts interface{}) {
	msg := fmt.Sprintf("{\"pong\":%d}", int(ts.(float64)))
	err := websocketmanager.SendMsg(wsConn, msg)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("HBK PONG %s\n", msg)
}

func subscribeWs(wsConn *websocket.Conn, pairs interface{}) {
	time.Sleep(time.Second * 1)
	for _, pair := range pairs.([]interface{}) {
		var pairInfo = strings.Split(pair.(string), ":")
		var market = strings.ToLower(pairInfo[0])
		var symbol = strings.ToLower(pairInfo[1])

		msg := fmt.Sprintf("{\"sub\": \"market.%s%s.depth.step0\"}", symbol, market)
		err := websocketmanager.SendMsg(wsConn, msg)
		fmt.Println("HBK websocket subscribe msg sent!")
		if err != nil {
			log.Fatalln(err)
		}
	}

}

func receiveWs(wsConn *websocket.Conn, exchange string) {
	for {
		_, message, err := wsConn.ReadMessage()
		if err != nil {
			log.Fatalln(err)
		}

		gzip, err := gzip.NewReader(bytes.NewReader(message))
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
			err = json.Unmarshal([]byte(gzipMsg.String()), &rJson)
			if err != nil {
				log.Fatalln(err)
			}
			pingTs := rJson.(map[string]interface{})["ping"]
			pingWs(wsConn, pingTs)

		} else if strings.Contains(gzipMsg.String(), "tick") {
			var rJson interface{}
			err = json.Unmarshal([]byte(gzipMsg.String()), &rJson)
			if err != nil {
				log.Fatalln(err)
			}

			err := SetOrderbook("W", exchange, rJson.(map[string]interface{}))
			if err != nil {
				log.Fatalln(err)
			}
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

	// // [ping]
	// wg.Add(1)
	// go pingWs(wsConn)

	// [subscribe websocket stream]
	wg.Add(1)
	go func() {
		subscribeWs(wsConn, pairs)
		wg.Done()
	}()

	// [receive websocket msg]
	wg.Add(1)
	go receiveWs(wsConn, exchange)

	// [rest]
	wg.Add(1)
	go rest(exchange, pairs)

	wg.Wait()
}