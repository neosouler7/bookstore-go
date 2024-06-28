package redismanager

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/config"
	"github.com/neosouler7/bookstore-go/tgmanager"

	"github.com/go-redis/redis/v8"
)

var (
	ctx        = context.Background()
	rdb        *redis.Client
	once       sync.Once
	syncMap    sync.Map // to escape 'concurrent map read and map write' error
	location   *time.Location
	StampMicro = "Jan _2 15:04:05.000000"
	errMsgCnt  = 0
	errMsg     = ""
)

type orderbook struct {
	exchange string
	market   string
	symbol   string
	askPrice string
	bidPrice string
	ts       string
	bsTs     string
}

func init() {
	location = commons.SetTimeZone("Redis")
}

func client() *redis.Client {
	once.Do(func() {
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
		bsTs:     "",
	}
	return ob
}

func (ob *orderbook) setOrderbook(api string) {
	now := time.Now().In(location).Format(StampMicro)
	obTs, errParseInt := strconv.ParseInt(ob.ts, 10, 64)
	tgmanager.HandleErr(ob.exchange, errParseInt)

	// if init
	prevObTs, ok := syncMap.Load(fmt.Sprintf("%s:%s", ob.market, ob.symbol))
	if !ok {
		fmt.Printf("REDIS init set of %s:%s\n", ob.market, ob.symbol)
		syncMap.Store(fmt.Sprintf("%s:%s", ob.market, ob.symbol), int(obTs))
		// prevObTs = int(obTs)
	} else {
		// get&set ts of bookstore
		bsTsStr := commons.FormatTs(fmt.Sprintf("%d", time.Now().UnixNano()/100000))
		bsTs, errParseInt := strconv.ParseInt(bsTsStr, 10, 64)
		tgmanager.HandleErr(ob.exchange, errParseInt)

		ob.bsTs = bsTsStr

		// set redis
		key := fmt.Sprintf("ob:%s:%s:%s", ob.exchange, ob.market, ob.symbol)
		var value string
		// value := fmt.Sprintf("%s|%s|%s|%s", ob.ts, ob.askPrice, ob.bidPrice, ob.bsTs)
		// value := fmt.Sprintf("%s|%s|%s", ob.ts, ob.askPrice, ob.bidPrice)

		obTsGap := int(obTs) - prevObTs.(int)
		bsTsGap := int(bsTs) - int(obTs)

		// set only when gap of bs / ob within range
		// if bsTsGap > 3*1000 {
		// 	errMsgCnt += 1
		// 	errMsg += fmt.Sprintf("%s:%dms\nbs: %d / ob: %d\n\n", key, bsTsGap, bsTs, obTs)
		// 	if errMsgCnt == 10 {
		// 		tgmanager.SendMsg(fmt.Sprintf("%s%s", "## BS-OB LATENCY ##\n\n", errMsg))
		// 		errMsgCnt = 0
		// 		errMsg = ""
		// 	}

		// 	// fmt.Printf("\npvTs: %d\n", prevObTs)
		// 	// fmt.Printf("obTs: %d\n", obTs)
		// 	// fmt.Printf("bsTs: %d\n", bsTs)

		// } else {
		if obTsGap > 0 { // 거래소별 서버 ts 기준, 최신 호가 저장
			value = fmt.Sprintf("%s|%s|%s|%s", ob.ts, ob.askPrice, ob.bidPrice, ob.bsTs)
			err := client().Set(ctx, key, value, 0).Err()
			tgmanager.HandleErr(ob.exchange, err)

			syncMap.Store(fmt.Sprintf("%s:%s", ob.market, ob.symbol), int(obTs))
			fmt.Printf("%s Set %s %s %4dms %4dms %4s %4s %4s %4s\n", now, api, key, obTsGap, bsTsGap, ob.ts, ob.bsTs, ob.askPrice, ob.bidPrice)

		} else if obTsGap == 0 && bsTsGap > 0 { // 장기간 호가 변동 없을 시
			value = fmt.Sprintf("%s|%s|%s|%s", fmt.Sprint(prevObTs), ob.askPrice, ob.bidPrice, ob.bsTs)
			err := client().Set(ctx, key, value, 0).Err()
			tgmanager.HandleErr(ob.exchange, err)

			syncMap.Store(fmt.Sprintf("%s:%s", ob.market, ob.symbol), prevObTs.(int))
			fmt.Printf("%s Rnw %s %s %4dms %4dms %4s %4s %4s %4s\n", now, api, key, obTsGap, bsTsGap, ob.ts, ob.bsTs, ob.askPrice, ob.bidPrice)

		} else {
			fmt.Printf("%s >>> %s %s %4dms %4dms (obTsGap / bsTsGap)\n", now, api, key, obTsGap, bsTsGap) // 이전의 goroutine이 도달하는 경우 obTsGap 음수값 리턴 가능
		}
	}

	// } else if realTsGap > 800 { // refresh - considering low traded coin, set if allowed time has past
	// 	value2 := fmt.Sprintf("%s|%s|%s", bsTsStr, ob.askPrice, ob.bidPrice)
	// 	err2 := client().Set(ctx, key, value2, 0).Err()
	// 	tgmanager.HandleErr(ob.exchange, err2)

	// 	syncMap.Store(fmt.Sprintf("%s:%s", ob.market, ob.symbol), int(realTs))
	// 	fmt.Printf("%s Ref %s %s %4dms %4s %4s %4s %4s\n", now, api, key, realTsGap, bsTsStr, ob.askPrice, ob.bidPrice, ob.bsTs)

	// 	// logType = "R" // refresh

	// logTs = ob.ts
	// if logType == "R" {
	// 	logTs = realTsStr
	// }

	// if logType == "R" || logType == "N" {
	// 	go alog.TraceLog(
	// 		ob.exchange,
	// 		fmt.Sprintf("%s|%s|%s|%s|%s|%s", logType, logTs, ob.market, ob.symbol, ob.askPrice, ob.bidPrice),
	// 	)
	// }
}
