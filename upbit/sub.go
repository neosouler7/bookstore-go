package upbit

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/redismanager"
)

func SetOrderbook(api string, exchange string, rJson map[string]interface{}) {
	var pair string
	switch api {
	case "R":
		pair = rJson["market"].(string)
	case "W":
		pair = rJson["code"].(string)
	}
	var pairInfo = strings.Split(pair, "-")
	var market, symbol = strings.ToLower(pairInfo[0]), strings.ToLower(pairInfo[1])

	tsFloat := int(rJson["timestamp"].(float64))
	ts := commons.FormatTs(strconv.Itoa(tsFloat))
	orderbooks := rJson["orderbook_units"].([]interface{})

	var askSlice, bidSlice []interface{}
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
