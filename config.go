package main

import (
	"flag"
	"os"
	"time"
)

var (
	httpPort                   = flag.String("httpPort", os.Getenv("PORT"), "HTTP port for the server.")
	subscribeHeartbeatDuration = flag.Duration("subscribeHeartbeatDuration", time.Second*10, "Heartbeat interval for HTTP stream subscriptions.")
)
