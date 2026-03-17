package redismanager

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/config"
)

var (
	ctx        = context.Background()
	rdb        *redis.Client
	cOnce      sync.Once
	sOnce      sync.Once
	sMap       *TimestampCache // changed from sync.Map to TimestampCache
	pMap       *TimestampCache // changed from sync.Map to TimestampCache
	location   *time.Location
	StampMicro = "Jan _2 15:04:05.000000"

	// memory monitoring
	memStats     runtime.MemStats
	lastMemCheck time.Time

	targetCache sync.Map // cache for target volume/amount maps, keyed by exchange
)

type orderbook struct {
	exchange     string
	market       string
	symbol       string
	safeAskPrice string
	safeBidPrice string
	bestAskPrice string
	bestBidPrice string
	ts           string
	localTs      string
}

func init() {
	// initialize cache (capacity ~2x number of trading pairs)
	sMap = NewTimestampCache(1000) // max 1000 keys
	pMap = NewTimestampCache(1000) // max 1000 keys

	location = commons.SetTimeZone("Redis")

	// start memory monitoring
	go monitorMemory()
}

// monitorMemory monitors heap memory usage
func monitorMemory() {
	ticker := time.NewTicker(30 * time.Second) // check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			runtime.ReadMemStats(&memStats)

			// log if memory usage is high
			if memStats.Alloc > 100*1024*1024 { // above 100MB
				log.Printf("⚠️  high memory usage: Alloc=%d MB, Sys=%d MB, NumGC=%d",
					memStats.Alloc/1024/1024,
					memStats.Sys/1024/1024,
					memStats.NumGC)

				// also log cache sizes
				log.Printf("📊 cache status: sMap=%d, pMap=%d", sMap.Len(), pMap.Len())
			}
		}
	}
}

func client() *redis.Client {
	cOnce.Do(func() {
		redisConfig := config.GetRedis()
		rdb = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", redisConfig.Host, redisConfig.Port),
			Password: redisConfig.Pwd,
			DB:       redisConfig.Db,
		})

		_, err := rdb.Ping(ctx).Result()
		if err != nil {
			log.Fatalln(err)
		}
	})
	return rdb
}

func getTargetMap(exchange string) map[string]string {
	if v, ok := targetCache.Load(exchange); ok {
		return v.(map[string]string)
	}
	m := commons.GetTargetVolumeOrAmountMap(exchange)
	targetCache.Store(exchange, m)
	return m
}

func PreHandleOrderbook(api, exchange, market, symbol string, askSlice, bidSlice []interface{}, ts string) error {
	ob := newOrderbook(exchange, market, symbol, ts)

	targetVolumeOrAmount := strings.Split(getTargetMap(exchange)[market+":"+symbol], "|")
	safeTarget, bestTarget := targetVolumeOrAmount[0], targetVolumeOrAmount[1]

	// calculate targetPrice by volume (deprecated June 2025 - for fbV1)
	// safeAskPrice, safeBidPrice := commons.GetTargetPriceByVolume(safeTarget, askSlice), commons.GetTargetPriceByVolume(safeTarget, bidSlice)
	// bestAskPrice, bestBidPrice := commons.GetTargetPriceByVolume(bestTarget, askSlice), commons.GetTargetPriceByVolume(bestTarget, bidSlice)
	// fmt.Printf("safeAskPrice2: %s, safeBidPrice2: %s, bestAskPrice2: %s, bestBidPrice2: %s\n", safeAskPrice2, safeBidPrice2, bestAskPrice2, bestBidPrice2)

	// calculate targetPrice by amount (for fbV2)
	safeAskPrice, safeBidPrice := commons.GetTargetPriceByAmount(safeTarget, askSlice), commons.GetTargetPriceByAmount(safeTarget, bidSlice)
	bestAskPrice, bestBidPrice := commons.GetTargetPriceByAmount(bestTarget, askSlice), commons.GetTargetPriceByAmount(bestTarget, bidSlice)

	ob.safeAskPrice, ob.safeBidPrice = safeAskPrice, safeBidPrice
	ob.bestAskPrice, ob.bestBidPrice = bestAskPrice, bestBidPrice

	return ob.setOrderbook(api)
}

func newOrderbook(exchange, market, symbol, ts string) *orderbook {
	ob := &orderbook{
		exchange:     exchange,
		market:       market,
		symbol:       symbol,
		safeAskPrice: "",
		safeBidPrice: "",
		bestAskPrice: "",
		bestBidPrice: "",
		ts:           ts,
		localTs:      "",
	}
	return ob
}

func (ob *orderbook) setOrderbook(api string) error {
	currentTsStr := commons.FormatTs(fmt.Sprintf("%d", time.Now().UTC().UnixNano()/100000))
	currentTs, errParseInt := strconv.ParseInt(currentTsStr, 10, 64)
	if errParseInt != nil {
		return fmt.Errorf("failed to parse current timestamp: %v", errParseInt)
	}

	obTs, errParseInt := strconv.ParseInt(ob.ts, 10, 64)
	if errParseInt != nil {
		return fmt.Errorf("failed to parse orderbook timestamp: %v", errParseInt)
	}

	key := fmt.Sprintf("ob:%s:%s:%s", ob.exchange, ob.market, ob.symbol)

	// manage syncTs to prevent race conditions between goroutines
	prevSyncTsStr := "0"
	if prevSyncTsTemp, ok := sMap.Load(key); ok {
		prevSyncTsStr = prevSyncTsTemp
	}
	prevSyncTs, err := strconv.ParseInt(prevSyncTsStr, 10, 64)
	if err != nil {
		prevSyncTs = 0
	}

	// calculate latency between actual pubTs and current time
	prevPubTsStr := "0"
	if prevPubTsTemp, ok := pMap.Load(key); ok {
		prevPubTsStr = prevPubTsTemp
	}
	prevPubTs, err := strconv.ParseInt(prevPubTsStr, 10, 64)
	if err != nil {
		prevPubTs = 0
	}

	serverLatency := int(obTs - prevSyncTs)     // server ts - prev local ts (only publish if positive)
	localLatency := int(currentTs - prevSyncTs) // current local - prev local (large value means server orderbook is stale)
	actualLatency := int(currentTs - prevPubTs) // current local - prev pub (should be stable for REST; our responsibility to maintain)

	var targetTs string
	// removed WS-primary / REST-secondary architecture (2025/12/21)
	if serverLatency >= 0 {
		if serverLatency == 0 && localLatency > 100 {
			targetTs = currentTsStr // use local time if orderbook hasn't updated for a while
			// sMap.Store(key, targetTs) // publish but do NOT store to sMap — local comparison should use previous ts
		} else {
			targetTs = ob.ts // normal case
			sMap.Store(key, targetTs)
		}

		if localLatency > 0 {
			if err := publish(key, targetTs, ob, serverLatency, localLatency, actualLatency, api); err != nil {
				return fmt.Errorf("failed to publish orderbook: %v", err)
			}
		}
	}

	// switch api {
	// case "W":
	// 	// if serverLatency > 0 { // temporarily removed stale-value discard logic (2025/03/26)
	// 	targetTs = ob.ts
	// 	sMap.Store(key, targetTs)
	// 	if err := publish(key, targetTs, ob, serverLatency, localLatency, actualLatency, api); err != nil {
	// 		return fmt.Errorf("failed to publish websocket orderbook: %v", err)
	// 	}
	// 	// }
	// case "R":
	// 	if serverLatency == 0 && localLatency > 100 {
	// 		targetTs = currentTsStr // use local time if orderbook hasn't updated for a while
	// 		// sMap.Store(key, targetTs) // publish but do NOT store to sMap — local comparison should use previous ts
	// 		if err := publish(key, targetTs, ob, serverLatency, localLatency, actualLatency, api); err != nil {
	// 			return fmt.Errorf("failed to publish rest orderbook: %v", err)
	// 		}
	// 	}
	// }

	sOnce.Do(func() {
		go subscribeCheck(ob.exchange)
	})

	return nil
}

const sampleRate = 1000      // sampling rate (e.g. 1 log per 1000)
const initialLogCount = 1000 // initial log count (first 1000 are always printed)
var logCount int32 = 0

func sampledLog(format string, v ...interface{}) {
	count := atomic.AddInt32(&logCount, 1)

	if count <= initialLogCount {
		fmt.Printf(format, v...)
		return
	}

	if rand.Intn(sampleRate) == 0 {
		fmt.Println("[Log Sampling...]")
		fmt.Printf(format, v...)
	}
}

func publish(key, targetTs string, ob *orderbook, serverLatency, localLatency, actualLatency int, api string) error {
	// uid := ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
	value := fmt.Sprintf("%s|%s|%s|%s|%s", ob.safeAskPrice, ob.bestAskPrice, ob.bestBidPrice, ob.safeBidPrice, targetTs)

	// Redis Pub
	err := client().Publish(ctx, key, value).Err()
	if err != nil {
		log.Fatalln(err)
	}

	// Local logging
	sampledLog("[pub] %s %-15s %s %4dms %4dms %4dms\n", api, key, value, serverLatency, localLatency, actualLatency)

	// // Temp CSV
	// if err := saveToCSV("orderbook_log.csv", key, value); err != nil {
	// 	log.Println("csv save error:", err)
	// }
	return nil
}

func subscribeCheck(exchange string) {
	pairs := config.GetPairs(exchange)
	channels := make([]string, 0)
	for _, pair := range pairs {
		pairInfo := strings.Split(pair, ":")
		channels = append(channels, fmt.Sprintf("ob:%s:%s:%s", exchange, pairInfo[0], pairInfo[1]))
	}

	pubsub := client().Subscribe(ctx, channels...)
	defer pubsub.Close()

	fmt.Printf("[channels] %v\n", channels)

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Fatalln(err)
		}

		// fmt.Printf("[sub] %-15s %s\n", msg.Channel, msg.Payload)

		subTsStr := strings.Split(msg.Payload, "|")[4]
		pMap.Store(msg.Channel, subTsStr)
	}
}

func WriteLastSentTime(exchange string, t time.Time) error {
	key := fmt.Sprintf("error:sent:%s", exchange)
	return client().Set(ctx, key, "1", 2*time.Second).Err()
}

func ReadLastSentTime(exchange string) (time.Time, error) {
	key := fmt.Sprintf("error:sent:%s", exchange)
	exists, err := client().Exists(ctx, key).Result()
	if err != nil {
		return time.Time{}, err
	}
	if exists == 1 {
		return time.Now(), nil
	}
	return time.Time{}, nil
}

func saveToCSV(filename, key, value string) error {
	// split value by '|'
	fields := strings.Split(value, "|")

	// final CSV column order
	record := []string{
		key,       // prepended key
		fields[0], // safeAskPrice
		fields[1], // bestAskPrice
		fields[2], // bestBidPrice
		fields[3], // safeBidPrice
		fields[4], // targetTs
	}

	fmt.Println(record)

	// open file (create if not exists, append otherwise)
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	if err := writer.Write(record); err != nil {
		return err
	}

	return nil
}
