package util

import (
	"flag"
	"os"
	"time"

	"github.com/heroku/rollbar"
)

// all the available cli flags
var (
	Creds              = flag.String("creds", os.Getenv("CREDS"), "user1:pass1|user2:pass2")
	EnforceHTTPS       = flag.Bool("enforceHttps", os.Getenv("ENFORCE_HTTPS") == "1", "Whether to enforce use of HTTPS.")
	HeartbeatDuration  = flag.Duration("subscribeHeartbeatDuration", time.Second*10, "Heartbeat interval for HTTP stream subscriptions.")
	HTTPPort           = flag.String("httpPort", os.Getenv("PORT"), "HTTP port for the server.")
	HTTPReadTimeout    = flag.Duration("httpReadTimeout", time.Hour, "Timeout for HTTP request reading")
	HTTPWriteTimeout   = flag.Duration("httpWriteTimeout", time.Hour, "Timeout for HTTP request writing")
	RollbarEnvironment = flag.String("rollbarEnvironment", os.Getenv("ROLLBAR_ENVIRONMENT"), "Rollbar Enviornment for this application (development/staging/production).")
	RollbarToken       = flag.String("rollbarToken", os.Getenv("ROLLBAR_TOKEN"), "Rollbar Token for sending issues to Rollbar.")
	StorageBaseURL     = flag.String("storageBaseURL", os.Getenv("STORAGE_BASE_URL"), "Optional persistent blob storage (i.e. S3)")
)

func init() {
	if *RollbarToken != "" {
		rollbar.Token = *RollbarToken
		rollbar.Environment = *RollbarEnvironment
		rollbar.ServerRoot = "github.com/heroku/busl"
	}
}
