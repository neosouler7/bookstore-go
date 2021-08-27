package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/neosouler7/bookstore-go/binance"
	"github.com/neosouler7/bookstore-go/bithumb"
	"github.com/neosouler7/bookstore-go/coinone"
	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/gopax"
	"github.com/neosouler7/bookstore-go/huobikorea"
	"github.com/neosouler7/bookstore-go/korbit"
	"github.com/neosouler7/bookstore-go/tgmanager"
	"github.com/neosouler7/bookstore-go/upbit"
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

	tgmanager.InitBot(
		commons.ReadConfig("Tg").(map[string]interface{})["token"].(string),
		commons.ReadConfig("Tg").(map[string]interface{})["chat_ids"].([]interface{}),
		commons.SetTimeZone("Tg"),
	)

	tgMsg := fmt.Sprintf("[bookstore-go] %s\n", *exchange)
	tgmanager.SendMsg(tgMsg)
	switch *exchange {
	default:
		usage()
	case "bin":
		binance.Run(*exchange)
	case "bmb":
		bithumb.Run(*exchange)
	case "con":
		coinone.Run(*exchange)
	case "gpx":
		gopax.Run(*exchange)
	case "hbk":
		huobikorea.Run(*exchange)
	case "kbt":
		korbit.Run(*exchange)
	case "upb":
		upbit.Run(*exchange)
	}
}
