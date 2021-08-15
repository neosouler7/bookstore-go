package redismanager

import (
	"context"
	"errors"
	"fmt"
	"log"
	"neosouler7/bookstore-go/commons"
	"os"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
)

var (
	ctx                = context.Background()
	rdb                *redis.Client
	errInitRedisClient = errors.New("[ERROR] connecting redis")
	errSetRedis        = errors.New("[ERROR] set redis")
	// errExecRedis       = errors.New("[ERROR] execu redis")
)

func init() {
	var redisInfo = commons.ReadConfig("Redis").(map[string]interface{})
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisInfo["host"], redisInfo["port"]),
		Password: redisInfo["pwd"].(string),
		DB:       int(redisInfo["db"].(float64)),
	})

	_, err := rdb.Ping(ctx).Result()
	commons.HandleErr(err, errInitRedisClient)
}

// func GetRedisPipeline() redis.Pipeliner {
// 	pipe := rdb.TxPipeline()
// 	return pipe
// }

// func SetPipeline(exchange string, pipe *redis.Pipeliner, key string, value string) {
// 	err := (*pipe).Set(ctx, key, value, 0).Err()
// 	if err != nil {
// 		log.Fatalln(errSetRedis)
// 	}
// 	fmt.Printf("%s pipeline set\n", exchange)
// }

// func ExecPipeline(exchange string, pipe *redis.Pipeliner) {
// 	_, err := (*pipe).Exec(ctx)
// 	if err != nil {
// 		log.Fatalln(errExecRedis)
// 	}
// 	fmt.Printf("%s pipeline executed\n", exchange)
// }

type orderbook struct {
	exchange string
	market   string
	symbol   string
	askPrice string
	bidPrice string
	ts       string
}

func NewOrderbook(exchange string, market string, symbol string, askPrice string, bidPrice string, ts string) *orderbook {
	ob := orderbook{exchange: exchange, market: market, symbol: symbol, askPrice: askPrice, bidPrice: bidPrice, ts: ts}
	return &ob
}

func SetOrderbook(api string, ob orderbook) error {
	logger := log.New(os.Stdout, " INFO: ", log.LstdFlags|log.Lmicroseconds)

	key := fmt.Sprintf("ob:%s:%s:%s", ob.exchange, ob.market, ob.symbol)
	value := fmt.Sprintf("%s|%s|%s", ob.ts, ob.askPrice, ob.bidPrice)

	prev, _ := rdb.Get(ctx, key).Result()
	if prev == "" { // if no ob stored,
		err := rdb.Set(ctx, key, value, 0).Err()
		commons.HandleErr(err, errSetRedis)

		logger.SetPrefix("DEBUG: ")
		logger.Printf("set %s redis since init!\n", key)
	}

	newTs, _ := strconv.ParseInt(ob.ts, 10, 64)
	prevTs, _ := strconv.ParseInt(strings.Split(prev, "|")[0], 10, 64)
	// currentTs := time.Now().UnixNano() / 100000

	if newTs > prevTs {
		err := rdb.Set(ctx, key, value, 0).Err()
		commons.HandleErr(err, errSetRedis)

		timeGap := newTs - prevTs
		logger.Printf("Set %s %s %4dms\n", api, key, timeGap)
	} else {
		logger.SetPrefix("DEBUG: ")
		logger.Printf(">>> %s %s\n", api, key)
	}
	return nil
}
