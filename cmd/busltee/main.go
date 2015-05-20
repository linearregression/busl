package main

import (
	"os"

	"github.com/heroku/busl/busltee"
)

func main() {
	conf, err := busltee.ParseFlags()
	if err != nil {
		busltee.Usage()
		os.Exit(1)
	}

	busltee.OpenLogs(conf.LogFile, conf.LogPrefix)
	defer busltee.CloseLogs()

	if exitCode := busltee.Run(conf.URL, conf.Args, conf); exitCode != 0 {
		os.Exit(exitCode)
	}
}
