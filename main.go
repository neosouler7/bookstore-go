package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/neosouler7/bookstore-go/binance"
	"github.com/neosouler7/bookstore-go/binancef"
	"github.com/neosouler7/bookstore-go/bithumb"
	"github.com/neosouler7/bookstore-go/coinone"
	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/config"
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
	err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://a3801863e72346b8b22df3177b41d3f3@o1395002.ingest.sentry.io/6717420",
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

	args := os.Args
	if len(args) == 1 {
		usage()
	}

	exchange := flag.String("e", "", "Set exchange code to run")
	flag.Parse()

	tgConfig := config.GetTg()
	tgmanager.InitBot(
		tgConfig.Token,
		tgConfig.Chat_ids,
		commons.SetTimeZone("Tg"),
	)

	tgMsg := fmt.Sprintf("[bookstore-go %s] %s", config.GetName(), *exchange)
	switch *exchange {
	default:
		usage()
	case "bin":
		tgmanager.SendMsg(tgMsg)
		binance.Run(*exchange)
	case "binf":
		tgmanager.SendMsg(tgMsg)
		binancef.Run(*exchange)
	case "bmb":
		tgmanager.SendMsg(tgMsg)
		bithumb.Run(*exchange)
	case "con":
		tgmanager.SendMsg(tgMsg)
		coinone.Run(*exchange)
	case "gpx":
		tgmanager.SendMsg(tgMsg)
		gopax.Run(*exchange)
	case "hbk":
		tgmanager.SendMsg(tgMsg)
		huobikorea.Run(*exchange)
	case "kbt":
		tgmanager.SendMsg(tgMsg)
		korbit.Run(*exchange)
	case "upb":
		tgmanager.SendMsg(tgMsg)
		upbit.Run(*exchange)
	}
}
