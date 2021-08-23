package websocketmanager

import (
	"errors"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/neosouler7/bookstore-go/tgmanager"
)

var (
	errGetConn = errors.New("[ERROR] connecting ws")
	errSendMsg = errors.New("[ERROR] sending msg on ws")
	ErrReadMsg = errors.New("[ERROR] reading msg on ws")
)

const (
	upbEndPoint string = "api.upbit.com"
	conEndPoint string = "public-ws-api.coinone.co.kr"
	binEndPoint string = "stream.binance.com:9443"
	kbtEndPoint string = "ws.korbit.co.kr"
	hbkEndPoint string = "api-cloud.huobi.co.kr"
	bmbEndPoint string = "pubwss.bithumb.com"
)

var w *websocket.Conn
var once sync.Once

func Conn(exchange string) *websocket.Conn {
	if w == nil {
		once.Do(func() {
			host, path := getHostPath(exchange)
			u := url.URL{Scheme: "wss", Host: host, Path: path}
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
	case "upb":
		host = upbEndPoint
		path = "/websocket/v1"
	case "con":
		host = conEndPoint
		path = ""
	case "bin":
		host = binEndPoint
		path = "/stream"
	case "kbt":
		host = kbtEndPoint
		path = "/v1/user/push"
	case "hbk":
		host = hbkEndPoint
		path = "/ws"
	case "bmb":
		host = bmbEndPoint
		path = "/pub/ws"
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
