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

func GetConn(exchange string) (*websocket.Conn, error) {
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

	u := url.URL{Scheme: "wss", Host: host, Path: path}
	wsConn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	commons.HandleErr(err, errGetConn)

	return wsConn, nil
}

func SendMsg(wsConn *websocket.Conn, msg string) error {
	err := wsConn.WriteMessage(websocket.TextMessage, []byte(msg))
	commons.HandleErr(err, errSendMsg)
	return nil
}
