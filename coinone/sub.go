package coinone

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

func ConSetOrderbook(api string, exchange string, rJson map[string]interface{}) error {
	// con differs "market-symbol" receive form
	var targetVolumeMap = commons.GetTargetVolumeMap()
	var market, symbol, targetVolume, ts string
	var askResponse, bidResponse []interface{}
	var askSlice, bidSlice []interface{}
	switch api {
	case "R":
		market = "krw"
		symbol = rJson["currency"].(string)
		targetVolume = targetVolumeMap[market+":"+symbol]

		tsString := rJson["timestamp"].(string)
		ts = commons.FormatTs(tsString)

		askResponse = rJson["ask"].([]interface{})
		bidResponse = rJson["bid"].([]interface{})
	case "W":
		rTopic := rJson["topic"].(map[string]interface{})
		market = strings.ToLower(rTopic["priceCurrency"].(string))
		symbol = strings.ToLower(rTopic["productCurrency"].(string))
		targetVolume = targetVolumeMap[market+":"+symbol]

		rData := rJson["data"]
		tsFloat := int(rData.(map[string]interface{})["timestamp"].(float64))
		ts = commons.FormatTs(fmt.Sprintf("%d", tsFloat))

		askResponse = rData.(map[string]interface{})["ask"].([]interface{})
		bidResponse = rData.(map[string]interface{})["bid"].([]interface{})
	}

	if len(askResponse) == len(bidResponse) {
		for i := range askResponse {
			askR := askResponse[i]
			bidR := bidResponse[i]
			ask := [2]string{fmt.Sprintf("%f", askR.(map[string]interface{})["price"]), fmt.Sprintf("%f", askR.(map[string]interface{})["qty"])}
			bid := [2]string{fmt.Sprintf("%f", bidR.(map[string]interface{})["price"]), fmt.Sprintf("%f", bidR.(map[string]interface{})["qty"])}
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
