package commons

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/neosouler7/bookstore-go/tgmanager"
)

type config struct {
	Redis     map[string]interface{}
	Tg        map[string]interface{}
	ApiKey    map[string]interface{}
	RateLimit map[string]interface{}
	Pairs     map[string]interface{}
}

func getAttr(obj interface{}, fieldName string) reflect.Value {
	pointToStruct := reflect.ValueOf(obj)
	curStruct := pointToStruct.Elem()
	if curStruct.Kind() != reflect.Struct {
		panic("not struct")
	}
	curField := curStruct.FieldByName(fieldName)
	if !curField.IsValid() {
		panic("not found:" + fieldName)
	}
	return curField
}

func ReadConfig(key string) interface{} {
	path, _ := os.Getwd()
	file, _ := os.Open(path + "/config/config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)

	c := config{}
	err := decoder.Decode(&c)
	tgmanager.HandleErr("ReadConfig", err)

	return getAttr(&c, key).Interface()
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

func GetTargetVolumeMap(exchange string) map[string]string {
	volumeMap := make(map[string]string)
	pairs := ReadConfig("Pairs").(map[string]interface{})[exchange]
	for _, p := range pairs.([]interface{}) {
		var pairInfo = strings.Split(p.(string), ":")
		market, symbol, targetVolume := pairInfo[0], pairInfo[1], pairInfo[2]

		volumeMap[market+":"+symbol] = targetVolume
	}
	return volumeMap
}

func GetObTargetPrice(volume string, orderbook interface{}) (string, error) {
	/*
		ask's price should go up, and bid should go down

		ask = [[p1, v1], [p2, v2], [p3, v3] ...]
		bid = [[p3, v3], [p2, v2], [p1, p1] ...]
	*/
	//
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
			return obInfo[0], nil
		}
	}
	return obSlice[len(obSlice)-1].([2]string)[0], nil
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

func GetPairMap(exchange string) map[string]interface{} {
	var pairs = ReadConfig("Pairs").(map[string]interface{})[exchange]
	pairInterface := make(map[string]interface{})
	for _, pair := range pairs.([]interface{}) {
		var pairInfo = strings.Split(pair.(string), ":")
		market, symbol := pairInfo[0], pairInfo[1]
		pairInterface[fmt.Sprintf("%s%s", symbol, market)] = map[string]string{"market": market, "symbol": symbol}
	}
	return pairInterface
}
