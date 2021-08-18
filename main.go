package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/neosouler7/bookstore-go/bithumb"
	"github.com/neosouler7/bookstore-go/coinone"
	"github.com/neosouler7/bookstore-go/huobikorea"
	"github.com/neosouler7/bookstore-go/korbit"
	"github.com/neosouler7/bookstore-go/upbit"

	"github.com/neosouler7/bookstore-go/binance"
	"github.com/neosouler7/bookstore-go/tgmanager"
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
	case "bmb":
		tgmanager.SendMsg(tgMsg)
		bithumb.Run(*exchange)
	}
}
