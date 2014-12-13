package util

import (
	"flag"
	"os"
	"time"
)

var (
	EnforceHTTPS      = flag.Bool("enforceHttps", os.Getenv("ENFORCE_HTTPS") == "1", "Whether to enforce use of HTTPS.")
	HttpPort          = flag.String("httpPort", os.Getenv("PORT"), "HTTP port for the server.")
	HeartbeatDuration = flag.Duration("subscribeHeartbeatDuration", time.Second*10, "Heartbeat interval for HTTP stream subscriptions.")
)
