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
	askPrices := make([]float64, len(askResponse))
	for i, v := range askResponse {
		priceStr := v.(map[string]interface{})["price"].(string)
		p, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			tgmanager.HandleErr(exchange, fmt.Errorf("price parse error: %v", err))
		}
		askPrices[i] = p
	}
	sort.Slice(askResponse, func(i, j int) bool {
		return askPrices[i] < askPrices[j]
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
