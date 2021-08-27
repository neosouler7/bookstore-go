package websocketmanager

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/tgmanager"
)

var (
	errGetConn = errors.New("[ERROR] connecting ws")
	errSendMsg = errors.New("[ERROR] sending msg on ws")
	ErrReadMsg = errors.New("[ERROR] reading msg on ws")
)

const (
	binEndPoint string = "stream.binance.com:9443"
	bmbEndPoint string = "pubwss.bithumb.com"
	conEndPoint string = "public-ws-api.coinone.co.kr"
	gpxEndPoint string = "wsapi.gopax.co.kr"
	hbkEndPoint string = "api-cloud.huobi.co.kr"
	kbtEndPoint string = "ws.korbit.co.kr"
	upbEndPoint string = "api.upbit.com"
)

var w *websocket.Conn
var once sync.Once

func Conn(exchange string) *websocket.Conn {
	if w == nil {
		once.Do(func() {
			host, path := getHostPath(exchange)
			u := url.URL{Scheme: "wss", Host: host, Path: path}
			if exchange == "gpx" {
				keyMap := commons.ReadConfig("ApiKey").(map[string]interface{})[exchange]
				publicKey := keyMap.(map[string]interface{})["public"].(string)
				secretKey := keyMap.(map[string]interface{})["secret"].(string)

				ts := commons.FormatTs(fmt.Sprintf("%d", time.Now().UnixNano()/100000))
				key, _ := base64.StdEncoding.DecodeString(secretKey)

				h := hmac.New(sha512.New, []byte(key))
				h.Write([]byte(fmt.Sprintf("t%s", ts)))
				signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

				params := url.Values{}
				params.Set("apiKey", publicKey)
				params.Set("timestamp", ts)
				params.Set("signature", signature)
				u.RawQuery = params.Encode()
			}

			wPointer, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			tgmanager.HandleErr(exchange, err)
			w = wPointer
		})
	}
	return w
}

func getHostPath(exchange string) (string, string) {
	var host, path string
	switch exchange {
	case "bin":
		host = binEndPoint
		path = "/stream"
	case "bmb":
		host = bmbEndPoint
		path = "/pub/ws"
	case "con":
		host = conEndPoint
		path = ""
	case "gpx":
		host = gpxEndPoint
		path = ""
	case "hbk":
		host = hbkEndPoint
		path = "/ws"
	case "kbt":
		host = kbtEndPoint
		path = "/v1/user/push"
	case "upb":
		host = upbEndPoint
		path = "/websocket/v1"
	}
	return host, path
}

func SendMsg(exchange string, msg string) {
	err := Conn(exchange).WriteMessage(websocket.TextMessage, []byte(msg))
	tgmanager.HandleErr(exchange, err)
}

func Pong(exchange string) {
	err := Conn(exchange).WriteMessage(websocket.PongMessage, []byte{})
	tgmanager.HandleErr(exchange, err)
}
