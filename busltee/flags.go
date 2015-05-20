package busltee

import (
	"errors"

	flag "github.com/heroku/busl/Godeps/_workspace/src/github.com/ogier/pflag"
)

const usage = "Usage: busltee <url> [-k|--insecure] [--connect-timeout N] -- <command>"

type flags struct {
	Insecure  bool
	Timeout   float64
	Retry     int
	URL       string
	Args      []string
	LogPrefix string
	LogFile   string
}

func ParseFlags() (*flags, error) {
	f := &flags{}

	// Connection related flags
	flag.BoolVarP(&f.Insecure, "insecure", "k", false, "allows insecure SSL connections")
	flag.IntVar(&f.Retry, "retry", 5, "max retries for connect timeout errors")
	flag.Float64Var(&f.Timeout, "connect-timeout", 1, "max number of seconds to connect to busl URL")

	// Logging related flags
	flag.StringVar(&f.LogPrefix, "log-prefix", "", "log prefix")
	flag.StringVar(&f.LogFile, "log-file", "", "log file")

	if flag.Parse(); len(flag.Args()) < 2 {
		return nil, errors.New(usage)
	}

	f.URL = flag.Arg(0)
	f.Args = flag.Args()[1:]

	return f, nil
}
