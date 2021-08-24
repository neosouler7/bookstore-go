package korbit

import (
	"fmt"
	"strings"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/redismanager"
)

func SetOrderbook(api string, exchange string, rJson map[string]interface{}) {
	var market, symbol, ts string

	switch api {
	case "R":
		market = rJson["market"].(string)
		symbol = rJson["symbol"].(string)
		ts = commons.FormatTs(fmt.Sprintf("%f", rJson["timestamp"].(float64)))
	case "W":
		market = strings.Split(rJson["data"].(map[string]interface{})["currency_pair"].(string), "_")[1]
		symbol = strings.Split(rJson["data"].(map[string]interface{})["currency_pair"].(string), "_")[0]
		ts = commons.FormatTs(fmt.Sprintf("%f", rJson["data"].(map[string]interface{})["timestamp"].(float64)))
	}

	var askResponse, bidResponse []interface{}
	var askSlice, bidSlice []interface{}

	switch api {
	case "R":
		askResponse = rJson["asks"].([]interface{})
		bidResponse = rJson["bids"].([]interface{})

		for i := 0; i < commons.Min(len(askResponse), len(bidResponse))-1; i++ {
			askR := askResponse[i].([]interface{})
			bidR := bidResponse[i].([]interface{})
			ask := [2]string{askR[0].(string), askR[1].(string)}
			bid := [2]string{bidR[0].(string), bidR[1].(string)}
			askSlice = append(askSlice, ask)
			bidSlice = append(bidSlice, bid)
		}
	case "W":
		rData := rJson["data"]
		askResponse = rData.(map[string]interface{})["asks"].([]interface{})
		bidResponse = rData.(map[string]interface{})["bids"].([]interface{})

		for i := 0; i < commons.Min(len(askResponse), len(bidResponse))-1; i++ {
			askR := askResponse[i].(map[string]interface{})
			bidR := bidResponse[i].(map[string]interface{})
			ask := [2]string{askR["price"].(string), askR["amount"].(string)}
			bid := [2]string{bidR["price"].(string), bidR["amount"].(string)}
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
