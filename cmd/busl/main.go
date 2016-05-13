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
	_ "github.com/joho/godotenv/autoload"
)

type cmdConfig struct {
	RollbarEnvironment string
	RollbarToken       string

	HTTPPort         string
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
}

func main() {
	logging.Configure()

	fmt.Println("Starting busl...")
	cmdConf, httpConf, err := parseFlags()
	if err != nil {
		os.Exit(1)
	}

	if cmdConf.RollbarToken != "" {
		rollbar.Token = cmdConf.RollbarToken
		rollbar.Environment = cmdConf.RollbarEnvironment
		rollbar.ServerRoot = "github.com/heroku/busl"
	}

	_, err = strconv.Atoi(cmdConf.HTTPPort)
	if err != nil {
		log.Printf("%s: $PORT must be an integer value.\n", os.Args[0])
		os.Exit(1)
	}

	s := server.NewServer(httpConf)
	s.ReadTimeout = cmdConf.HTTPReadTimeout
	s.WriteTimeout = cmdConf.HTTPWriteTimeout
	s.Start(cmdConf.HTTPPort, awaitSignals(syscall.SIGURG))
}

func parseFlags() (*cmdConfig, *server.Config, error) {
	httpConf := &server.Config{}
	cmdConf := &cmdConfig{}

	flag.StringVar(&cmdConf.RollbarEnvironment, "rollbarEnvironment", os.Getenv("ROLLBAR_ENVIRONMENT"), "Rollbar Enviornment for this application (development/staging/production).")
	flag.StringVar(&cmdConf.RollbarToken, "rollbarToken", os.Getenv("ROLLBAR_TOKEN"), "Rollbar Token for sending issues to Rollbar.")

	flag.StringVar(&cmdConf.HTTPPort, "httpPort", os.Getenv("PORT"), "HTTP port for the server.")
	flag.DurationVar(&cmdConf.HTTPReadTimeout, "httpReadTimeout", time.Hour, "Timeout for HTTP request reading")
	flag.DurationVar(&cmdConf.HTTPWriteTimeout, "httpWriteTimeout", time.Hour, "Timeout for HTTP request writing")

	flag.StringVar(&httpConf.Credentials, "creds", os.Getenv("CREDS"), "user1:pass1|user2:pass2")
	flag.BoolVar(&httpConf.EnforceHTTPS, "enforceHttps", os.Getenv("ENFORCE_HTTPS") == "1", "Whether to enforce use of HTTPS.")
	flag.DurationVar(&httpConf.HeartbeatDuration, "subscribeHeartbeatDuration", time.Second*10, "Heartbeat interval for HTTP stream subscriptions.")
	flag.StringVar(&httpConf.StorageBaseURL, "storageBaseURL", os.Getenv("STORAGE_BASE_URL"), "Optional persistent blob storage (i.e. S3)")

	flag.Parse()

	return cmdConf, httpConf, nil
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
