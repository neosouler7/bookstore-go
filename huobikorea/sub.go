package huobikorea

import (
	"errors"
	"fmt"
	"log"
	"neosouler7/bookstore-go/commons"
	"neosouler7/bookstore-go/redismanager"
	"strings"
)

var (
	errGetObTargetPrice = errors.New("[ERROR] getting ob target price")
	errSetOrderbook     = errors.New("[ERROR] setting ob")
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

func SetOrderbook(api string, exchange string, rJson map[string]interface{}) error {
	var pair, market, symbol, ts string
	var pairInterface = getPairInterface(exchange)

	pair = strings.Split(rJson["ch"].(string), ".")[1]
	market = pairInterface[pair].(map[string]string)["market"]
	symbol = pairInterface[pair].(map[string]string)["symbol"]
	ts = commons.FormatTs(fmt.Sprintf("%f", rJson["ts"].(float64)))

	var targetVolumeMap = commons.GetTargetVolumeMap(exchange)
	var targetVolume = targetVolumeMap[market+":"+symbol]

	var askResponse, bidResponse []interface{}
	var askSlice, bidSlice []interface{}

	rData := rJson["tick"]
	askResponse = rData.(map[string]interface{})["asks"].([]interface{})
	bidResponse = rData.(map[string]interface{})["bids"].([]interface{})

	// hbk does not guarantee ask/bid same length
	for i := 0; i < commons.Min(len(askResponse), len(bidResponse)); i++ {
		for i := range askResponse {
			askR := askResponse[i].([]interface{})
			bidR := bidResponse[i].([]interface{})
			ask := [2]string{fmt.Sprintf("%f", askR[0].(float64)), fmt.Sprintf("%f", askR[1].(float64))}
			bid := [2]string{fmt.Sprintf("%f", bidR[0].(float64)), fmt.Sprintf("%f", bidR[1].(float64))}
			askSlice = append(askSlice, ask)
			bidSlice = append(bidSlice, bid)
		}
	}

	askPrice, err := commons.GetObTargetPrice(targetVolume, askSlice)
	if err != nil {
		log.Fatalln(errGetObTargetPrice)
		return err
	}
	bidPrice, err := commons.GetObTargetPrice(targetVolume, bidSlice)
	if err != nil {
		log.Fatalln(errGetObTargetPrice)
		return err
	}

	ob := redismanager.NewOrderbook(exchange, strings.ToLower(market), strings.ToLower(symbol), askPrice, bidPrice, ts)
	err = redismanager.SetOrderbook(api, *ob)
	if err != nil {
		log.Fatalln(errSetOrderbook)
		return err
	}
	return nil
}
