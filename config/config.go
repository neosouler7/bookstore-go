package config

import (
	"encoding/json"
	"os"
	"reflect"

	"github.com/neosouler7/bookstore-go/tgmanager"
)

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

type config struct {
	Redis     redis
	Tg        tg
	ApiKey    map[string]interface{}
	RateLimit map[string]interface{}
	Pairs     map[string]interface{}
}

func readConfig(obj interface{}, fieldName string) reflect.Value {
	curStruct := reflect.ValueOf(obj).Elem()
	if curStruct.Kind() != reflect.Struct {
		panic("not struct")
	}
	curField := curStruct.FieldByName(fieldName)
	if !curField.IsValid() {
		panic("not found:" + fieldName)
	}
	return curField
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

func GetPairs(exchange string) []string {
	var after []string
	for _, pair := range getConfig("Pairs").(map[string]interface{})[exchange].([]interface{}) {
		after = append(after, pair.(string))
	}
	return after
}
