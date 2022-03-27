package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/neosouler7/bookstore-go/tgmanager"
)

var (
	errNotStruct = errors.New("config not struct")
	errNoField   = errors.New("field not found")
)

type config struct {
	Redis     redis
	Tg        tg
	ApiKey    map[string]interface{}
	RateLimit map[string]interface{}
	Pairs     map[string]interface{}
	Pairs2    map[string]interface{}
}

type redis struct {
	Host string
	Port string
	Pwd  string
	Db   int
}

type tg struct {
	Token    string
	Chat_ids []int
}

type apiKey struct {
	Public string
	Secret string
}

func readConfig(obj interface{}, fieldName string) reflect.Value {
	s := reflect.ValueOf(obj).Elem()
	if s.Kind() != reflect.Struct {
		tgmanager.HandleErr("readConfig", errNotStruct)
	}
	f := s.FieldByName(fieldName)
	if !f.IsValid() {
		tgmanager.HandleErr("readConfig", errNoField)
	}
	return f
}

func getConfig(key string) interface{} {
	path, _ := os.Getwd()
	file, _ := os.Open(path + "/config/config.json")
	defer file.Close()

	c := config{}
	tgmanager.HandleErr("getConfig", json.NewDecoder(file).Decode(&c))
	return readConfig(&c, key).Interface()
}

func GetRedis() redis {
	return getConfig("Redis").(redis)
}

func GetTg() tg {
	return getConfig("Tg").(tg)
}

func GetApiKey(exchange string) apiKey {
	c := getConfig("ApiKey").(map[string]interface{})[exchange].(map[string]interface{})
	return apiKey{c["public"].(string), c["secret"].(string)}
}

func GetRateLimit(exchange string) (float64, float64) {
	c := getConfig("RateLimit").(map[string]interface{})
	return c["buffer"].(float64), c[exchange].(float64)
}

// deprecated
// func GetPairs(exchange string) []string {
// 	var pairs []string
// 	for _, p := range getConfig("Pairs").(map[string]interface{})[exchange].([]interface{}) {
// 		pairs = append(pairs, p.(string))
// 	}
// 	return pairs
// }

func GetPairs(exchange string) []string {
	var pairs []string
	for market, symbols := range getConfig("Pairs").(map[string]interface{})[exchange].(map[string]interface{}) {
		for _, symbolInfo := range symbols.([]interface{}) {
			pairs = append(pairs, fmt.Sprintf("%s:%s", market, symbolInfo))
		}
	}
	return pairs
}
