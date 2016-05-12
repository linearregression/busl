package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/heroku/busl/logging"
	"github.com/heroku/busl/server"
	"github.com/heroku/rollbar"
)

// global cli options
var (
	HTTPPort         = flag.String("httpPort", os.Getenv("PORT"), "HTTP port for the server.")
	HTTPReadTimeout  = flag.Duration("httpReadTimeout", time.Hour, "Timeout for HTTP request reading")
	HTTPWriteTimeout = flag.Duration("httpWriteTimeout", time.Hour, "Timeout for HTTP request writing")

	RollbarEnvironment = flag.String("rollbarEnvironment", os.Getenv("ROLLBAR_ENVIRONMENT"), "Rollbar Enviornment for this application (development/staging/production).")
	RollbarToken       = flag.String("rollbarToken", os.Getenv("ROLLBAR_TOKEN"), "Rollbar Token for sending issues to Rollbar.")
)

func main() {
	logging.Configure()

	fmt.Println("Starting busl...")
	conf, err := parseFlags()
	if err != nil {
		os.Exit(1)
	}

	if *RollbarToken != "" {
		rollbar.Token = *RollbarToken
		rollbar.Environment = *RollbarEnvironment
		rollbar.ServerRoot = "github.com/heroku/busl"
	}

	_, err = strconv.Atoi(*HTTPPort)
	if err != nil {
		log.Printf("%s: $PORT must be an integer value.\n", os.Args[0])
		os.Exit(1)
	}

	s := server.NewServer(conf)
	s.ReadTimeout = *HTTPReadTimeout
	s.WriteTimeout = *HTTPWriteTimeout
	s.Start(*HTTPPort, awaitSignals(syscall.SIGURG))
}

func parseFlags() (*server.Config, error) {
	conf := &server.Config{}

	flag.StringVar(&conf.Credentials, "creds", os.Getenv("CREDS"), "user1:pass1|user2:pass2")
	flag.BoolVar(&conf.EnforceHTTPS, "enforceHttps", os.Getenv("ENFORCE_HTTPS") == "1", "Whether to enforce use of HTTPS.")
	flag.DurationVar(&conf.HeartbeatDuration, "subscribeHeartbeatDuration", time.Second*10, "Heartbeat interval for HTTP stream subscriptions.")
	flag.StringVar(&conf.StorageBaseURL, "storageBaseURL", os.Getenv("STORAGE_BASE_URL"), "Optional persistent blob storage (i.e. S3)")

	flag.Parse()

	return conf, nil
}

func awaitSignals(signals ...os.Signal) <-chan struct{} {
	s := make(chan os.Signal, 1)
	signal.Notify(s, signals...)
	log.Printf("signals.await signals=%v\n", signals)

	received := make(chan struct{})
	go func() {
		log.Printf("signals.received signal=%v\n", <-s)
		close(received)
	}()

	return received
}
