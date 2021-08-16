package upbit

import (
	"fmt"
	"neosouler7/bookstore-go/commons"
	"neosouler7/bookstore-go/redismanager"
	"strings"
)

func SetOrderbook(api string, exchange string, rJson map[string]interface{}) {
	// upb differs "market-symbol" receive form by api type
	var pair string
	switch api {
	case "R":
		pair = rJson["market"].(string)
	case "W":
		pair = rJson["code"].(string)
	}
	var pairInfo = strings.Split(pair, "-")
	var market = strings.ToLower(pairInfo[0])
	var symbol = strings.ToLower(pairInfo[1])

	tsFloat := int(rJson["timestamp"].(float64))
	ts := commons.FormatTs(fmt.Sprintf("%d", tsFloat))
	orderbooks := rJson["orderbook_units"].([]interface{})

	var askSlice []interface{}
	var bidSlice []interface{}
	for _, ob := range orderbooks {
		o := ob.(map[string]interface{})
		ask := [2]string{fmt.Sprintf("%f", o["ask_price"]), fmt.Sprintf("%f", o["ask_size"])}
		bid := [2]string{fmt.Sprintf("%f", o["bid_price"]), fmt.Sprintf("%f", o["bid_size"])}
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
