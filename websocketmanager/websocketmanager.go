package websocketmanager

import (
	"errors"
	"neosouler7/bookstore-go/commons"
	"net/url"

	"github.com/gorilla/websocket"
)

var (
	errGetConn = errors.New("[ERROR] connecting ws")
	errSendMsg = errors.New("[ERROR] sending msg on ws")
)

const (
	upbEndPoint string = "api.upbit.com"
	conEndPoint string = "public-ws-api.coinone.co.kr"
	binEndPoint string = "stream.binance.com:9443"
	kbtEndPoint string = "ws.korbit.co.kr"
	hbkEndPoint string = "api-cloud.huobi.co.kr"
)

var w *websocket.Conn

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
	}
	return host, path
}

func GetConn(exchange string) *websocket.Conn {
	if w == nil {
		host, path := getHostPath(exchange)
		u := url.URL{Scheme: "wss", Host: host, Path: path}
		wPointer, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		commons.HandleErr(err, errGetConn)
		w = wPointer
	}
	return w
}

func SendMsg(exchange string, msg string) error {
	err := GetConn(exchange).WriteMessage(websocket.TextMessage, []byte(msg))
	commons.HandleErr(err, errSendMsg)
	return nil
}
