package commons

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	errDecode = errors.New("[ERROR] decoding")
	errEncode = errors.New("[ERROR] encoding")
)

type config struct {
	Redis map[string]interface{}
	Tg    map[string]interface{}
	Pairs map[string]interface{}
}

func getAttr(obj interface{}, fieldName string) reflect.Value {
	pointToStruct := reflect.ValueOf(obj)
	curStruct := pointToStruct.Elem()
	if curStruct.Kind() != reflect.Struct {
		panic("not struct")
	}
	curField := curStruct.FieldByName(fieldName) // type: reflect.Value
	if !curField.IsValid() {
		panic("not found:" + fieldName)
	}
	return curField
}

func ReadConfig(key string) interface{} {
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)

	c := config{}
	err := decoder.Decode(&c)
	if err != nil {
		fmt.Println("error:", err)
	}
	return getAttr(&c, key).Interface()
}

func FormatTs(ts string) string {
	if len(ts) <= 13 {
		add := strings.Repeat("0", 13-len(ts))
		return fmt.Sprintf("%s%s", ts, add)
	} else if len(ts) > 13 {
		return ts[:13]
	} else {
		return ts
	}
}

func GetTargetVolumeMap(exchange string) map[string]string {
	volumeMap := make(map[string]string)

	pairs := ReadConfig("Pairs").(map[string]interface{})[exchange]
	for _, p := range pairs.([]interface{}) {
		var pairInfo = strings.Split(p.(string), ":")
		var market = pairInfo[0]
		var symbol = pairInfo[1]
		var targetVolume = pairInfo[2]

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
	targetVolume, _ := strconv.ParseFloat(volume, 64)

	obSlice := orderbook.([]interface{})
	for _, ob := range obSlice {
		obInfo := ob.([2]string)
		volume, _ := strconv.ParseFloat(obInfo[1], 64)
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

// get err as returned error, and log with specific errMsg
func HandleErr(err error, errMsg error) {
	if err != nil {
		log.Fatalln(errMsg)
	}
}

func Bytes2Json(data []byte, i interface{}) {
	r := bytes.NewReader(data)
	err := json.NewDecoder(r).Decode(i)
	HandleErr(err, errDecode)
}

func SetTimeZone(name string) *time.Location {
	tz := os.Getenv("TZ")
	if tz == "" {
		tz = "Asia/Seoul"
		fmt.Printf("%s follows default %s\n", name, tz)
	} else {
		fmt.Printf("%s follows SERVER %s\n", name, tz)
	}
	location, _ := time.LoadLocation(tz)
	return location
}
