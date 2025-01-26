package korbit

import (
	"strconv"
	"strings"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/redismanager"
)

func SetOrderbook(api string, exchange string, rJson map[string]interface{}) {
	var market, symbol, ts string
	switch api {
	case "R":
		market, symbol = rJson["market"].(string), rJson["symbol"].(string)
	case "W":
		market = strings.Split(rJson["data"].(map[string]interface{})["currency_pair"].(string), "_")[1]
		symbol = strings.Split(rJson["data"].(map[string]interface{})["currency_pair"].(string), "_")[0]
	}
	timestamp := rJson["data"].(map[string]interface{})["timestamp"].(float64)
	ts = commons.FormatTs(strconv.FormatFloat(timestamp, 'f', -1, 64))

	var askResponse, bidResponse []interface{}
	var askSlice, bidSlice []interface{}

	rData := rJson["data"]
	askResponse = rData.(map[string]interface{})["asks"].([]interface{})
	bidResponse = rData.(map[string]interface{})["bids"].([]interface{})

	switch api {
	case "R":
		// askResponse = rJson["asks"].([]interface{})
		// bidResponse = rJson["bids"].([]interface{})

		for i := 0; i < commons.Min(len(askResponse), len(bidResponse))-1; i++ {
			// askR, bidR := askResponse[i].([]interface{}), bidResponse[i].([]interface{})
			// ask := [2]string{askR[0].(string), askR[1].(string)}
			// bid := [2]string{bidR[0].(string), bidR[1].(string)}
			askR, bidR := askResponse[i].(map[string]interface{}), bidResponse[i].(map[string]interface{})
			ask := [2]string{askR["price"].(string), askR["qty"].(string)}
			bid := [2]string{bidR["price"].(string), bidR["qty"].(string)}
			askSlice = append(askSlice, ask)
			bidSlice = append(bidSlice, bid)
		}
	case "W":
		// rData := rJson["data"]
		// askResponse = rData.(map[string]interface{})["asks"].([]interface{})
		// bidResponse = rData.(map[string]interface{})["bids"].([]interface{})

		for i := 0; i < commons.Min(len(askResponse), len(bidResponse))-1; i++ {
			askR, bidR := askResponse[i].(map[string]interface{}), bidResponse[i].(map[string]interface{})
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
