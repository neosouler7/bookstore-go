package redismanager

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/config"
	"github.com/neosouler7/bookstore-go/tgmanager"
)

var (
	ctx        = context.Background()
	rdb        *redis.Client
	cOnce      sync.Once
	sOnce      sync.Once
	sMap       sync.Map
	pMap       sync.Map
	location   *time.Location
	StampMicro = "Jan _2 15:04:05.000000"
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
	location = commons.SetTimeZone("Redis")
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
		tgmanager.HandleErr("redis", err)
	})
	return rdb
}

func PreHandleOrderbook(api, exchange, market, symbol string, askSlice, bidSlice []interface{}, ts string) {
	ob := newOrderbook(exchange, market, symbol, ts)

	targetVolume := strings.Split(commons.GetTargetVolumeMap(exchange)[market+":"+symbol], "|")
	safeTargetVolume, bestTargetVolume := targetVolume[0], targetVolume[1]

	safeAskPrice, safeBidPrice := commons.GetObTargetPrice(safeTargetVolume, askSlice), commons.GetObTargetPrice(safeTargetVolume, bidSlice)
	bestAskPrice, bestBidPrice := commons.GetObTargetPrice(bestTargetVolume, askSlice), commons.GetObTargetPrice(bestTargetVolume, bidSlice)

	ob.safeAskPrice, ob.safeBidPrice = safeAskPrice, safeBidPrice
	ob.bestAskPrice, ob.bestBidPrice = bestAskPrice, bestBidPrice

	ob.setOrderbook(api)
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

func (ob *orderbook) setOrderbook(api string) {
	currentTsStr := commons.FormatTs(fmt.Sprintf("%d", time.Now().UTC().UnixNano()/100000))
	currentTs, errParseInt := strconv.ParseInt(currentTsStr, 10, 64)
	tgmanager.HandleErr(ob.exchange, errParseInt)

	obTs, errParseInt := strconv.ParseInt(ob.ts, 10, 64)
	tgmanager.HandleErr(ob.exchange, errParseInt)

	key := fmt.Sprintf("ob:%s:%s:%s", ob.exchange, ob.market, ob.symbol)

	// 내부 goroutine의 race issue 방지 위해 syncTs를 관리
	prevSyncTsStr := "0"
	if prevSyncTsTemp, ok := sMap.Load(key); ok {
		if str, ok := prevSyncTsTemp.(string); ok {
			prevSyncTsStr = str
		}
	}
	prevSyncTs, err := strconv.ParseInt(prevSyncTsStr, 10, 64)
	if err != nil {
		prevSyncTs = 0
	}

	// 실제 pubTs와 현재 간의 Latency를 계산하기 위해
	prevPubTsStr := "0"
	if prevPubTsTemp, ok := pMap.Load(key); ok {
		if str, ok := prevPubTsTemp.(string); ok {
			prevPubTsStr = str
		}
	}
	prevPubTs, err := strconv.ParseInt(prevPubTsStr, 10, 64)
	if err != nil {
		prevPubTs = 0
	}

	serverLatency := int(obTs - prevSyncTs)     // 서버 - 이전 로컬 저장 (양수일 경우만 pub)
	localLatency := int(currentTs - prevSyncTs) // 현재 로컬 - 이전 로컬 저장 (값이 커지면 서버의 호가 갱신이 되지 않고 있다는 뜻)
	actualLatency := int(currentTs - prevPubTs) // 현재 로컬 - 이전 pub (R의 경우 현재 ts를 pub하므로 일정해야함. 내 관리 포인트)

	var targetTs string
	switch api {
	case "W":
		if serverLatency > 0 {
			targetTs = ob.ts
			sMap.Store(key, targetTs)
			publish(key, targetTs, ob, serverLatency, localLatency, actualLatency, api)
		}
	case "R":
		if serverLatency == 0 && localLatency > 100 {
			targetTs = currentTsStr
			// sMap.Store(key, targetTs) // pub은 하지만 로컬 비교를 위한 map에는 저장하지 않는 것이 맞음
			publish(key, targetTs, ob, serverLatency, localLatency, actualLatency, api)
		}
	}

	sOnce.Do(func() {
		subscribeCheck(ob.exchange)
	})
}

const sampleRate = 1000    // 샘플링 비율 (예: 1000개의 로그 중 1개만 출력)
const initialLogCount = 10 // 초기 출력 카운트 (처음 10개는 무조건 출력)
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

func publish(key, targetTs string, ob *orderbook, serverLatency, localLatency, actualLatency int, api string) {
	value := fmt.Sprintf("%s|%s|%s|%s|%s", ob.safeAskPrice, ob.bestAskPrice, ob.bestBidPrice, ob.safeBidPrice, targetTs)

	err := client().Publish(ctx, key, value).Err()
	tgmanager.HandleErr(ob.exchange, err)

	sampledLog("[pub] %s %-15s %s %4dms %4dms %4dms\n", api, key, value, serverLatency, localLatency, actualLatency)
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
			log.Fatalln("Error receiving message:", err)
		}
		// fmt.Printf("[sub] %-15s %s\n", msg.Channel, msg.Payload)

		subTsStr := strings.Split(msg.Payload, "|")[4]
		pMap.Store(msg.Channel, subTsStr)
	}
}
