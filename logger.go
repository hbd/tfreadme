package main

import (
	"log"
)

type cliLogger struct {
	log.Logger
}

func (l *cliLogger) debugf(format string, v ...interface{}) {
	if verboseFlag {
		log.Printf(format, v...)
	}
}
