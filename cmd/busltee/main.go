package main

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	flag "github.com/heroku/busl/Godeps/_workspace/src/github.com/ogier/pflag"
)

// TODO: Use net/http when this issue has been fixed:
// @see https://github.com/golang/go/issues/6574
func stream(url string, stdin io.Reader, insecure bool, timeout string) error {
	if url == "" {
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

func run(args []string, writer io.WriteCloser) error {
	defer writer.Close()

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = io.MultiWriter(writer, os.Stdout)
	cmd.Stderr = io.MultiWriter(writer, os.Stderr)

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

func main() {
	insecure := flag.BoolP("insecure", "k", false, "allows insecure SSL connections")
	timeout := flag.String("connect-timeout", "5", "max number of seconds to connect to busl URL")

	if flag.Parse(); len(flag.Args()) < 2 {
		println(usage)
		os.Exit(1)
	}

	reader, writer := io.Pipe()
	uploaded := make(chan struct{})

	url := flag.Arg(0)
	args := flag.Args()[1:]

	go func() {
		if err := stream(url, reader, *insecure, *timeout); err != nil {
			// Prevent writes from blocking.
			io.Copy(ioutil.Discard, reader)
		}
		close(uploaded)
	}()

	err := run(args, writer)

	// Wait for http request to complete
	<-uploaded

	if err != nil {
		os.Exit(exitStatus(err))
	}
}

var usage = "Usage: busltee <url> [-k|--insecure] [--connect-timeout N] -- <command>"
