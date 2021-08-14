package restmanager

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/valyala/fasthttp"
)

const (
	upbEndPoint string = "https://api.upbit.com"
	conEndPoint string = "https://api.coinone.co.kr"
	binEndPoint string = "https://api.binance.com"
	kbtEndPoint string = "https://api.korbit.co.kr"
	hbkEndPoint string = "https://api-cloud.huobi.co.kr"
)

func FastHttpRequest(c chan<- map[string]interface{}, exchange string, method string, pair string) {
	var pairInfo = strings.Split(pair, ":")
	var market = pairInfo[0]
	var symbol = pairInfo[1]
	var endPoint, queryString string

	switch exchange {
	default:
		fmt.Println("proper exchange needed on FastHttpRequest")
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
	}

	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()

	req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI(endPoint)
	req.URI().SetQueryString(queryString)

	client := &fasthttp.Client{}
	err := client.Do(req, res)
	if err != nil {
		fmt.Println(req)
		fmt.Println(res)
		log.Fatal(err)
	}
	if len(res.Body()) == 0 {
		log.Fatal("missing request body")
	}

	body := res.Body()
	switch exchange {
	case "upb":
		var rJson []interface{}
		err = json.Unmarshal(body, &rJson)
		if err != nil {
			log.Fatalln(err)
		}
		c <- rJson[0].(map[string]interface{})
	case "con":
		var rJson interface{}
		err = json.Unmarshal(body, &rJson)
		if err != nil {
			log.Fatalln(err)
		}
		c <- rJson.(map[string]interface{})
	case "bin":
		var rJson interface{}
		err = json.Unmarshal(body, &rJson)
		if err != nil {
			log.Fatalln(err)
		}
		// add market, symbol since no value on return
		rJson.(map[string]interface{})["market"] = market
		rJson.(map[string]interface{})["symbol"] = symbol
		c <- rJson.(map[string]interface{})
	case "kbt":
		var rJson interface{}
		err = json.Unmarshal(body, &rJson)
		if err != nil {
			log.Fatalln(err)
		}
		// add market, symbol since no value on return
		rJson.(map[string]interface{})["market"] = market
		rJson.(map[string]interface{})["symbol"] = symbol
		c <- rJson.(map[string]interface{})
	case "hbk":
		var rJson interface{}
		err = json.Unmarshal(body, &rJson)
		if err != nil {
			log.Fatalln(err)
		}
		c <- rJson.(map[string]interface{})
	}
}

// deprecated to fasthttp
// func HttpRequest(c chan<- map[string]interface{}, exchange string, method string, pair string) {
// 	var pairInfo = strings.Split(pair, ":")
// 	var market = pairInfo[0]
// 	var symbol = pairInfo[1]
// 	var endPoint string

// 	// set end point
// 	switch exchange {
// 	default:
// 		fmt.Println("proper exchange needed on HttpRequest")
// 	case "upb":
// 		endPoint = upbEndPoint + "/v1/orderbook"
// 	case "con":
// 		endPoint = conEndPoint + "/orderbook/"
// 	}

// 	req, err := http.NewRequest(method, endPoint, nil)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	q := req.URL.Query()
// 	// set query params
// 	switch exchange {
// 	default:
// 		fmt.Println("proper query params needed on HttpRequest")
// 	case "upb":
// 		q.Add("markets", strings.ToUpper(market)+"-"+strings.ToUpper(symbol))
// 	case "con":
// 		q.Add("currency", strings.ToUpper(symbol))
// 	}

// 	req.URL.RawQuery = q.Encode()

// 	client := &http.Client{}
// 	res, err := client.Do(req)
// 	if err != nil || res.StatusCode > 400 {
// 		log.Fatalln("request statusCode:", res.StatusCode)
// 	}
// 	defer res.Body.Close()

// 	buf, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	switch exchange {
// 	case "upb":
// 		var rJson []interface{}
// 		err = json.Unmarshal(buf, &rJson)
// 		if err != nil {
// 			log.Fatalln(err)
// 		}
// 		c <- rJson[0].(map[string]interface{})
// 	case "con":
// 		var rJson interface{}
// 		err = json.Unmarshal(buf, &rJson)
// 		if err != nil {
// 			log.Fatalln(err)
// 		}
// 		c <- rJson.(map[string]interface{})
// 	}
// }
