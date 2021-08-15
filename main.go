package main

import (
	"errors"
	"flag"
	"fmt"
	"neosouler7/bookstore-go/binance"
	"neosouler7/bookstore-go/coinone"
	"neosouler7/bookstore-go/huobikorea"
	"neosouler7/bookstore-go/korbit"
	"neosouler7/bookstore-go/tgmanager"
	"neosouler7/bookstore-go/upbit"
	"os"
)

var (
	errExchangeNotFound = errors.New("[ERROR] enter proper exchange")
)

func usage() {
	fmt.Print("Welcome to bookstore-go\n\n")
	fmt.Print("Please use the following commands\n\n")
	fmt.Print("-e : Set exchange code to run\n")
	os.Exit(0)
}

func main() {
	args := os.Args
	if len(args) == 1 {
		usage()
	}

	exchange := flag.String("e", "", "Set exchange code to run")
	flag.Parse()

	tgMsg := fmt.Sprintf("[bookstore-go] %s\n", *exchange)
	switch *exchange {
	default:
		usage()
	case "upb":
		tgmanager.SendMsg(tgMsg)
		upbit.Run(*exchange)
	case "con":
		tgmanager.SendMsg(tgMsg)
		coinone.Run(*exchange)
	case "bin":
		tgmanager.SendMsg(tgMsg)
		binance.Run(*exchange)
	case "kbt":
		tgmanager.SendMsg(tgMsg)
		korbit.Run(*exchange)
	case "hbk":
		tgmanager.SendMsg(tgMsg)
		huobikorea.Run(*exchange)
	}
}
