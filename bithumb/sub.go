package bithumb

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/redismanager"
)

func SetOrderbook(api string, exchange string, rJson map[string]interface{}) {
	var market, symbol string
	var askSlice, bidSlice []interface{}

	marketKey := map[string]string{"R": "market", "W": "code"}[api]
	pairInfo := strings.Split(rJson[marketKey].(string), "-")

	market, symbol = strings.ToLower(pairInfo[0]), strings.ToLower(pairInfo[1])
	tsFloat := int(rJson["timestamp"].(float64))
	ts := commons.FormatTs(strconv.Itoa(tsFloat))

	orderbookUnits := rJson["orderbook_units"].([]interface{})

	for _, unit := range orderbookUnits {
		orderbookUnit := unit.(map[string]interface{})
		askSlice = append(askSlice, [2]string{
			fmt.Sprintf("%f", orderbookUnit["ask_price"].(float64)),
			fmt.Sprintf("%f", orderbookUnit["ask_size"].(float64)),
		})
		bidSlice = append(bidSlice, [2]string{
			fmt.Sprintf("%f", orderbookUnit["bid_price"].(float64)),
			fmt.Sprintf("%f", orderbookUnit["bid_size"].(float64)),
		})
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
