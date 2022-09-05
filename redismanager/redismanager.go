package redismanager

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/neosouler7/bookstore-go/alog"
	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/config"
	"github.com/neosouler7/bookstore-go/tgmanager"

	"github.com/go-redis/redis/v8"
)

var (
	ctx        = context.Background()
	r          *redis.Client
	once       sync.Once
	syncMap    sync.Map // to escape 'concurrent map read and map write' error
	location   *time.Location
	StampMicro = "Jan _2 15:04:05.000000"
)

type orderbook struct {
	exchange string
	market   string
	symbol   string
	askPrice string
	bidPrice string
	ts       string
}

func init() {
	location = commons.SetTimeZone("Redis")
}

func client() *redis.Client {
	once.Do(func() {
		redisConfig := config.GetRedis()
		r = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", redisConfig.Host, redisConfig.Port),
			Password: redisConfig.Pwd,
			DB:       redisConfig.Db,
		})

		_, err := r.Ping(ctx).Result()
		tgmanager.HandleErr("redis", err)
	})
	return r
}

func PreHandleOrderbook(api, exchange, market, symbol string, askSlice, bidSlice []interface{}, ts string) {
	ob := newOrderbook(exchange, market, symbol, ts)

	targetVolume := commons.GetTargetVolumeMap(exchange)[market+":"+symbol]
	askPrice, bidPrice := commons.GetObTargetPrice(targetVolume, askSlice), commons.GetObTargetPrice(targetVolume, bidSlice)
	ob.askPrice, ob.bidPrice = askPrice, bidPrice

	ob.setOrderbook(api)
}

func newOrderbook(exchange, market, symbol, ts string) *orderbook {
	ob := &orderbook{
		exchange: exchange,
		market:   market,
		symbol:   symbol,
		askPrice: "",
		bidPrice: "",
		ts:       ts,
	}
	return ob
}

func (ob *orderbook) setOrderbook(api string) {
	now := time.Now().In(location).Format(StampMicro)

	key := fmt.Sprintf("ob:%s:%s:%s", ob.exchange, ob.market, ob.symbol)
	value := fmt.Sprintf("%s|%s|%s", ob.ts, ob.askPrice, ob.bidPrice)

	ts, errParseInt := strconv.ParseInt(ob.ts, 10, 64)
	tgmanager.HandleErr("redisParseInt", errParseInt)
	prevTs, ok := syncMap.Load(fmt.Sprintf("%s:%s", ob.market, ob.symbol))
	if !ok {
		fmt.Printf("REDIS init set of %s:%s\n", ob.market, ob.symbol)
		syncMap.Store(fmt.Sprintf("%s:%s", ob.market, ob.symbol), int(ts))
		prevTs = int(ts)
	}
	realTsStr := commons.FormatTs(fmt.Sprintf("%d", time.Now().UnixNano()/100000))
	realTs, errParseInt := strconv.ParseInt(realTsStr, 10, 64)
	tgmanager.HandleErr("redisParseInt", errParseInt)

	serverTsGap := int(ts) - prevTs.(int)
	realTsGap := int(realTs) - prevTs.(int)

	var logType string
	var logTs string
	if serverTsGap > 0 {
		err := client().Set(ctx, key, value, 0).Err()
		tgmanager.HandleErr(ob.exchange, err)

		syncMap.Store(fmt.Sprintf("%s:%s", ob.market, ob.symbol), int(ts))
		fmt.Printf("%s Set %s %s %4dms %4s %4s %4s\n", now, api, key, serverTsGap, ob.ts, ob.askPrice, ob.bidPrice)

		logType = "N" // normal

	} else if realTsGap > 800 { // refresh - considering low traded coin, set if allowed time has past
		value2 := fmt.Sprintf("%s|%s|%s", realTsStr, ob.askPrice, ob.bidPrice)
		err2 := client().Set(ctx, key, value2, 0).Err()
		tgmanager.HandleErr(ob.exchange, err2)

		syncMap.Store(fmt.Sprintf("%s:%s", ob.market, ob.symbol), int(realTs))
		fmt.Printf("%s Ref %s %s %4dms %4s %4s %4s\n", now, api, key, realTsGap, realTsStr, ob.askPrice, ob.bidPrice)

		logType = "R" // refresh

	} else {
		fmt.Printf("%s >>> %s %s\n", now, api, key)
	}

	logTs = ob.ts
	if logType == "R" {
		logTs = realTsStr
	}

	if logType == "R" || logType == "N" {
		go alog.TraceLog(
			ob.exchange,
			fmt.Sprintf("%s|%s|%s|%s|%s|%s", logType, logTs, ob.market, ob.symbol, ob.askPrice, ob.bidPrice),
		)
	}
}
