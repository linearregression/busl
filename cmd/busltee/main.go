package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/heroku/busl/publisher"
	flag "github.com/ogier/pflag"
)

func main() {
	conf, err := parseFlags()
	if err != nil {
		usage()
		os.Exit(1)
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
