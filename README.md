# Welcome to bookstore-go!
본 프로젝트는 RESTful / Websocket 두 가지 API 방식을 종합 활용하여,  
여러 가상화폐 거래소들의 호가 정보를 별도 저장소에 **최소 Latency로 Clone 하는 것을 지향**합니다.
#
#
## 지원 거래소
현재 아래의 국내/해외 주요 거래소 대상으로 지원하고 있으며, 향후 지속 추가 예정입니다.
|구분|이름|코드|
|:---:|:---:|:---:|
|해외|바이낸스|bin|
|국내|빗썸|bmb|
|국내|코인원|con|
|국내|고팍스|gpx|
|국내|후오비코리아|hbk|
|국내|코빗|kbt|
|국내|업비트|upb|
  > 코인원 Websocket의 경우 자체VIP 선정 되어야 사용 가능  
#
#
#
# 코드 설명
### External Package
- [go-redis](https://pkg.go.dev/github.com/go-redis/redis/v8")
- [fasthttp](https://pkg.go.dev/github.com/valyala/fasthttp)
- [gorilla Websocket](https://pkg.go.dev/github.com/gorilla/websocket)
- [telegram-bot-api](https://pkg.go.dev/github.com/go-telegram-bot-api/telegram-bot-api)  
#
#

### 구조
모든 거래소에서 공통으로 사용하는 기능들을 패키지로 각각 별도 구성하였고,  
> commons / redis / telegram / rest / websocket  

거래소별로는 주요 로직이 담겨 있는 core.go 와 부수적인 기능의 sub.go 로 구분 작성하였습니다.  
또한 flag로 전달 받은 거래소 코드별 독립 process로 실행됩니다.
