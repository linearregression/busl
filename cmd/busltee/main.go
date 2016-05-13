package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/heroku/busl/logging"
	"github.com/heroku/busl/publisher"
	"github.com/heroku/rollbar"

	_ "github.com/joho/godotenv/autoload"
	flag "github.com/ogier/pflag"
)

// global cli options
var (
	RollbarEnvironment = flag.String("rollbarEnvironment", os.Getenv("ROLLBAR_ENVIRONMENT"), "Rollbar Enviornment for this application (development/staging/production).")
	RollbarToken       = flag.String("rollbarToken", os.Getenv("ROLLBAR_TOKEN"), "Rollbar Token for sending issues to Rollbar.")
)

type cmdConfig struct {
	RollbarEnvironment string
	RollbarToken       string
}

func main() {
	logging.Configure()

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

	publisher.OpenLogs(publisherConf.LogFile, publisherConf.LogPrefix)
	defer publisher.CloseLogs()

	if exitCode := publisher.Run(publisherConf.URL, publisherConf.Args, publisherConf); exitCode != 0 {
		os.Exit(exitCode)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] <url> -- <command>\n", os.Args[0])
	flag.PrintDefaults()
}

func parseFlags() (*cmdConfig, *publisher.Config, error) {
	publisherConf := &publisher.Config{}
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
