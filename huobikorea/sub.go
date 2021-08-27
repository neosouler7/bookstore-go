package huobikorea

import (
	"fmt"
	"strings"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/redismanager"
)

func SetOrderbook(api string, exchange string, rJson map[string]interface{}) {
	var pairMap = commons.GetPairMap(exchange)
	pair := strings.Split(rJson["ch"].(string), ".")[1]
	market, symbol := pairMap[pair].(map[string]string)["market"], pairMap[pair].(map[string]string)["symbol"]
	ts := commons.FormatTs(fmt.Sprintf("%f", rJson["ts"].(float64)))

	rData := rJson["tick"]
	askResponse := rData.(map[string]interface{})["asks"].([]interface{})
	bidResponse := rData.(map[string]interface{})["bids"].([]interface{})

	var askSlice, bidSlice []interface{}
	for i := 0; i < commons.Min(len(askResponse), len(bidResponse))-1; i++ {
		askR, bidR := askResponse[i].([]interface{}), bidResponse[i].([]interface{})
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
