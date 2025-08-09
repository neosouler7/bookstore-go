package upbit

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/config"
	"github.com/neosouler7/bookstore-go/restmanager"
	"github.com/neosouler7/bookstore-go/tgmanager"
	"github.com/neosouler7/bookstore-go/websocketmanager"

	"github.com/google/uuid"
)

var (
	exchange string
	// pingMsg  string = "PING"
)

func pongWs(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			websocketmanager.Ping(exchange)
		case <-ctx.Done():
			return
		}
	}
}

func subscribeWs(pairs []string, wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(time.Second * 1)

	var streamSlice []string
	for _, pair := range pairs {
		pairInfo := strings.Split(pair, ":")
		market, symbol := strings.ToUpper(pairInfo[0]), strings.ToUpper(pairInfo[1])
		streamSlice = append(streamSlice, fmt.Sprintf("'%s-%s'", market, symbol))
	}

	uuid := uuid.NewString()
	streams := strings.Join(streamSlice, ",")
	msg := fmt.Sprintf("[{'ticket':'%s'}, {'type': 'orderbook', 'codes': [%s]}]", uuid, streams)

	websocketmanager.SendMsg(exchange, msg)
	fmt.Printf(websocketmanager.SubscribeMsg, exchange)
}

func receiveWs(ctx context.Context, cancel context.CancelFunc, msgQueue chan<- []byte) {
	defer close(msgQueue)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, msgBytes, err := websocketmanager.Conn(exchange).ReadMessage()
			if err != nil {
				tgmanager.HandleErr(exchange, err)
				cancel() // 에러 발생 시 모든 관련 작업 취소
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

func processWsMessages(ctx context.Context, msgQueue <-chan []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		case msgBytes := <-msgQueue:
			if strings.Contains(string(msgBytes), "status") {
				fmt.Println("PONG") // {"status":"UP"}
			} else {
				var rJson interface{}
				commons.Bytes2Json(msgBytes, &rJson)

				// 컨텍스트로 제어하는 고루틴
				go func() {
					// 컨텍스트 취소 시 즉시 종료
					select {
					case <-ctx.Done():
						return
					default:
						// 작업 완료 후 컨텍스트 다시 확인
						if ctx.Err() != nil {
							return
						}
						SetOrderbook("W", exchange, rJson.(map[string]interface{}))
					}
				}()
			}
		}
	}
}

func rest(ctx context.Context, pairs []string, restQueue chan<- map[string]interface{}) {
	ticker := time.NewTicker(500 * time.Millisecond)
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

			pair := pairs[pairIndex] // 현재 인덱스의 페어에 대해서만 고루틴으로 호출
			go func(p string) {
				rJson := restmanager.FastHttpRequest2(exchange, "GET", p)
				select {
				case restQueue <- rJson:
				case <-ctx.Done():
					return
				}
			}(pair)

			pairIndex = (pairIndex + 1) % len(pairs)
		}
	}
}

func processRestResponses(ctx context.Context, restQueue <-chan map[string]interface{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case rJson := <-restQueue:
			// 컨텍스트로 제어하는 고루틴
			go func() {
				select {
				case <-ctx.Done():
					return
				default:
					if ctx.Err() != nil {
						return
					}
					SetOrderbook("R", exchange, rJson)
				}
			}()
		}
	}
}

func Run(e string) {
	exchange = e
	pairs := config.GetPairs(exchange)
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Run 함수 종료 시 모든 컨텍스트 취소

	wsQueue := make(chan []byte, 100)                            // WebSocket 메시지 큐
	restQueue := make(chan map[string]interface{}, len(pairs)*2) // REST 응답 큐

	// ping
	wg.Add(1)
	go func() {
		defer wg.Done()
		pongWs(ctx)
	}()

	// subscribe websocket stream
	wg.Add(1)
	go subscribeWs(pairs, &wg)

	// receive websocket msg
	wg.Add(1)
	go func() {
		defer wg.Done()
		receiveWs(ctx, cancel, wsQueue)
	}()

	// process websocket messages
	wg.Add(1)
	go func() {
		defer wg.Done()
		processWsMessages(ctx, wsQueue)
	}()

	// rest
	wg.Add(1)
	go func() {
		defer wg.Done()
		rest(ctx, pairs, restQueue)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		processRestResponses(ctx, restQueue)
	}()

	<-ctx.Done() // 웹소켓 에러 등으로 컨텍스트가 취소될 때까지 대기
	wg.Wait()    // 모든 고루틴이 정상적으로 종료될 때까지 대기
}
