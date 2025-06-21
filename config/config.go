package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"sync"
)

var (
	errNotStruct = errors.New("config not struct")
	errNoField   = errors.New("field not found")
)

type config struct {
	Name      string
	Redis     redis
	Tg        tg
	ApiKey    map[string]interface{}
	RateLimit map[string]interface{}
	Pairs     map[string]interface{}
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

var (
	configCache config
	cacheOnce   sync.Once
	cacheMutex  sync.RWMutex
)

func loadConfig() {
	path, _ := os.Getwd()
	file, err := os.Open(path + "/config/config.json")
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer file.Close()

	var c config
	err = json.NewDecoder(file).Decode(&c)
	if err != nil {
		log.Fatalf("Failed to decode config file: %v", err)
	}

	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	configCache = c
}

func getCachedConfig(key string) reflect.Value {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	s := reflect.ValueOf(&configCache).Elem()
	if s.Kind() != reflect.Struct {
		log.Fatalln(errNotStruct)
	}

	f := s.FieldByName(key)
	if !f.IsValid() {
		log.Fatalln(errNoField)
	}
	return f
}

func getConfig(key string) interface{} {
	cacheOnce.Do(loadConfig) // 최초 1회만 실행
	return getCachedConfig(key).Interface()
}

func GetName() string {
	return getConfig("Name").(string)
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
	pairsData := getConfig("Pairs").(map[string]interface{})
	exchangeData, ok := pairsData[exchange].(map[string]interface{})
	if !ok {
		log.Printf("Exchange %s not found in Pairs", exchange)
		return nil
	}

	var pairs []string
	for market, symbols := range exchangeData {
		symbolsList, ok := symbols.([]interface{})
		if !ok {
			log.Printf("Invalid symbols data for market %s in exchange %s", market, exchange)
			continue
		}
		for _, symbol := range symbolsList {
			symbolStr, ok := symbol.(string)
			if !ok {
				log.Printf("Invalid symbol format for %v in market %s", symbol, market)
				continue
			}
			pairs = append(pairs, fmt.Sprintf("%s:%s", market, symbolStr))
		}
	}

	return pairs
}
