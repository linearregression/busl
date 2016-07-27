package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"

	"github.com/heroku/busl/server"
	"github.com/heroku/busl/util"
)

func main() {
	os.Setenv("GODEBUG", "http2server=0")

	fmt.Println("Starting busl...")
	flag.Parse()

	_, err := strconv.Atoi(*util.HTTPPort)
	if err != nil {
		log.Printf("%s: $PORT must be an integer value.\n", os.Args[0])
		os.Exit(1)
	}
	server.Start(*util.HTTPPort, util.AwaitSignals(syscall.SIGURG))
}
