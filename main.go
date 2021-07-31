package main

import (
	"errors"
	"fmt"
	"log"
	"neosouler7/bookstore-go/coinone"
	"neosouler7/bookstore-go/tgmanager"
	"neosouler7/bookstore-go/upbit"
	"os"
)

var (
	errorExchangeNotFound = errors.New("[ERROR] enter proper exchange")
)

func main() {
	args := os.Args
	if len(args) != 2 {
		panic(errorExchangeNotFound)
	}
	exchange := args[1]

	// send start msg
	tgMsg := fmt.Sprintf("[bookstore-GO] %s\n", exchange)

	switch exchange {
	default:
		log.Fatalln(errorExchangeNotFound)
	case "upb":
		tgmanager.SendTgMessage(tgMsg)
		upbit.Run(exchange)
	case "con":
		tgmanager.SendTgMessage(tgMsg)
		coinone.Run(exchange)
	}
}
