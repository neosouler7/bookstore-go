package main

import (
	"flag"
	"fmt"

	_ "net/http/pprof"
	"os"

	"github.com/neosouler7/bookstore-go/binance"
	"github.com/neosouler7/bookstore-go/binancef"
	"github.com/neosouler7/bookstore-go/bithumb"
	"github.com/neosouler7/bookstore-go/coinone"
	"github.com/neosouler7/bookstore-go/commons"
	"github.com/neosouler7/bookstore-go/config"
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
	// only for pprof
	// additional command: go tool pprof -http :8080 http://localhost:6060/debug/pprof/profile\?seconds\=120
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	// for profiling
	// go func() {
	// 	for {
	// 		var m runtime.MemStats
	// 		runtime.ReadMemStats(&m)
	// 		fmt.Printf("HeapAlloc = %v", (m.HeapAlloc))
	// 		fmt.Printf("\tHeapObjects = %v", (m.HeapObjects))
	// 		fmt.Printf("\tHeapSys = %v", (m.Sys))
	// 		fmt.Printf("\tNumGC = %v\n", m.NumGC)
	// 		time.Sleep(5 * time.Second)
	// 	}
	// }()

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

	runners := map[string]func(string){
		"bin": binance.Run,
		"bif": binancef.Run,
		"bmb": bithumb.Run,
		"con": coinone.Run,
		"kbt": korbit.Run,
		"upb": upbit.Run,
	}

	run, ok := runners[*exchange]
	if !ok {
		usage()
	}

	tgMsg := fmt.Sprintf("## START %s %s\n- version 1.2.0", config.GetName(), *exchange)
	tgmanager.SendMsg(tgMsg)
	run(*exchange)
}
