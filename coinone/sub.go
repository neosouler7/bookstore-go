package coinone

import (
	"fmt"
	"strings"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/redismanager"
)

func SetOrderbook(api string, exchange string, rJson map[string]interface{}) {
	// con differs "market-symbol" receive form by api type
	var market, symbol, ts string
	var askResponse, bidResponse []interface{}
	var askSlice, bidSlice []interface{}
	switch api {
	case "R":
		market = "krw"
		symbol = rJson["currency"].(string)

		tsString := rJson["timestamp"].(string)
		ts = commons.FormatTs(tsString)

		askResponse = rJson["ask"].([]interface{})
		bidResponse = rJson["bid"].([]interface{})
	case "W":
		rTopic := rJson["topic"].(map[string]interface{})
		market = strings.ToLower(rTopic["priceCurrency"].(string))
		symbol = strings.ToLower(rTopic["productCurrency"].(string))

		rData := rJson["data"]
		tsFloat := int(rData.(map[string]interface{})["timestamp"].(float64))
		ts = commons.FormatTs(fmt.Sprintf("%d", tsFloat))

		askResponse = rData.(map[string]interface{})["ask"].([]interface{})
		bidResponse = rData.(map[string]interface{})["bid"].([]interface{})
	}

	if len(askResponse) == len(bidResponse) {
		for i := range askResponse {
			askR := askResponse[i]
			bidR := bidResponse[i]
			ask := [2]string{fmt.Sprintf("%s", askR.(map[string]interface{})["price"]), fmt.Sprintf("%s", askR.(map[string]interface{})["qty"])}
			bid := [2]string{fmt.Sprintf("%s", bidR.(map[string]interface{})["price"]), fmt.Sprintf("%s", bidR.(map[string]interface{})["qty"])}
			askSlice = append(askSlice, ask)
			bidSlice = append(bidSlice, bid)
		}
	}

	redismanager.PreHandleOrderbook(
		api,
		exchange,
		market,
		symbol,
		askSlice,
		bidSlice,
		ts,
	)
}
