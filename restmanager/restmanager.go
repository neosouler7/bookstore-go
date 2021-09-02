package restmanager

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/tgmanager"

	"github.com/valyala/fasthttp"
)

var (
	errHttpRequest = errors.New("http request")
)

const (
	binEndPoint string = "https://api.binance.com"
	bmbEndPoint string = "https://api.bithumb.com"
	conEndPoint string = "https://api.coinone.co.kr"
	gpxEndPoint string = "https://api.gopax.co.kr"
	hbkEndPoint string = "https://api-cloud.huobi.co.kr"
	kbtEndPoint string = "https://api3.korbit.co.kr"
	upbEndPoint string = "https://api.upbit.com"
)

var c *fasthttp.Client
var once sync.Once

func fastHttpClient() *fasthttp.Client {
	if c == nil {
		once.Do(func() {
			clientPointer := &fasthttp.Client{}
			c = clientPointer
		})
	}
	return c
}

func getEndpointQuerystring(exchange string, market string, symbol string) (string, string) {
	var endPoint, queryString string
	switch exchange {
	case "bin":
		endPoint = binEndPoint + "/api/v3/depth"
		queryString = fmt.Sprintf("limit=50&symbol=%s%s", strings.ToUpper(symbol), strings.ToUpper(market))
	case "bmb":
		endPoint = bmbEndPoint + fmt.Sprintf("/public/orderbook/%s_%s", strings.ToUpper(symbol), strings.ToUpper(market))
		queryString = "" // bmb does not use querystring
	case "con":
		endPoint = conEndPoint + "/orderbook/"
		queryString = fmt.Sprintf("currency=%s", strings.ToUpper(symbol))
	case "gpx":
		endPoint = gpxEndPoint + fmt.Sprintf("/trading-pairs/%s-%s/book", strings.ToUpper(symbol), strings.ToUpper(market))
		queryString = "" // gpx does not use querystring
	case "hbk":
		endPoint = hbkEndPoint + "/market/depth"
		queryString = fmt.Sprintf("symbol=%s%s&depth=20&type=step0", strings.ToLower(symbol), strings.ToLower(market))
	case "kbt":
		endPoint = kbtEndPoint + "/v1/orderbook"
		queryString = fmt.Sprintf("currency_pair=%s_%s", strings.ToLower(symbol), strings.ToLower(market))
	case "upb":
		endPoint = upbEndPoint + "/v1/orderbook"
		queryString = fmt.Sprintf("markets=%s-%s", strings.ToUpper(market), strings.ToUpper(symbol))
	}
	return endPoint, queryString
}

func FastHttpRequest(c chan<- map[string]interface{}, exchange string, method string, pair string) {
	var pairInfo = strings.Split(pair, ":")
	market, symbol := pairInfo[0], pairInfo[1]
	endPoint, queryString := getEndpointQuerystring(exchange, market, symbol)

	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()

	req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI(endPoint)
	req.URI().SetQueryString(queryString)

	err := fastHttpClient().Do(req, res)
	tgmanager.HandleErr(exchange, err)

	body, statusCode := res.Body(), res.StatusCode()
	if len(res.Body()) == 0 {
		msg := fmt.Sprintf("%s empty response body\n", exchange)
		tgmanager.HandleErr(msg, errHttpRequest)
	}
	if statusCode != fasthttp.StatusOK {
		msg := fmt.Sprintf("%s restapi error with status: %d\n", exchange, statusCode)
		tgmanager.HandleErr(msg, errHttpRequest)
	}

	switch exchange {
	case "bin":
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
