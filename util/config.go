package util

import (
	"flag"
	"os"
	"time"
)

var (
	HttpPort          = flag.String("httpPort", os.Getenv("PORT"), "HTTP port for the server.")
	HeartbeatDuration = flag.Duration("subscribeHeartbeatDuration", time.Second*10, "Heartbeat interval for HTTP stream subscriptions.")
)
