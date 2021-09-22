package binance

import (
	"fmt"
	"strings"
	"time"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/redismanager"
)

func SetOrderbook(api string, exchange string, rJson map[string]interface{}) {
	var pair, market, symbol string

	switch api {
	case "R":
		market = rJson["market"].(string)
		symbol = rJson["symbol"].(string)
	case "W":
		var pairMap = commons.GetPairMap(exchange)
		pair = strings.Split(rJson["stream"].(string), "@")[0]
		market, symbol = pairMap[pair].(map[string]string)["market"], pairMap[pair].(map[string]string)["symbol"]
	}

	ts := commons.FormatTs(fmt.Sprintf("%d", int(time.Now().UnixMilli())))

	var askResponse, bidResponse []interface{}
	switch api {
	case "R":
		askResponse, bidResponse = rJson["asks"].([]interface{}), rJson["bids"].([]interface{})
	case "W":
		rData := rJson["data"]
		askResponse, bidResponse = rData.(map[string]interface{})["asks"].([]interface{}), rData.(map[string]interface{})["bids"].([]interface{})
	}

	var askSlice, bidSlice []interface{}
	for i := 0; i < commons.Min(len(askResponse), len(bidResponse))-1; i++ {
		askR, bidR := askResponse[i].([]interface{}), bidResponse[i].([]interface{})
		ask := [2]string{askR[0].(string), askR[1].(string)}
		bid := [2]string{bidR[0].(string), bidR[1].(string)}
		askSlice = append(askSlice, ask)
		bidSlice = append(bidSlice, bid)
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
