package busltee

import (
	"errors"

	flag "github.com/heroku/busl/Godeps/_workspace/src/github.com/ogier/pflag"
)

const usage = "Usage: busltee <url> [-k|--insecure] [--connect-timeout N] -- <command>"

type config struct {
	Insecure  bool
	Timeout   float64
	Retry     int
	URL       string
	Args      []string
	LogPrefix string
	LogFile   string
}

func Config() (*config, error) {
	conf := &config{}

	// Connection related flags
	flag.BoolVarP(&conf.Insecure, "insecure", "k", false, "allows insecure SSL connections")
	flag.IntVar(&conf.Retry, "retry", 5, "max retries for connect timeout errors")
	flag.Float64Var(&conf.Timeout, "connect-timeout", 1, "max number of seconds to connect to busl URL")

	// Logging related flags
	flag.StringVar(&conf.LogPrefix, "log-prefix", "", "log prefix")
	flag.StringVar(&conf.LogFile, "log-file", "", "log file")

	if flag.Parse(); len(flag.Args()) < 2 {
		return nil, errors.New(usage)
	}

	conf.URL = flag.Arg(0)
	conf.Args = flag.Args()[1:]

	return conf, nil
}
