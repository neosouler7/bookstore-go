package coinone

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/redismanager"
	"github.com/neosouler7/bookstore-go/tgmanager"
)

func SetOrderbook(api string, exchange string, rJson map[string]interface{}) {
	var rData map[string]interface{}
	switch api {
	case "R":
		rData = rJson
	case "W":
		rData = rJson["data"].(map[string]interface{})
	}
	market, symbol := strings.ToLower(rData["quote_currency"].(string)), strings.ToLower(rData["target_currency"].(string))
	tsFloat := int(rData["timestamp"].(float64))
	ts := commons.FormatTs(strconv.Itoa(tsFloat))

	var askResponse, bidResponse []interface{}
	var askSlice, bidSlice []interface{}

	askResponse = rData["asks"].([]interface{})
	bidResponse = rData["bids"].([]interface{})

	// WS returns asks in descending price order; sort ascending before passing to GetObTargetPrice
	sort.Slice(askResponse, func(i, j int) bool {
		priceIStr := askResponse[i].(map[string]interface{})["price"].(string)
		priceJStr := askResponse[j].(map[string]interface{})["price"].(string)

		priceI, err1 := strconv.ParseFloat(priceIStr, 64)
		priceJ, err2 := strconv.ParseFloat(priceJStr, 64)

		if err1 != nil || err2 != nil {
			tgmanager.HandleErr(exchange, fmt.Errorf("price parse error: %v, %v\n", err1, err2))
		}
		return priceI < priceJ
	})

	for i := 0; i < commons.Min(len(askResponse), len(bidResponse))-1; i++ {
		askR, bidR := askResponse[i].(map[string]interface{}), bidResponse[i].(map[string]interface{})
		ask := [2]string{askR["price"].(string), askR["qty"].(string)}
		bid := [2]string{bidR["price"].(string), bidR["qty"].(string)}
		askSlice = append(askSlice, ask)
		bidSlice = append(bidSlice, bid)
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
