package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/heroku/busl/busltee"
	"github.com/heroku/rollbar"
	flag "github.com/ogier/pflag"
)

type cmdConfig struct {
	RollbarEnvironment string
	RollbarToken       string
}

func main() {
	cmdConf, publisherConf, err := parseFlags()
	if err != nil {
		usage()
		os.Exit(1)
	}

	if cmdConf.RollbarToken != "" {
		rollbar.Token = cmdConf.RollbarToken
		rollbar.Environment = cmdConf.RollbarEnvironment
		rollbar.ServerRoot = "github.com/heroku/busl"
	}

	busltee.OpenLogs(publisherConf.LogFile, publisherConf.LogPrefix)
	defer busltee.CloseLogs()

	if exitCode := busltee.Run(publisherConf.URL, publisherConf.Args, publisherConf); exitCode != 0 {
		os.Exit(exitCode)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] <url> -- <command>\n", os.Args[0])
	flag.PrintDefaults()
}

func parseFlags() (*cmdConfig, *busltee.Config, error) {
	publisherConf := &busltee.Config{}
	cmdConf := &cmdConfig{}

	flag.StringVar(&cmdConf.RollbarEnvironment, "rollbarEnvironment", os.Getenv("ROLLBAR_ENVIRONMENT"), "Rollbar Enviornment for this application (development/staging/production).")
	flag.StringVar(&cmdConf.RollbarToken, "rollbarToken", os.Getenv("ROLLBAR_TOKEN"), "Rollbar Token for sending issues to Rollbar.")

	// Connection related flags
	flag.BoolVarP(&publisherConf.Insecure, "insecure", "k", false, "allows insecure SSL connections")
	flag.IntVar(&publisherConf.Retry, "retry", 5, "max retries for connect timeout errors")
	flag.Float64Var(&publisherConf.Timeout, "connect-timeout", 1, "max number of seconds to connect to busl URL")

	// Logging related flags
	flag.StringVar(&publisherConf.LogPrefix, "log-prefix", "", "log prefix")
	flag.StringVar(&publisherConf.LogFile, "log-file", "", "log file")
	flag.StringVar(&publisherConf.RequestID, "request-id", "", "request id")

	if flag.Parse(); len(flag.Args()) < 2 {
		return nil, nil, errors.New("insufficient args")
	}

	publisherConf.URL = flag.Arg(0)
	publisherConf.Args = flag.Args()[1:]

	return cmdConf, publisherConf, nil
}
