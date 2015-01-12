package main

import (
	"flag"
	"fmt"

	"github.com/heroku/busl/server"
)

func main() {
	fmt.Println("Starting busl...")
	flag.Parse()
	server.Start()
}
