package redismanager

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/neosouler7/bookstore-go/commons"

	"github.com/go-redis/redis/v8"
)

var (
	ctx                 = context.Background()
	r                   *redis.Client
	once                sync.Once
	tsMap               map[string]int
	errGetObTargetPrice = errors.New("[ERROR] getting ob target price")
	errInitRedisClient  = errors.New("[ERROR] connecting redis")
	errSetRedis         = errors.New("[ERROR] set redis")
)

func client() *redis.Client {
	if r == nil {
		once.Do(func() {
			var redisInfo = commons.ReadConfig("Redis").(map[string]interface{})
			r = redis.NewClient(&redis.Options{
				Addr:     fmt.Sprintf("%s:%s", redisInfo["host"], redisInfo["port"]),
				Password: redisInfo["pwd"].(string),
				DB:       int(redisInfo["db"].(float64)),
			})

			_, err := r.Ping(ctx).Result()
			commons.HandleErr(err, errInitRedisClient)

			tsMap = make(map[string]int)
		})
	}
	return r
}

type orderbook struct {
	exchange string
	market   string
	symbol   string
	askPrice string
	bidPrice string
	ts       string
}

func PreHandleOrderbook(api string, exchange string, market string, symbol string, askSlice []interface{}, bidSlice []interface{}, ts string) {
	var targetVolumeMap = commons.GetTargetVolumeMap(exchange)
	var targetVolume = targetVolumeMap[market+":"+symbol]

	askPrice, err := commons.GetObTargetPrice(targetVolume, askSlice)
	commons.HandleErr(err, errGetObTargetPrice)
	bidPrice, err := commons.GetObTargetPrice(targetVolume, bidSlice)
	commons.HandleErr(err, errGetObTargetPrice)

	ob := newOrderbook(exchange, market, symbol, askPrice, bidPrice, ts)
	ob.setOrderbook(api)
}

func newOrderbook(exchange string, market string, symbol string, askPrice string, bidPrice string, ts string) *orderbook {
	ob := &orderbook{
		exchange: exchange,
		market:   market,
		symbol:   symbol,
		askPrice: askPrice,
		bidPrice: bidPrice,
		ts:       ts,
	}
	return ob
}

func (ob *orderbook) setOrderbook(api string) {
	logger := log.New(os.Stdout, " INFO: ", log.LstdFlags|log.Lmicroseconds)

	key := fmt.Sprintf("ob-go:%s:%s:%s", ob.exchange, ob.market, ob.symbol)
	value := fmt.Sprintf("%s|%s|%s", ob.ts, ob.askPrice, ob.bidPrice)

	prevTs := tsMap[fmt.Sprintf("%s:%s", ob.market, ob.symbol)]
	ts, _ := strconv.ParseInt(ob.ts, 10, 64)
	timeGap := int(ts) - prevTs
	if timeGap > 0 {
		err := client().Set(ctx, key, value, 0).Err()
		commons.HandleErr(err, errSetRedis)
		tsMap[fmt.Sprintf("%s:%s", ob.market, ob.symbol)] = int(ts)
		logger.Printf("Set %s %s %4dms\n", api, key, timeGap)
	} else {
		logger.SetPrefix("DEBUG: ")
		logger.Printf(">>> %s %s\n", api, key)
	}
}
