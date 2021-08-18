package huobikorea

import (
	"fmt"
	"strings"

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
	var pair, market, symbol, ts string
	var pairInterface = getPairInterface(exchange)

	pair = strings.Split(rJson["ch"].(string), ".")[1]
	market = pairInterface[pair].(map[string]string)["market"]
	symbol = pairInterface[pair].(map[string]string)["symbol"]
	ts = commons.FormatTs(fmt.Sprintf("%f", rJson["ts"].(float64)))

	var askResponse, bidResponse []interface{}
	var askSlice, bidSlice []interface{}

	rData := rJson["tick"]
	askResponse = rData.(map[string]interface{})["asks"].([]interface{})
	bidResponse = rData.(map[string]interface{})["bids"].([]interface{})

	// hbk does not guarantee ask/bid same length
	for i := 0; i < commons.Min(len(askResponse), len(bidResponse))-1; i++ {
		askR := askResponse[i].([]interface{})
		bidR := bidResponse[i].([]interface{})
		ask := [2]string{fmt.Sprintf("%f", askR[0].(float64)), fmt.Sprintf("%f", askR[1].(float64))}
		bid := [2]string{fmt.Sprintf("%f", bidR[0].(float64)), fmt.Sprintf("%f", bidR[1].(float64))}
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
