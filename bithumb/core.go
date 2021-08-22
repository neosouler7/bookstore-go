package bithumb

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/restmanager"
	"github.com/neosouler7/bookstore-go/websocketmanager"
)

var (
	exchange string
	rJsonMap map[string]interface{}
)

func subscribeWs(pairs interface{}) {
	time.Sleep(time.Second * 1)
	var streamSlice []string
	for _, pair := range pairs.([]interface{}) {
		var pairInfo = strings.Split(pair.(string), ":")
		var market = strings.ToUpper(pairInfo[0])
		var symbol = strings.ToUpper(pairInfo[1])

		streamSlice = append(streamSlice, fmt.Sprintf("\"%s_%s\"", symbol, market))
	}
	streams := strings.Join(streamSlice, ",")
	msg := fmt.Sprintf("{\"type\": \"orderbookdepth\",\"symbols\": [%s]}", streams)

	websocketmanager.SendMsg(exchange, msg)
	fmt.Println("BMB websocket subscribe msg sent!")
}

func restWs(pairs interface{}) {
	c := make(chan map[string]interface{})
	rJsonMap = make(map[string]interface{})

	// rest for each pairs just once
	for _, pair := range pairs.([]interface{}) {
		go restmanager.FastHttpRequest(c, exchange, "GET", pair.(string))
	}

	// and save whole data for tracing the changes by websocket
	for i := 0; i < len(pairs.([]interface{})); i++ {
		rJson := <-c
		market := strings.ToLower(rJson["payment_currency"].(string))
		symbol := strings.ToLower(rJson["order_currency"].(string))
		rJsonMap[fmt.Sprintf("%s:%s", market, symbol)] = rJson
	}

	// init websocket
	go subscribeWs(pairs)

	for {
		_, msgBytes, err := websocketmanager.Conn(exchange).ReadMessage()
		commons.HandleErr(err, websocketmanager.ErrReadMsg)

		if strings.Contains(string(msgBytes), "Successfully") {
			fmt.Printf("%s\n", string(msgBytes))
		} else if strings.Contains(string(msgBytes), "orderbookdepth") {
			var rJson interface{}
			commons.Bytes2Json(msgBytes, &rJson)

			changed := rJson.(map[string]interface{})["content"].(map[string]interface{})["list"].([]interface{})
			var pairInfo = strings.Split(changed[0].(map[string]interface{})["symbol"].(string), "_")
			var market = strings.ToLower(pairInfo[1])
			var symbol = strings.ToLower(pairInfo[0])

			if rJsonMap[fmt.Sprintf("%s:%s", market, symbol)] == nil {
				fmt.Printf("pass ws of %s:%s since no rest value\n", market, symbol)
			} else {
				fmt.Println(changed) // content > list
				// TODO
				// from websocket 변경호가
				// {"type":"orderbookdepth","content":{"list":[{"symbol":"XRP_KRW","orderType":"ask","price":"1448","quantity":"12565.6221","total":"7"},{"symbol":"XRP_KRW","orderType":"ask","price":"1449","quantity":"18449.1301","total":"10"},{"symbol":"XRP_KRW","orderType":"ask","price":"1456","quantity":"223405.2277","total":"18"},{"symbol":"XRP_KRW","orderType":"bid","price":"1444","quantity":"52551.0719","total":"61"},{"symbol":"XRP_KRW","orderType":"bid","price":"1442","quantity":"78614.4644","total":"33"},{"symbol":"XRP_KRW","orderType":"bid","price":"1441","quantity":"70382.8896","total":"20"},{"symbol":"XRP_KRW","orderType":"bid","price":"1440","quantity":"27325.6912","total":"38"},{"symbol":"XRP_KRW","orderType":"bid","price":"1437","quantity":"22941.0609","total":"25"}],"datetime":"1629599007105405"}}

				// get rJson[market:symbol]
				// for each ask/bid loop, and change value or pop
				// set on orderbook

				// SetOrderbook("W", exchange, rJson[market:symbol].(map[string]interface{}))
			}

		}
	}
}

func onlyRest(pairs interface{}) {
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

	// since bmb returns changed value of orderbooks
	wg.Add(1)
	// go onlyRest(pairs)
	go restWs(pairs)

	wg.Wait()
}
