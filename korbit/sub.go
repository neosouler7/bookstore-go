package korbit

import (
	"strconv"
	"strings"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/redismanager"
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

	rData := rJson["data"]
	timestamp := rData.(map[string]interface{})["timestamp"].(float64)
	ts = commons.FormatTs(strconv.FormatFloat(timestamp, 'f', -1, 64))

	var askResponse, bidResponse []interface{}
	var askSlice, bidSlice []interface{}

	askResponse = rData.(map[string]interface{})["asks"].([]interface{})
	bidResponse = rData.(map[string]interface{})["bids"].([]interface{})

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
