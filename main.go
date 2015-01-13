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
	fmt.Println("Starting busl...")
	flag.Parse()

	_, err := strconv.Atoi(*util.HttpPort)
	if err != nil {
		log.Printf("%s: $PORT must be an integer value.\n", os.Args[0])
		os.Exit(1)
	}
	server.Start(*util.HttpPort, util.AwaitSignals(syscall.SIGURG))
}
