package restmanager

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/tgmanager"

	"github.com/valyala/fasthttp"
)

var (
	c    *fasthttp.Client
	once sync.Once
)

const (
	bin string = "https://api.binance.com"
	bif string = "https://fapi.binance.com"
	bmb string = "https://api.bithumb.com"
	con string = "https://api.coinone.co.kr"
	kbt string = "https://api.korbit.co.kr"
	upb string = "https://api.upbit.com"
)

type epqs struct {
	endPoint    string
	queryString string
}

func (e *epqs) getEpqs(exchange, market, symbol string) {
	switch exchange {
	case "bin":
		e.endPoint = bin + "/api/v3/depth"
		e.queryString = fmt.Sprintf("limit=50&symbol=%s%s", strings.ToUpper(symbol), strings.ToUpper(market))
	case "bif":
		e.endPoint = bif + "/fapi/v1/depth"
		e.queryString = fmt.Sprintf("limit=50&symbol=%s%s", strings.ToUpper(symbol), strings.ToUpper(market))
	case "bmb":
		e.endPoint = bmb + "/v1/orderbook"
		e.queryString = fmt.Sprintf("markets=%s-%s", strings.ToUpper(market), strings.ToUpper(symbol))
	case "con":
		e.endPoint = con + fmt.Sprintf("/public/v2/orderbook/%s/%s", strings.ToUpper(market), strings.ToUpper(symbol))
		e.queryString = ""
	case "kbt":
		// e.endPoint = kbt + "/v1/orderbook"
		// e.queryString = fmt.Sprintf("currency_pair=%s_%s", strings.ToLower(symbol), strings.ToLower(market))
		e.endPoint = kbt + "/v2/orderbook"
		e.queryString = fmt.Sprintf("symbol=%s_%s", strings.ToLower(symbol), strings.ToLower(market))
	case "upb":
		e.endPoint = upb + "/v1/orderbook"
		e.queryString = fmt.Sprintf("markets=%s-%s", strings.ToUpper(market), strings.ToUpper(symbol))
	}
}

func fastHttpClient() *fasthttp.Client {
	once.Do(func() {
		clientPointer := &fasthttp.Client{}
		c = clientPointer
	})
	return c
}

func FastHttpRequest(exchange, method, pair string) map[string]interface{} {
	var pairInfo = strings.Split(pair, ":")
	market, symbol := pairInfo[0], pairInfo[1]
	epqs := &epqs{}
	epqs.getEpqs(exchange, market, symbol)

	url := epqs.endPoint
	if epqs.queryString != "" {
		url += "?" + epqs.queryString
	}
	statusCode, body, err := fastHttpClient().GetTimeout(nil, url, time.Duration(5)*time.Second)
	if err != nil {
		errRestResult := fmt.Errorf("%s|%s HTTP failed for %w", market, symbol, err)
		tgmanager.HandleErr(exchange, errRestResult)
	}
	if len(body) == 0 {
		errHttpResponseBody := fmt.Errorf("%s|%s HTTP empty response body", market, symbol)
		tgmanager.HandleErr(exchange, errHttpResponseBody)
	}
	if statusCode != fasthttp.StatusOK {
		errHttpResponseStatus := fmt.Errorf("%s|%s HTTP with status %d", market, symbol, statusCode)
		tgmanager.HandleErr(exchange, errHttpResponseStatus)
	}

	var value map[string]interface{}

	switch exchange {
	case "bin", "bif":
		var rJson interface{}
		commons.Bytes2Json(body, &rJson)

		// add market, symbol since no value on return
		rJson.(map[string]interface{})["market"] = market
		rJson.(map[string]interface{})["symbol"] = symbol
		value = rJson.(map[string]interface{})

	case "bmb":
		var rJson []interface{}
		commons.Bytes2Json(body, &rJson)

		value = rJson[0].(map[string]interface{})

	case "kbt":
		var rJson interface{}
		commons.Bytes2Json(body, &rJson)

		// add market, symbol since no value on return
		rJson.(map[string]interface{})["market"] = market
		rJson.(map[string]interface{})["symbol"] = symbol

		value = rJson.(map[string]interface{})

	case "upb":
		var rJson []interface{}
		commons.Bytes2Json(body, &rJson)

		value = rJson[0].(map[string]interface{})

	case "con":
		var rJson interface{}
		commons.Bytes2Json(body, &rJson)

		value = rJson.(map[string]interface{})
	}

	return value
}
