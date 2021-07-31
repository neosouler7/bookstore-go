package websocketmanager

import (
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

var (
	errGetWsConn = errors.New("[ERROR] connecting ws")
	errSendWsMsg = errors.New("[ERROR] sending msg on ws")
)

const UPB_WS string = "api.upbit.com"
const CON_WS string = "public-ws-api.coinone.co.kr"

func GetWsConn(exchange string) (*websocket.Conn, error) {
	var host, path string
	switch exchange {
	case "upb":
		host = UPB_WS
		path = "/websocket/v1"
	case "con":
		host = CON_WS
		path = ""
	}

	u := url.URL{Scheme: "wss", Host: host, Path: path}
	fmt.Printf("connecting to %s\n", u.String())
	wsConn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, errGetWsConn
	}
	// defer wsConn.Close()

	return wsConn, nil
}

func SendWsMsg(wsConn *websocket.Conn, msg string) error {
	err := wsConn.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		log.Fatalln(errSendWsMsg)
		return errSendWsMsg
	}
	return nil
}
