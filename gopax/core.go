package gopax

import (
	"encoding/json"
	"fmt"
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
	exchange string
	syncMap  sync.Map
)

type delta struct {
	action string
	ob     []interface{}
}

func pongWs(msg string) {
	websocketmanager.SendMsg(exchange, msg)
	fmt.Printf("%s PONG %s\n", exchange, msg)
}

func subscribeWs(pairs []string) {
	for _, pair := range pairs {
		var pairInfo = strings.Split(pair, ":")
		market, symbol := strings.ToUpper(pairInfo[0]), strings.ToUpper(pairInfo[1])

		jsonBytes, _ := json.Marshal(map[string]interface{}{
			"n": "SubscribeToOrderBook",
			"o": map[string]interface{}{
				"tradingPairName": fmt.Sprintf("%s-%s", symbol, market),
			},
		})
		websocketmanager.SendMsg(exchange, string(jsonBytes))
	}
	fmt.Printf(websocketmanager.SubscribeMsg, exchange)
}

func receiveWs() {
	for {
		_, msgBytes, err := websocketmanager.Conn(exchange).ReadMessage()
		tgmanager.HandleErr(exchange, err)

		var rJson interface{}

		msg := string(msgBytes)
		if strings.Contains(msg, "ping") {
			pongWs(strings.Replace(msg, "ping", "pong", -1))
		} else if strings.Contains(msg, "SubscribeToOrderBook") { // set full (init)
			commons.Bytes2Json(msgBytes, &rJson)

			o := rJson.(map[string]interface{})["o"].(map[string]interface{})
			pair := o["tradingPairName"].(string)
			symbol, market := strings.ToLower(strings.Split(pair, "-")[0]), strings.ToLower(strings.Split(pair, "-")[1])
			key := fmt.Sprintf("%s:%s", market, symbol)

			syncMap.Store(key, o)
		} else if strings.Contains(msg, "OrderBookEvent") { // set delta
			commons.Bytes2Json(msgBytes, &rJson)

			o := rJson.(map[string]interface{})["o"].(map[string]interface{})
			pair := o["tradingPairName"].(string)
			symbol, market := strings.ToLower(strings.Split(pair, "-")[0]), strings.ToLower(strings.Split(pair, "-")[1])
			key := fmt.Sprintf("%s:%s", market, symbol)

			var changed []interface{}
			changed = append(changed, delta{action: "ask", ob: o["ask"].([]interface{})})
			changed = append(changed, delta{action: "bid", ob: o["bid"].([]interface{})})

			var obAsk, obBid []interface{}
			var syncMapValue interface{}
			for _, c := range changed {
				if len(c.(delta).ob) > 0 {
					price := c.(delta).ob[0].(map[string]interface{})["price"].(float64)
					volume := c.(delta).ob[0].(map[string]interface{})["volume"].(float64)
					d := map[string]interface{}{
						"price":  price,
						"volume": volume,
					}

					syncMapValue, _ = syncMap.Load(key)
					obAsk = syncMapValue.(map[string]interface{})["ask"].([]interface{})
					obBid = syncMapValue.(map[string]interface{})["bid"].([]interface{})

					switch c.(delta).action {
					case "ask": // ask price going up
						min_price := obAsk[0].(map[string]interface{})["price"].(float64)
						max_price := obAsk[len(obAsk)-1].(map[string]interface{})["price"].(float64)

						if price < min_price { // prepend
							obAsk = append([]interface{}{d}, obAsk...)
						}

						for i := range obAsk { // exchange volume
							obPrice := obAsk[i].(map[string]interface{})["price"].(float64)
							if price == obPrice {
								obAsk[i].(map[string]interface{})["volume"] = volume
							}
						}

						if price > max_price { // append
							obAsk = append(obAsk, d)
						}

					case "bid": // bid price going down
						max_price := obBid[0].(map[string]interface{})["price"].(float64)
						min_price := obBid[len(obBid)-1].(map[string]interface{})["price"].(float64)

						if price > max_price { // prepend
							obBid = append([]interface{}{d}, obBid...)
						}

						for i := range obBid { // exchange volume
							obPrice := obBid[i].(map[string]interface{})["price"].(float64)
							if price == obPrice {
								obBid[i].(map[string]interface{})["volume"] = volume
							}
						}

						if price < min_price { // append
							obBid = append(obBid, d)
						}
					}
				}
			}

			// remove orderbook of volume 0
			var obAskF, obBidF []interface{}
			for i := range obAsk {
				obVolume := obAsk[i].(map[string]interface{})["volume"].(float64)
				if obVolume != 0 {
					obAskF = append(obAskF, obAsk[i])
				}
			}
			for i := range obBid {
				obVolume := obBid[i].(map[string]interface{})["volume"].(float64)
				if obVolume != 0 {
					obBidF = append(obBidF, obBid[i])
				}
			}
			syncMapValue.(map[string]interface{})["ask"] = obAskF
			syncMapValue.(map[string]interface{})["bid"] = obBidF
			syncMapValue.(map[string]interface{})["market"] = market
			syncMapValue.(map[string]interface{})["symbol"] = symbol

			SetOrderbook("W", exchange, syncMapValue.(map[string]interface{}))
		}
	}
}

func rest(pairs []string) {
	c := make(chan map[string]interface{})
	buffer, rateLimit := config.GetRateLimit(exchange)

	for {
		for _, pair := range pairs {
			go restmanager.FastHttpRequest(c, exchange, "GET", pair)
		}

		for i := 0; i < len(pairs); i++ {
			rJson := <-c
			SetOrderbook("R", exchange, rJson)
		}

		// 1번에 (1s / rateLimit)s 만큼 쉬어야 하고, 동시에 pair 만큼 api hit 하니, 그만큼 쉬어야함.
		// ex) 1 / 10 s * 2 = 0.2s => 200ms
		pairsLength := float64(len(pairs)) * buffer
		time.Sleep(time.Millisecond * time.Duration(int(1/rateLimit*pairsLength*10*100)))
	}
}

func Run(e string) {
	exchange = e
	pairs := config.GetPairs(exchange)
	var wg sync.WaitGroup

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
