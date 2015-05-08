package main

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	flag "github.com/heroku/busl/Godeps/_workspace/src/github.com/ogier/pflag"
)

const usage string = "Usage: busltee <url> [-k|--insecure] [--connect-timeout N] -- <command>"

type config struct {
	insecure  bool
	timeout   string
	logPrefix string
	logFile   string
}

func main() {
	conf := &config{}

	flag.BoolVarP(&conf.insecure, "insecure", "k", false, "allows insecure SSL connections")
	flag.StringVar(&conf.timeout, "connect-timeout", "5", "max number of seconds to connect to busl URL")
	flag.StringVar(&conf.logPrefix, "log-prefix", "", "log prefix")
	flag.StringVar(&conf.logFile, "log-file", "", "log file")

	if flag.Parse(); len(flag.Args()) < 2 {
		log.Fatal(usage)
	}

	out := getLogOutput(conf.logFile)
	log.SetPrefix(conf.logPrefix + " ")
	log.SetOutput(out)
	log.SetFlags(0)
	if f, ok := out.(io.Closer); ok {
		defer f.Close()
	}

	url := flag.Arg(0)
	args := flag.Args()[1:]

	err := busltee(conf, url, args)
	if err != nil {
		log.Printf("busltee.main.error count#busltee.main.error=1 error=%v", err.Error())
		os.Exit(exitStatus(err))
	}
}

func monitor(subject string, ts time.Time) {
	log.Printf("%s.time time=%f", subject, time.Now().Sub(ts).Seconds())
}

func busltee(conf *config, url string, args []string) error {
	defer monitor("busltee.busltee", time.Now())

	reader, writer := io.Pipe()
	uploaded := make(chan struct{})

	go func() {
		if err := stream(url, reader, conf.insecure, conf.timeout); err != nil {
			log.Printf("busltee.stream.error count#busltee.stream.error=1 error=%v", err.Error())
			// Prevent writes from blocking.
			io.Copy(ioutil.Discard, reader)
		} else {
			log.Printf("busltee.stream.success count#busltee.stream.success=1")
		}
		close(uploaded)
	}()

	err := run(args, writer, writer)
	<-uploaded

	return err
}

// TODO: Use net/http when this issue has been fixed:
// @see https://github.com/golang/go/issues/6574
func stream(url string, stdin io.Reader, insecure bool, timeout string) error {
	defer monitor("busltee.stream", time.Now())

	if url == "" {
		log.Printf("count#busltee.stream.missingurl")
		return errors.New("Missing URL")
	}

	args := []string{
		"--connect-timeout", timeout,
		"-T", "-",
		"-H", "Transfer-Encoding: chunked",
		"-XPOST", url,
	}

	if insecure {
		args = append(args, "-k")
	}

	cmd := exec.Command("curl", args...)
	cmd.Stdin = stdin

	return cmd.Run()
}

func run(args []string, stdout, stderr io.WriteCloser) error {
	defer stdout.Close()
	defer stderr.Close()
	defer monitor("busltee.run", time.Now())

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = io.MultiWriter(stdout, os.Stdout)
	cmd.Stderr = io.MultiWriter(stderr, os.Stderr)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc)
	go func() {
		s := <-sigc
		cmd.Process.Signal(s)
	}()

	return cmd.Run()
}

func exitStatus(err error) int {
	if exit, ok := err.(*exec.ExitError); ok {
		if status, ok := exit.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}

	return 0
}

func getLogOutput(logFile string) io.Writer {
	if logFile == "" {
		return ioutil.Discard
	}
	if file, err := os.OpenFile(logFile, os.O_RDWR|os.O_APPEND, 0660); err != nil {
		return ioutil.Discard
	} else {
		return file
	}
}
