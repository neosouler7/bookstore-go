package restmanager

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/neosouler7/bookstore-go/commons"

	"github.com/valyala/fasthttp"
)

var (
	errHttpRequest = errors.New("[ERROR] http request")
)

const (
	upbEndPoint string = "https://api.upbit.com"
	conEndPoint string = "https://api.coinone.co.kr"
	binEndPoint string = "https://api.binance.com"
	kbtEndPoint string = "https://api.korbit.co.kr"
	hbkEndPoint string = "https://api-cloud.huobi.co.kr"
	bmbEndPoint string = "https://api.bithumb.com"
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
	case "upb":
		endPoint = upbEndPoint + "/v1/orderbook"
		queryString = fmt.Sprintf("markets=%s-%s", strings.ToUpper(market), strings.ToUpper(symbol))
	case "con":
		endPoint = conEndPoint + "/orderbook/"
		queryString = fmt.Sprintf("currency=%s", strings.ToUpper(symbol))
	case "bin":
		endPoint = binEndPoint + "/api/v3/depth"
		queryString = fmt.Sprintf("limit=50&symbol=%s%s", strings.ToUpper(symbol), strings.ToUpper(market))
	case "kbt":
		endPoint = kbtEndPoint + "/v1/orderbook"
		queryString = fmt.Sprintf("currency_pair=%s_%s", strings.ToLower(symbol), strings.ToLower(market))
	case "hbk":
		endPoint = hbkEndPoint + "/market/depth"
		queryString = fmt.Sprintf("symbol=%s%s&depth=20&type=step0", strings.ToLower(symbol), strings.ToLower(market))
	case "bmb":
		endPoint = bmbEndPoint + "/public/orderbook/"
		queryString = fmt.Sprintf("%s_%s", strings.ToUpper(symbol), strings.ToUpper(market))
	}
	return endPoint, queryString
}

func FastHttpRequest(c chan<- map[string]interface{}, exchange string, method string, pair string) {
	var pairInfo = strings.Split(pair, ":")
	var market = pairInfo[0]
	var symbol = pairInfo[1]
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
	commons.HandleErr(err, errHttpRequest)

	body, statusCode := res.Body(), res.StatusCode()
	switch statusCode {
	case 200:
		switch exchange {
		case "upb":
			var rJson []interface{}
			commons.Bytes2Json(body, &rJson)

			c <- rJson[0].(map[string]interface{})
		case "con":
			var rJson interface{}
			commons.Bytes2Json(body, &rJson)

			c <- rJson.(map[string]interface{})
		case "bin":
			var rJson interface{}
			commons.Bytes2Json(body, &rJson)

			// add market, symbol since no value on return
			rJson.(map[string]interface{})["market"] = market
			rJson.(map[string]interface{})["symbol"] = symbol
			c <- rJson.(map[string]interface{})
		case "kbt":
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
		case "bmb":
			var rJson interface{}
			commons.Bytes2Json(body, &rJson)

			c <- rJson.(map[string]interface{})["data"].(map[string]interface{})
		}
	default:
		fmt.Printf("%s restapi error with status: %d\n", exchange, statusCode)
	}

}
