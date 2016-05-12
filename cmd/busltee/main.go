package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/heroku/busl/logging"
	"github.com/heroku/busl/publisher"
	"github.com/heroku/rollbar"
	flag "github.com/ogier/pflag"
)

// global cli options
var (
	RollbarEnvironment = flag.String("rollbarEnvironment", os.Getenv("ROLLBAR_ENVIRONMENT"), "Rollbar Enviornment for this application (development/staging/production).")
	RollbarToken       = flag.String("rollbarToken", os.Getenv("ROLLBAR_TOKEN"), "Rollbar Token for sending issues to Rollbar.")
)

func main() {
	logging.Configure()

	conf, err := parseFlags()
	if err != nil {
		usage()
		os.Exit(1)
	}

	if *RollbarToken != "" {
		rollbar.Token = *RollbarToken
		rollbar.Environment = *RollbarEnvironment
		rollbar.ServerRoot = "github.com/heroku/busl"
	}

	publisher.OpenLogs(conf.LogFile, conf.LogPrefix)
	defer publisher.CloseLogs()

	if exitCode := publisher.Run(conf.URL, conf.Args, conf); exitCode != 0 {
		os.Exit(exitCode)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] <url> -- <command>\n", os.Args[0])
	flag.PrintDefaults()
}

func parseFlags() (*publisher.Config, error) {
	conf := &publisher.Config{}

	// Connection related flags
	flag.BoolVarP(&conf.Insecure, "insecure", "k", false, "allows insecure SSL connections")
	flag.IntVar(&conf.Retry, "retry", 5, "max retries for connect timeout errors")
	flag.Float64Var(&conf.Timeout, "connect-timeout", 1, "max number of seconds to connect to busl URL")

	// Logging related flags
	flag.StringVar(&conf.LogPrefix, "log-prefix", "", "log prefix")
	flag.StringVar(&conf.LogFile, "log-file", "", "log file")
	flag.StringVar(&conf.RequestID, "request-id", "", "request id")

	if flag.Parse(); len(flag.Args()) < 2 {
		return nil, errors.New("insufficient args")
	}

	conf.URL = flag.Arg(0)
	conf.Args = flag.Args()[1:]

	return conf, nil
}
