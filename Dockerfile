FROM golang:alpine

LABEL maintainer=neosouler@gmail.com

ENV GOPATH=/go
ENV TZ=Asia/Seoul

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

WORKDIR /go
RUN apk update
RUN apk --update add openrc --no-cache
RUN apk --update add git --no-cache 
RUN apk --update add supervisor --no-cache

WORKDIR /go/src
RUN git clone https://github.com/neosouler7/bookstore-go

WORKDIR /go/src/bookstore-go
RUN go get .
RUN go build .
RUN mkdir /etc/supervisor.d

CMD ["/usr/bin/supervisord", "-n"]

# docker build -t neosouler/bookstore-go:1.1.17 . --no-cache —-platform linux/amd64
# docker push neosouler/bookstore-go:1.1.17