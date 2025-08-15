package upbit

import (
	"strconv"
	"strings"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/redismanager"
	"github.com/neosouler7/bookstore-go/tgmanager"
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

	askSlice := make([]interface{}, 0, len(orderbooks)) // 용량 미리 할당
	bidSlice := make([]interface{}, 0, len(orderbooks)) // 용량 미리 할당

	for _, ob := range orderbooks {
		o := ob.(map[string]interface{})
		ap, as := o["ask_price"].(float64), o["ask_size"].(float64)
		bp, bs := o["bid_price"].(float64), o["bid_size"].(float64)

		ask := [2]string{
			strconv.FormatFloat(ap, 'f', -1, 64),
			strconv.FormatFloat(as, 'f', -1, 64),
		}
		bid := [2]string{
			strconv.FormatFloat(bp, 'f', -1, 64),
			strconv.FormatFloat(bs, 'f', -1, 64),
		}
		askSlice = append(askSlice, ask)
		bidSlice = append(bidSlice, bid)
	}

	if err := redismanager.PreHandleOrderbook(
		api,
		exchange,
		market,
		symbol,
		askSlice,
		bidSlice,
		ts,
	); err != nil {
		tgmanager.HandleErr(exchange, err)
	}
}
