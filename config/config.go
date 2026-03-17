package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
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
	pairsCache  sync.Map
)

func loadConfig() {
	path, _ := os.Getwd()
	file, err := os.Open(path + "/config/config.json")
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&configCache)
	if err != nil {
		log.Fatalf("Failed to decode config file: %v", err)
	}
}

func getCachedConfig() config {
	cacheOnce.Do(loadConfig)
	return configCache
}

func GetName() string {
	return getCachedConfig().Name
}

func GetRedis() redis {
	return getCachedConfig().Redis
}

func GetTg() tg {
	return getCachedConfig().Tg
}

func GetRateLimit(exchange string) (float64, float64) {
	c := getCachedConfig().RateLimit
	return c["buffer"].(float64), c[exchange].(float64)
}

func GetPairs(exchange string) []string {
	if v, ok := pairsCache.Load(exchange); ok {
		return v.([]string)
	}

	pairsData := getCachedConfig().Pairs
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

	pairsCache.Store(exchange, pairs)
	return pairs
}
