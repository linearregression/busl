package busltee

import (
	"crypto/tls"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func Run(url string, args []string, flag *flags) (exitCode int) {
	defer monitor("busltee.busltee", time.Now())

	reader, writer := io.Pipe()
	uploaded := make(chan struct{})

	go func() {
		if err := stream(flag.Retry, url, reader, flag.Insecure, flag.Timeout); err != nil {
			log.Printf("busltee.stream.error count#busltee.stream.error=1 error=%v", err.Error())
			// Prevent writes from blocking.
			io.Copy(ioutil.Discard, reader)
		} else {
			log.Printf("busltee.stream.success count#busltee.stream.success=1")
		}
		close(uploaded)
	}()

	if err := run(args, writer, writer); err != nil {
		log.Printf("busltee.Run.error count#busltee.Run.error=1 error=%v", err.Error())
		exitCode = exitStatus(err)
	}

	// Only wait atmost 1 second after the full command has completed.
	// If it's not done by then, it probably means something else has gone
	// wrong and it's not worth waiting any longer.
	select {
	case <-uploaded:
	case <-time.After(time.Second):
		log.Printf("busltee.Run.upload.timeout count#busltee.Run.upload.timeout=1")
	}

	return exitCode
}

func monitor(subject string, ts time.Time) {
	log.Printf("%s.time time=%f", subject, time.Now().Sub(ts).Seconds())
}

func stream(retry int, url string, stdin io.Reader, insecure bool, timeout float64) (err error) {
	for retries := retry; retries >= 0; retries-- {
		if err = streamNoRetry(url, stdin, insecure, timeout); !isTimeout(err) {
			return err
		}
		log.Printf("count#busltee.stream.retry")
	}
	return err
}

var errMissingURL = errors.New("Missing URL")

func streamNoRetry(url string, stdin io.Reader, insecure bool, timeout float64) error {
	defer monitor("busltee.stream", time.Now())

	if url == "" {
		log.Printf("count#busltee.stream.missingurl")
		return errMissingURL
	}

	tr := &http.Transport{}

	// Using `Timeout` with a sub second long connect timeout
	// doesn't work. Using Deadline works though, which is pretty
	//  much the same thing, except a bit more verbose.
	if timeout > 0 {
		tr.Dial = (&net.Dialer{
			KeepAlive: 30 * time.Second,
			Deadline:  time.Now().Add(time.Duration(timeout) * time.Second),
		}).Dial
	}

	if insecure {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Prevent net/http from closing the reader on failure -- otherwise
	// we'll get broken pipe errors.
	req, err := http.NewRequest("POST", url, ioutil.NopCloser(stdin))
	if err != nil {
		return err
	}

	res, err := tr.RoundTrip(req)
	if res != nil {
		defer res.Body.Close()
	}
	return err
}

func run(args []string, stdout, stderr io.WriteCloser) error {
	defer stdout.Close()
	defer stderr.Close()
	defer monitor("busltee.run", time.Now())

	// Setup command with output multiplexed out to
	// stdout/stderr and also to the designated output
	// streams.
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = io.MultiWriter(stdout, os.Stdout)
	cmd.Stderr = io.MultiWriter(stderr, os.Stderr)

	if err := cmd.Start(); err != nil {
		return err
	}

	// Catch any signals sent to busltee, and pass those along.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc)
	go func() {
		s := <-sigc
		cmd.Process.Signal(s)
	}()

	return cmd.Wait()
}

func isTimeout(err error) bool {
	e, ok := err.(net.Error)
	return ok && e.Timeout()
}

func exitStatus(err error) int {
	if exit, ok := err.(*exec.ExitError); ok {
		if status, ok := exit.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}

	// Default to exit status 1 if we can't type assert the error.
	return 1
}
