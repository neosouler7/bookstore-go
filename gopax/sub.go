package gopax

import (
	"fmt"
	"strconv"
	"time"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/redismanager"
)

func SetOrderbook(api string, exchange string, rJson map[string]interface{}) {
	market, symbol := rJson["market"].(string), rJson["symbol"].(string)

	// gopax does not return timestamp on return
	ts := commons.FormatTs(strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10))

	askResponse, bidResponse := rJson["ask"].([]interface{}), rJson["bid"].([]interface{})

	var askSlice, bidSlice []interface{}
	for i := 0; i < commons.Min(len(askResponse), len(bidResponse))-1; i++ {
		ask, bid := [2]string{"", ""}, [2]string{"", ""}
		switch api {
		case "R":
			askR, bidR := askResponse[i].([]interface{}), bidResponse[i].([]interface{})
			ask = [2]string{fmt.Sprintf("%f", askR[1].(float64)), fmt.Sprintf("%f", askR[2].(float64))}
			bid = [2]string{fmt.Sprintf("%f", bidR[1].(float64)), fmt.Sprintf("%f", bidR[2].(float64))}
		case "W":
			askR, bidR := askResponse[i].(map[string]interface{}), bidResponse[i].(map[string]interface{})
			ask = [2]string{fmt.Sprintf("%f", askR["price"].(float64)), fmt.Sprintf("%f", askR["volume"].(float64))}
			bid = [2]string{fmt.Sprintf("%f", bidR["price"].(float64)), fmt.Sprintf("%f", bidR["volume"].(float64))}
		}
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
