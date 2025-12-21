package binance

import (
	"strconv"
	"strings"
	"time"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/redismanager"
	"github.com/neosouler7/bookstore-go/tgmanager"
)

func SetOrderbook(api string, exchange string, rJson map[string]interface{}) {
	var market, symbol string

	switch api {
	case "R":
		market = rJson["market"].(string)
		symbol = rJson["symbol"].(string)
	case "W":
		pairMap := commons.GetPairMap(exchange)
		pair := strings.Split(rJson["stream"].(string), "@")[0]
		market, symbol = pairMap[pair].(map[string]string)["market"], pairMap[pair].(map[string]string)["symbol"]
	}

	tsFloat := time.Now().UnixNano() / 100000
	ts := commons.FormatTs(strconv.FormatInt(tsFloat, 10))

	var asks, bids []interface{}
	switch api {
	case "R":
		asks, bids = rJson["asks"].([]interface{}), rJson["bids"].([]interface{})
	case "W":
		rData := rJson["data"]
		asks, bids = rData.(map[string]interface{})["asks"].([]interface{}), rData.(map[string]interface{})["bids"].([]interface{})
	}

	depth := commons.Min(len(asks), len(bids))
	askSlice := make([]interface{}, 0, depth)
	bidSlice := make([]interface{}, 0, depth)

	for i := 0; i < depth; i++ {
		askR, bidR := asks[i].([]interface{}), bids[i].([]interface{})
		ask := [2]string{askR[0].(string), askR[1].(string)}
		bid := [2]string{bidR[0].(string), bidR[1].(string)}
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
