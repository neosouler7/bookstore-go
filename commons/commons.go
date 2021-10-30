package commons

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/neosouler7/bookstore-go/config"
	"github.com/neosouler7/bookstore-go/tgmanager"
)

func GetObTargetPrice(volume string, orderbook interface{}) string {
	/*
		ask's price should go up, and bid should go down

		ask = [[p1, v1], [p2, v2], [p3, v3] ...]
		bid = [[p3, v3], [p2, v2], [p1, p1] ...]
	*/
	currentVolume := 0.0
	targetVolume, err := strconv.ParseFloat(volume, 64)
	tgmanager.HandleErr("GetObTargetPrice", err)

	obSlice := orderbook.([]interface{})
	for _, ob := range obSlice {
		obInfo := ob.([2]string)
		volume, err := strconv.ParseFloat(obInfo[1], 64)
		tgmanager.HandleErr("GetObTargetPrice", err)

		currentVolume += volume
		if currentVolume >= targetVolume {
			return obInfo[0]
		}
	}
	return obSlice[len(obSlice)-1].([2]string)[0]
}

func GetTargetVolumeMap(exchange string) map[string]string {
	m := make(map[string]string)
	for _, p := range config.GetPairs(exchange) {
		var pairInfo = strings.Split(p, ":")
		market, symbol, targetVolume := pairInfo[0], pairInfo[1], pairInfo[2]
		m[market+":"+symbol] = targetVolume
	}
	return m
}

func GetPairMap(exchange string) map[string]interface{} {
	m := make(map[string]interface{})
	for _, pair := range config.GetPairs(exchange) {
		var pairInfo = strings.Split(pair, ":")
		market, symbol := pairInfo[0], pairInfo[1]
		m[fmt.Sprintf("%s%s", symbol, market)] = map[string]string{"market": market, "symbol": symbol}
	}
	return m
}

func FormatTs(ts string) string {
	if len(ts) < 13 {
		add := strings.Repeat("0", 13-len(ts))
		return fmt.Sprintf("%s%s", ts, add)
	} else if len(ts) == 13 { // if millisecond
		return ts
	} else {
		return ts[:13]
		// tm, err := strconv.ParseInt(ts[:13], 10, 64)
		// if err != nil {
		// 	log.Fatalln(err)
		// }
		// convertedTime := time.Unix(0, tm*int64(time.Millisecond))
		// return fmt.Sprintf("%d", convertedTime.UnixMilli()) // to millisecond
	}
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Bytes2Json(data []byte, i interface{}) {
	r := bytes.NewReader(data)
	err := json.NewDecoder(r).Decode(i)
	tgmanager.HandleErr("Bytes2Json", err)
}

func SetTimeZone(name string) *time.Location {
	tz := os.Getenv("TZ")
	if tz == "" {
		tz = "Asia/Seoul"
		fmt.Printf("%s : DEFAULT %s\n", name, tz)
	} else {
		fmt.Printf("%s : SERVER %s\n", name, tz)
	}
	location, _ := time.LoadLocation(tz)
	return location
}
