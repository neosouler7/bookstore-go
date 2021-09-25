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
	"github.com/neosouler7/bookstore-go/config"
	"github.com/neosouler7/bookstore-go/tgmanager"
)

var (
	w            *websocket.Conn
	once         sync.Once
	ErrReadMsg   = errors.New("[ERROR] reading msg on ws")
	SubscribeMsg = "%s websocket subscribed!\n"
	FilteredMsg  = "%s websocket msg filtered - %s\n"
)

const (
	bin string = "stream.binance.com:9443"
	bmb string = "pubwss.bithumb.com"
	con string = "public-ws-api.coinone.co.kr"
	gpx string = "wsapi.gopax.co.kr"
	hbk string = "api-cloud.huobi.co.kr"
	kbt string = "ws.korbit.co.kr"
	upb string = "api.upbit.com"
)

type hostPath struct {
	host string
	path string
}

func Conn(exchange string) *websocket.Conn {
	once.Do(func() {
		h := &hostPath{}
		h.getHostPath(exchange)

		u := url.URL{Scheme: "wss", Host: h.host, Path: h.path}
		if exchange == "gpx" {
			apiKey := config.GetApiKey(exchange)
			publicKey, secretKey := apiKey.Public, apiKey.Secret

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
	return w
}

func (h *hostPath) getHostPath(exchange string) {
	switch exchange {
	case "bin":
		h.host = bin
		h.path = "/stream"
	case "bmb":
		h.host = bmb
		h.path = "/pub/ws"
	case "con":
		h.host = con
		h.path = ""
	case "gpx":
		h.host = gpx
		h.path = ""
	case "hbk":
		h.host = hbk
		h.path = "/ws"
	case "kbt":
		h.host = kbt
		h.path = "/v1/user/push"
	case "upb":
		h.host = upb
		h.path = "/websocket/v1"
	}
}

func SendMsg(exchange, msg string) {
	err := Conn(exchange).WriteMessage(websocket.TextMessage, []byte(msg))
	tgmanager.HandleErr(exchange, err)
}

func Pong(exchange string) {
	err := Conn(exchange).WriteMessage(websocket.PongMessage, []byte{})
	tgmanager.HandleErr(exchange, err)
}
