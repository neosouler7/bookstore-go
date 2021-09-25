package bithumb

import (
	"fmt"
	"strings"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/redismanager"
)

func SetOrderbook(api string, exchange string, rJson map[string]interface{}) {
	market, symbol := strings.ToLower(rJson["payment_currency"].(string)), strings.ToLower(rJson["order_currency"].(string))

	ts := commons.FormatTs(fmt.Sprintf("%s", rJson["timestamp"]))

	askResponse, bidResponse := rJson["asks"].([]interface{}), rJson["bids"].([]interface{})

	var askSlice, bidSlice []interface{}
	for i := 0; i < commons.Min(len(askResponse), len(bidResponse))-1; i++ {
		askR, bidR := askResponse[i].(map[string]interface{}), bidResponse[i].(map[string]interface{})
		ask := [2]string{fmt.Sprintf("%s", askR["price"]), fmt.Sprintf("%s", askR["quantity"])}
		bid := [2]string{fmt.Sprintf("%s", bidR["price"]), fmt.Sprintf("%s", bidR["quantity"])}
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
