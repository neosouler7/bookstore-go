package redismanager

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/tgmanager"

	"github.com/go-redis/redis/v8"
)

var (
	ctx                 = context.Background()
	r                   *redis.Client
	once                sync.Once
	tsMap               map[string]int
	location            *time.Location
	StampMicro          = "Jan _2 15:04:05.000000"
	errGetObTargetPrice = errors.New("[ERROR] getting ob target price")
	errInitRedisClient  = errors.New("[ERROR] connecting redis")
	errSetRedis         = errors.New("[ERROR] set redis")
)

func init() {
	location = commons.SetTimeZone("redis")
}

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
			tgmanager.HandleErr("redis", err, errInitRedisClient)

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
	tgmanager.HandleErr(exchange, err, errGetObTargetPrice)
	bidPrice, err := commons.GetObTargetPrice(targetVolume, bidSlice)
	tgmanager.HandleErr(exchange, err, errGetObTargetPrice)

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
	now := time.Now().In(location).Format(StampMicro)

	// key := fmt.Sprintf("ob-go:%s:%s:%s", ob.exchange, ob.market, ob.symbol)
	key := fmt.Sprintf("ob:%sgo:%s:%s", ob.exchange, ob.market, ob.symbol)
	value := fmt.Sprintf("%s|%s|%s", ob.ts, ob.askPrice, ob.bidPrice)

	prevTs := tsMap[fmt.Sprintf("%s:%s", ob.market, ob.symbol)]
	ts, _ := strconv.ParseInt(ob.ts, 10, 64)
	timeGap := int(ts) - prevTs
	if timeGap > 0 {
		err := client().Set(ctx, key, value, 0).Err()
		tgmanager.HandleErr(ob.exchange, err, errSetRedis)
		tsMap[fmt.Sprintf("%s:%s", ob.market, ob.symbol)] = int(ts)
		fmt.Printf("%s Set %s %s %4dms\n", now, api, key, timeGap)
	} else {
		fmt.Printf("%s >>> %s %s\n", now, api, key)
	}
}
