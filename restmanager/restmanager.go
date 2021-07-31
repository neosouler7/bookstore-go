package restmanager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/valyala/fasthttp"
)

const UPB_REST string = "https://api.upbit.com"
const CON_REST string = "https://api.coinone.co.kr"

func FastHttpRequest(c chan<- map[string]interface{}, exchange string, method string, pair string) {
	var pairInfo = strings.Split(pair, ":")
	var market = pairInfo[0]
	var symbol = pairInfo[1]
	var endPoint, queryString string

	switch exchange {
	default:
		fmt.Println("proper exchange needed on HttpRequest")
	case "upb":
		endPoint = UPB_REST + "/v1/orderbook"
		queryString = fmt.Sprintf("markets=%s-%s", strings.ToUpper(market), strings.ToUpper(symbol))
	case "con":
		endPoint = CON_REST + "/orderbook/"
		queryString = fmt.Sprintf("currency=%s", strings.ToUpper(symbol))
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
	}
}

// deprecated to fasthttp
func HttpRequest(c chan<- map[string]interface{}, exchange string, method string, pair string) {
	var pairInfo = strings.Split(pair, ":")
	var market = pairInfo[0]
	var symbol = pairInfo[1]
	var endPoint string

	// set end point
	switch exchange {
	default:
		fmt.Println("proper exchange needed on HttpRequest")
	case "upb":
		endPoint = UPB_REST + "/v1/orderbook"
	case "con":
		endPoint = CON_REST + "/orderbook/"
	}

	req, err := http.NewRequest(method, endPoint, nil)
	if err != nil {
		log.Fatalln(err)
	}

	q := req.URL.Query()
	// set query params
	switch exchange {
	default:
		fmt.Println("proper query params needed on HttpRequest")
	case "upb":
		q.Add("markets", strings.ToUpper(market)+"-"+strings.ToUpper(symbol))
	case "con":
		q.Add("currency", strings.ToUpper(symbol))
	}

	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil || res.StatusCode > 400 {
		log.Fatalln("request statusCode:", res.StatusCode)
	}
	defer res.Body.Close()

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln(err)
	}

	switch exchange {
	case "upb":
		var rJson []interface{}
		err = json.Unmarshal(buf, &rJson)
		if err != nil {
			log.Fatalln(err)
		}
		c <- rJson[0].(map[string]interface{})
	case "con":
		var rJson interface{}
		err = json.Unmarshal(buf, &rJson)
		if err != nil {
			log.Fatalln(err)
		}
		c <- rJson.(map[string]interface{})
	}
}
