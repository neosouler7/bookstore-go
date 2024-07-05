package restmanager

import (
	"errors"
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
	gpx string = "https://api.gopax.co.kr"
	hbk string = "https://api-cloud.huobi.co.kr"
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
		e.endPoint = bmb + fmt.Sprintf("/public/orderbook/%s_%s", strings.ToUpper(symbol), strings.ToUpper(market))
		e.queryString = fmt.Sprintf("count=10")
	case "con":
		e.endPoint = con + "/orderbook/"
		e.queryString = fmt.Sprintf("currency=%s", strings.ToUpper(symbol))
	case "gpx":
		e.endPoint = gpx + fmt.Sprintf("/trading-pairs/%s-%s/book", strings.ToUpper(symbol), strings.ToUpper(market))
		e.queryString = "" // gpx does not use querystring
	case "hbk":
		e.endPoint = hbk + "/market/depth"
		e.queryString = fmt.Sprintf("symbol=%s%s&depth=20&type=step0", strings.ToLower(symbol), strings.ToLower(market))
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

func FastHttpRequest(c chan<- map[string]interface{}, exchange, method, pair string) {
	var pairInfo = strings.Split(pair, ":")
	market, symbol := pairInfo[0], pairInfo[1]
	epqs := &epqs{}
	epqs.getEpqs(exchange, market, symbol)

	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()

	req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI(epqs.endPoint)
	req.URI().SetQueryString(epqs.queryString)

	statusCode, body, err := fastHttpClient().GetTimeout(nil, req.URI().String(), time.Duration(5)*time.Second)
	tgmanager.HandleErr(exchange, err)
	if len(body) == 0 {
		errHttpResponseBody := errors.New("empty response body")
		tgmanager.HandleErr(exchange, errHttpResponseBody)
	}
	if statusCode != fasthttp.StatusOK {
		errHttpResponseStatus := fmt.Errorf("restapi error with status %d", statusCode)
		tgmanager.HandleErr(exchange, errHttpResponseStatus)
	}

	// err := fastHttpClient().Do(req, res)
	// tgmanager.HandleErr(exchange, err)

	// body, statusCode := res.Body(), res.StatusCode()
	// if len(res.Body()) == 0 {
	// 	msg := fmt.Sprintf("%s empty response body\n", exchange)
	// 	tgmanager.HandleErr(msg, errHttpRequest)
	// }
	// if statusCode != fasthttp.StatusOK {
	// 	msg := fmt.Sprintf("%s restapi error with status: %d\n", exchange, statusCode)
	// 	tgmanager.HandleErr(msg, errHttpRequest)
	// }

	switch exchange {
	case "bin":
		var rJson interface{}
		commons.Bytes2Json(body, &rJson)

		// add market, symbol since no value on return
		rJson.(map[string]interface{})["market"] = market
		rJson.(map[string]interface{})["symbol"] = symbol
		c <- rJson.(map[string]interface{})

	case "bif":
		var rJson interface{}
		commons.Bytes2Json(body, &rJson)

		// add market, symbol since no value on return
		rJson.(map[string]interface{})["market"] = market
		rJson.(map[string]interface{})["symbol"] = symbol
		c <- rJson.(map[string]interface{})

	case "bmb":
		var rJson interface{}
		commons.Bytes2Json(body, &rJson)

		c <- rJson.(map[string]interface{})["data"].(map[string]interface{})
	case "con":
		var rJson interface{}
		commons.Bytes2Json(body, &rJson)

		c <- rJson.(map[string]interface{})
	case "gpx":
		var rJson interface{}
		commons.Bytes2Json(body, &rJson)

		// add market, symbol since no value on return
		rJson.(map[string]interface{})["market"] = market
		rJson.(map[string]interface{})["symbol"] = symbol
		c <- rJson.(map[string]interface{})
	case "hbk":
		var rJson interface{}
		commons.Bytes2Json(body, &rJson)

		c <- rJson.(map[string]interface{})
	case "kbt":
		var rJson interface{}
		commons.Bytes2Json(body, &rJson)

		// add market, symbol since no value on return
		rJson.(map[string]interface{})["market"] = market
		rJson.(map[string]interface{})["symbol"] = symbol
		c <- rJson.(map[string]interface{})
	case "upb":
		var rJson []interface{}
		commons.Bytes2Json(body, &rJson)

		c <- rJson[0].(map[string]interface{})
	}
}

func FastHttpRequest2(exchange, method, pair string) map[string]interface{} {
	var pairInfo = strings.Split(pair, ":")
	market, symbol := pairInfo[0], pairInfo[1]
	epqs := &epqs{}
	epqs.getEpqs(exchange, market, symbol)

	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()

	req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI(epqs.endPoint)
	req.URI().SetQueryString(epqs.queryString)

	statusCode, body, err := fastHttpClient().GetTimeout(nil, req.URI().String(), time.Duration(5)*time.Second)
	tgmanager.HandleErr(exchange, err)
	if len(body) == 0 {
		errHttpResponseBody := errors.New("empty response body")
		tgmanager.HandleErr(exchange, errHttpResponseBody)
	}
	if statusCode != fasthttp.StatusOK {
		errHttpResponseStatus := fmt.Errorf("restapi error with status %d", statusCode)
		tgmanager.HandleErr(exchange, errHttpResponseStatus)
	}

	value := make(map[string]interface{})
	switch exchange {
	case "bmb":
		var rJson interface{}
		commons.Bytes2Json(body, &rJson)

		value = rJson.(map[string]interface{})["data"].(map[string]interface{})
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
	}
	return value
}
