# Welcome to bookstore-go!

본 프로젝트는 RESTful / Websocket 두 가지 API 방식을 모두 활용하여,
여러 가상화폐 거래소들의 정보를 **별도 공통 Redis에 최소 Latency로 Clone 하는 것을 지향**합니다.

## 지원 거래소

현재 아래의 국내/해외 주요 거래소 대상으로 지원하고 있으며, 향후 지속 추가 예정입니다.

- 업비트: upb (RESTful, Websocket)
- 코인원: con (RESTful, Websocket)
- 바이낸스: bin (RESTful, Websocket)
- 코빗: kbt (RESTful, Websocket)
- 후오비코리아: hbk (RESTful, Websocket)
- 빗썸: bmb (RESTful, Websocket)
- 고팍스: gpx (RESTful, Websocket)
  > 코인원 Websocket의 경우 코인원 VIP 선정 되어야 사용 가능

## 실행 방법

sample.json을 참고하여 데이터 저장을 위한 **Redis** 및 모니터링을 위한 **TelegramBot** 의 정보를 포함한 config.json 작성 후, 아래 cli의 형태로 main 함수를 실행합니다.

> go run main.go -e=**{exchangeCode}**

# 코드 설명

## 사용 External Package

- [go-redis](https://pkg.go.dev/github.com/go-redis/redis/v8")
- [fasthttp](https://pkg.go.dev/github.com/valyala/fasthttp)
- [gorilla Websocket](https://pkg.go.dev/github.com/gorilla/websocket)
- [telegram-bot-api](https://pkg.go.dev/github.com/go-telegram-bot-api/telegram-bot-api)

## 구조

모든 거래소에서 공통으로 사용하는 기능들을 패키지로 각각 별도 구성하였고,

> commons / redis / telegram / rest / websocket

거래소별로는 주요 로직이 담겨 있는 core.go 와 부수적인 기능의 sub.go 로 구분 작성하였습니다.

따라서, main.go 에서는 flag로 전달 받은 거래소의 core.go를 호출합니다.

## goroutine

to be continued...
