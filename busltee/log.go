package busltee

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

var logOutput io.Writer

func OpenLogs(logFile, logPrefix string) {
	logOutput = getLogOutput(logFile)

	log.SetPrefix(logPrefix + " ")
	log.SetOutput(logOutput)
	log.SetFlags(0)
}

func CloseLogs() {
	if f, ok := logOutput.(io.Closer); ok {
		f.Close()
	}
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
