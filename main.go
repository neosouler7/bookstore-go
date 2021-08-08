package main

import (
	"errors"
	"fmt"
	"log"
	"neosouler7/bookstore-go/binance"
	"neosouler7/bookstore-go/coinone"
	"neosouler7/bookstore-go/tgmanager"
	"neosouler7/bookstore-go/upbit"
	"os"
)

var (
	errExecFormat       = errors.New("[ERROR] ex) go run main.go {exchange}")
	errExchangeNotFound = errors.New("[ERROR] enter proper exchange")
)

func main() {
	args := os.Args
	if len(args) != 2 {
		log.Fatalln(errExecFormat)
	}

	exchange := args[1]
	tgMsg := fmt.Sprintf("[bookstore-go] %s\n", exchange)

	switch exchange {
	default:
		log.Fatalln(errExchangeNotFound)
	case "upb":
		tgmanager.SendMsg(tgMsg)
		upbit.Run(exchange)
	case "con":
		tgmanager.SendMsg(tgMsg)
		coinone.Run(exchange)
	case "bin":
		tgmanager.SendMsg(tgMsg)
		binance.Run(exchange)
	}
}
