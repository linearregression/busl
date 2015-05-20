package main

import (
	"fmt"
	"os"

	"github.com/heroku/busl/busltee"
)

func main() {
	conf, err := busltee.ParseFlags()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	busltee.OpenLogs(conf.LogFile, conf.LogPrefix)
	defer busltee.CloseLogs()

	if exitCode := busltee.Run(conf.URL, conf.Args, conf); exitCode != 0 {
		os.Exit(exitCode)
	}
}
