package upbit

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

func UpbSetOrderbook(api string, exchange string, rJson map[string]interface{}) error {
	// upb differs "market-symbol" receive form
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
	var targetVolumeMap = commons.GetTargetVolumeMap()
	var targetVolume = targetVolumeMap[market+":"+symbol]

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
