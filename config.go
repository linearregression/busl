package busl

import (
	"flag"
	"os"
)

var (
	httpPort = flag.String("httpPort", os.Getenv("PORT"), "HTTP port for the server.")
)
