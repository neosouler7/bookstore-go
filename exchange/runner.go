package exchange

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/neosouler7/bookstore-go/config"
	"github.com/neosouler7/bookstore-go/restmanager"
	"github.com/neosouler7/bookstore-go/tgmanager"
	"github.com/neosouler7/bookstore-go/websocketmanager"
)

// Exchange defines the per-exchange behaviour delegated to by the shared runner.
type Exchange interface {
	Ping()
	Subscribe(pairs []string, wg *sync.WaitGroup)
	HandleWsMessage(msgBytes []byte)
	HandleRestResponse(rJson map[string]interface{})
}

func pongWs(ctx context.Context, e Exchange) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.Ping()
		case <-ctx.Done():
			return
		}
	}
}

func receiveWs(ctx context.Context, cancel context.CancelFunc, name string, msgQueue chan<- []byte) {
	defer close(msgQueue)

	// unblock read immediately when ctx is cancelled
	go func() {
		<-ctx.Done()
		websocketmanager.Close()
		tgmanager.HandleErr(name, fmt.Errorf("receiveWs lost"))
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, msgBytes, err := websocketmanager.Conn(name).ReadMessage()
			if err != nil {
				tgmanager.HandleErr(name, err)
				cancel() // cancel all related goroutines on error
				return
			}

			select {
			case msgQueue <- msgBytes:
			case <-ctx.Done():
				return
			}
		}
	}
}

func processWsMessages(ctx context.Context, e Exchange, msgQueue <-chan []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		case msgBytes, ok := <-msgQueue:
			if !ok {
				return
			}
			e.HandleWsMessage(msgBytes)
		}
	}
}

func rest(ctx context.Context, name string, pairs []string, restQueue chan<- map[string]interface{}) {
	buffer, rateLimit := config.GetRateLimit(name)
	ticker := time.NewTicker(time.Millisecond * time.Duration(int(1/rateLimit*10*100*buffer)))
	defer ticker.Stop()

	pairIndex := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if ctx.Err() != nil {
				return
			}

			p := pairs[pairIndex]
			pairIndex = (pairIndex + 1) % len(pairs)
			rJson := restmanager.FastHttpRequest2(name, "GET", p)
			select {
			case restQueue <- rJson:
			case <-ctx.Done():
				return
			}
		}
	}
}

func processRestResponses(ctx context.Context, e Exchange, restQueue <-chan map[string]interface{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case rJson, ok := <-restQueue:
			if !ok {
				return
			}
			e.HandleRestResponse(rJson)
		}
	}
}

// Run wires all goroutines for the given exchange.
func Run(name string, e Exchange) {
	pairs := config.GetPairs(name)
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel all contexts when Run exits

	wsQueue := make(chan []byte, 100)                            // WebSocket message queue
	restQueue := make(chan map[string]interface{}, len(pairs)*2) // REST response queue

	// ping
	wg.Add(1)
	go func() {
		defer wg.Done()
		pongWs(ctx, e)
	}()

	// subscribe websocket stream
	wg.Add(1)
	go e.Subscribe(pairs, &wg)

	// receive websocket msg
	wg.Add(1)
	go func() {
		defer wg.Done()
		receiveWs(ctx, cancel, name, wsQueue)
	}()

	// process websocket messages
	wg.Add(1)
	go func() {
		defer wg.Done()
		processWsMessages(ctx, e, wsQueue)
	}()

	// rest
	wg.Add(1)
	go func() {
		defer wg.Done()
		rest(ctx, name, pairs, restQueue)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		processRestResponses(ctx, e, restQueue)
	}()

	<-ctx.Done() // wait until context is cancelled (e.g. WebSocket error)
	wg.Wait()    // wait for all goroutines to finish cleanly
}
