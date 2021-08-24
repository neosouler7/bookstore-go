package binance

import (
	"fmt"
	"strings"
	"time"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/redismanager"
)

func getPairInterface(exchange string) map[string]interface{} {
	var pairs = commons.ReadConfig("Pairs").(map[string]interface{})[exchange]

	pairInterface := make(map[string]interface{})
	for _, pair := range pairs.([]interface{}) {
		var pairInfo = strings.Split(pair.(string), ":")
		var market = pairInfo[0]
		var symbol = pairInfo[1]

		pairInterface[fmt.Sprintf("%s%s", symbol, market)] = map[string]string{"market": market, "symbol": symbol}
	}

	return pairInterface
}

func SetOrderbook(api string, exchange string, rJson map[string]interface{}) {
	var pair, market, symbol string

	switch api {
	case "R":
		market = rJson["market"].(string)
		symbol = rJson["symbol"].(string)
	case "W":
		var pairInterface = getPairInterface(exchange)
		pair = strings.Split(rJson["stream"].(string), "@")[0]
		market = pairInterface[pair].(map[string]string)["market"]
		symbol = pairInterface[pair].(map[string]string)["symbol"]
	}

	ts := commons.FormatTs(fmt.Sprintf("%d", time.Now().UnixNano()/100000))

	var askResponse, bidResponse []interface{}
	var askSlice, bidSlice []interface{}

	switch api {
	case "R":
		askResponse = rJson["asks"].([]interface{})
		bidResponse = rJson["bids"].([]interface{})
	case "W":
		rData := rJson["data"]
		askResponse = rData.(map[string]interface{})["asks"].([]interface{})
		bidResponse = rData.(map[string]interface{})["bids"].([]interface{})
	}

	for i := 0; i < commons.Min(len(askResponse), len(bidResponse))-1; i++ {
		askR := askResponse[i].([]interface{})
		bidR := bidResponse[i].([]interface{})
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
