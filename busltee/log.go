package busltee

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

var out io.Writer

// OpenLogs configures the log file
func OpenLogs(logFile, logPrefix string) {
	out = output(logFile)

	log.SetPrefix(logPrefix + " ")
	log.SetOutput(out)
	log.SetFlags(0)
}

// CloseLogs closes an open log file
func CloseLogs() {
	if f, ok := out.(io.Closer); ok {
		f.Close()
	}
}

func output(logFile string) io.Writer {
	if logFile == "" {
		return ioutil.Discard
	}
	if file, err := os.OpenFile(logFile, os.O_RDWR|os.O_APPEND, 0660); err != nil {
		return ioutil.Discard
	}
	return file
}
