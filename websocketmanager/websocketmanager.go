package websocketmanager

import (
	"errors"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/neosouler7/bookstore-go/tgmanager"
)

var (
	w            *websocket.Conn
	mu           sync.RWMutex // protects conn
	wmu          sync.Mutex   // serializes all writes
	once         sync.Once
	ErrReadMsg   = errors.New("reading msg on ws")
	SubscribeMsg = "%s websocket subscribed!\n"
	FilteredMsg  = "%s websocket msg filtered - %s\n"
)

const (
	bin string = "stream.binance.com:9443"
	bif string = "fstream.binance.com"
	bmb string = "ws-api.bithumb.com"
	con string = "stream.coinone.co.kr"
	kbt string = "ws-api.korbit.co.kr" // "ws2.korbit.co.kr"
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
		wPointer, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		tgmanager.HandleErr(exchange, err)
		w = wPointer
	})
	return w
}

func Close() {
	mu.Lock()
	if w != nil {
		_ = w.Close()
		w = nil
	}
	mu.Unlock()
}

func (h *hostPath) getHostPath(exchange string) {
	switch exchange {
	case "bin":
		h.host = bin
		h.path = "/stream"
	case "bif":
		h.host = bif
		h.path = "/stream"
	case "bmb":
		h.host = bmb
		h.path = "/websocket/v1"
	case "con":
		h.host = con
		h.path = ""
	case "kbt":
		h.host = kbt
		h.path = "/v2/public" // "/v1/user/push"
	case "upb":
		h.host = upb
		h.path = "/websocket/v1"
	}
}

func SendMsg(exchange, msg string) {
	wmu.Lock()
	err := Conn(exchange).WriteMessage(websocket.TextMessage, []byte(msg))
	wmu.Unlock()
	tgmanager.HandleErr(exchange, err)
}

func Ping(exchange string) {
	wmu.Lock()
	err := Conn(exchange).WriteMessage(websocket.PingMessage, []byte{})
	wmu.Unlock()
	tgmanager.HandleErr(exchange, err)
}
