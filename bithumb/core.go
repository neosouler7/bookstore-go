package bithumb

import (
	"fmt"
	"strconv"
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
	rJsonWs  map[string]interface{}
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

func receiveWs(pairs interface{}) {
	c := make(chan map[string]interface{})
	rJsonWs = make(map[string]interface{})

	// rest for each pairs just once
	for _, pair := range pairs.([]interface{}) {
		go restmanager.FastHttpRequest(c, exchange, "GET", pair.(string))
	}

	// and save whole data for tracing the changes by websocket
	for i := 0; i < len(pairs.([]interface{})); i++ {
		rJson := <-c
		market := strings.ToLower(rJson["payment_currency"].(string))
		symbol := strings.ToLower(rJson["order_currency"].(string))
		rJsonWs[fmt.Sprintf("%s:%s", market, symbol)] = rJson
	}

	// init websocket
	go subscribeWs(pairs)

	for {
		_, msgBytes, err := websocketmanager.Conn(exchange).ReadMessage()
		tgmanager.HandleErr(exchange, err, websocketmanager.ErrReadMsg)

		if strings.Contains(string(msgBytes), "Successfully") {
			fmt.Printf("%s\n", string(msgBytes))
		} else if strings.Contains(string(msgBytes), "orderbookdepth") {
			var rJson interface{}
			commons.Bytes2Json(msgBytes, &rJson)

			content := rJson.(map[string]interface{})["content"]
			ts := content.(map[string]interface{})["datetime"].(string)
			changed := content.(map[string]interface{})["list"].([]interface{})
			pairInfo := strings.Split(changed[0].(map[string]interface{})["symbol"].(string), "_")
			market := strings.ToLower(pairInfo[1])
			symbol := strings.ToLower(pairInfo[0])
			key := fmt.Sprintf("%s:%s", market, symbol)

			if rJsonWs[key] == nil {
				fmt.Printf("pass ws of %s:%s since no rest value\n", market, symbol)
			} else {
				var obAsk []interface{}
				var obBid []interface{}

				for _, c := range changed {
					action := c.(map[string]interface{})["orderType"]
					price, _ := strconv.ParseFloat(c.(map[string]interface{})["price"].(string), 64)
					quantity := c.(map[string]interface{})["quantity"]

					obAsk = rJsonWs[key].(map[string]interface{})["asks"].([]interface{})
					obBid = rJsonWs[key].(map[string]interface{})["bids"].([]interface{})
					switch action {
					case "ask": // ask price going up
						min_price, _ := strconv.ParseFloat(obAsk[0].(map[string]interface{})["price"].(string), 64)
						max_price, _ := strconv.ParseFloat(obAsk[len(obAsk)-1].(map[string]interface{})["price"].(string), 64)
						a := map[string]interface{}{
							"price":    fmt.Sprintf("%f", price),
							"quantity": quantity,
						}

						if price < min_price { // prepend
							obAsk = append([]interface{}{a}, obAsk...)
						}

						for i := range obAsk {
							obPrice, _ := strconv.ParseFloat(obAsk[i].(map[string]interface{})["price"].(string), 64)
							if price == obPrice {
								obAsk[i].(map[string]interface{})["quantity"] = quantity
							}
						}

						if price > max_price {
							obAsk = append(obAsk, a)
						}

					case "bid": // bid price going down
						max_price, _ := strconv.ParseFloat(obBid[0].(map[string]interface{})["price"].(string), 64)
						min_price, _ := strconv.ParseFloat(obBid[len(obBid)-1].(map[string]interface{})["price"].(string), 64)
						b := map[string]interface{}{
							"price":    fmt.Sprintf("%f", price),
							"quantity": quantity,
						}

						if price > max_price { // prepend
							obBid = append([]interface{}{b}, obBid...)
						}

						for i := range obBid {
							obPrice, _ := strconv.ParseFloat(obBid[i].(map[string]interface{})["price"].(string), 64)
							if price == obPrice {
								obBid[i].(map[string]interface{})["quantity"] = quantity
							}
						}

						if price < min_price {
							obBid = append(obBid, b)
						}
					}
				}

				// remove orderbook of volume 0
				var obAskF []interface{}
				var obBidF []interface{}
				for i := range obAsk {
					obQuantity, _ := strconv.ParseFloat(obAsk[i].(map[string]interface{})["quantity"].(string), 64)
					if obQuantity != 0 {
						obAskF = append(obAskF, obAsk[i])
					}
				}
				for i := range obBid {
					obQuantity, _ := strconv.ParseFloat(obBid[i].(map[string]interface{})["quantity"].(string), 64)
					if obQuantity != 0 {
						obBidF = append(obBidF, obBid[i])
					}
				}
				rJsonWs[key].(map[string]interface{})["asks"] = obAskF
				rJsonWs[key].(map[string]interface{})["bids"] = obBidF

				// update timestamp
				rJsonWs[key].(map[string]interface{})["timestamp"] = ts

				SetOrderbook("W", exchange, rJsonWs[key].(map[string]interface{}))
			}
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
			market := strings.ToLower(rJson["payment_currency"].(string))
			symbol := strings.ToLower(rJson["order_currency"].(string))
			rJsonWs[fmt.Sprintf("%s:%s", market, symbol)] = rJson
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

	// bmb returns CHANGED orderbooks
	wg.Add(1)
	go receiveWs(pairs)

	wg.Add(1)
	go rest(pairs)

	wg.Wait()
}
