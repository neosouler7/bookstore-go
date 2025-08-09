package korbit

import (
	"strconv"
	"strings"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/redismanager"
	"github.com/neosouler7/bookstore-go/tgmanager"
)

func SetOrderbook(api string, exchange string, rJson map[string]interface{}) {
	// fmt.Println(rJson)
	var market, symbol, ts string
	switch api {
	case "R":
		market, symbol = rJson["market"].(string), rJson["symbol"].(string)
	case "W":
		market = strings.Split(rJson["symbol"].(string), "_")[1]
		symbol = strings.Split(rJson["symbol"].(string), "_")[0]
	}

	rData := rJson["data"].(map[string]interface{})
	timestamp := rData["timestamp"].(float64)
	ts = commons.FormatTs(strconv.FormatFloat(timestamp, 'f', -1, 64))

	var askResponse, bidResponse []interface{}
	askResponse = rData["asks"].([]interface{})
	bidResponse = rData["bids"].([]interface{})

	askSlice := make([]interface{}, 0, commons.Max(len(askResponse), len(bidResponse))) // 용량 미리 할당
	bidSlice := make([]interface{}, 0, commons.Max(len(askResponse), len(bidResponse))) // 용량 미리 할당

	switch api {
	case "R":
		for i := 0; i < commons.Min(len(askResponse), len(bidResponse))-1; i++ {
			askR, bidR := askResponse[i].(map[string]interface{}), bidResponse[i].(map[string]interface{})
			ask := [2]string{askR["price"].(string), askR["qty"].(string)}
			bid := [2]string{bidR["price"].(string), bidR["qty"].(string)}
			askSlice = append(askSlice, ask)
			bidSlice = append(bidSlice, bid)
		}
	case "W":
		for i := 0; i < commons.Min(len(askResponse), len(bidResponse))-1; i++ {
			askR, bidR := askResponse[i].(map[string]interface{}), bidResponse[i].(map[string]interface{})
			ask := [2]string{askR["price"].(string), askR["qty"].(string)}
			bid := [2]string{bidR["price"].(string), bidR["qty"].(string)}
			askSlice = append(askSlice, ask)
			bidSlice = append(bidSlice, bid)
		}
	}

	if err := redismanager.PreHandleOrderbook(
		api,
		exchange,
		market,
		symbol,
		askSlice,
		bidSlice,
		ts,
	); err != nil {
		tgmanager.HandleErr(exchange, err)
	}
}
