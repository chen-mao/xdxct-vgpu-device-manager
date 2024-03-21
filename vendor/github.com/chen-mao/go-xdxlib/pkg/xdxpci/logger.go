package xdxpci

import (
	"log"
)

type logger interface {
	Warningf(string, ...interface{})
}

type simpleLogger struct{}

func (l simpleLogger) Warningf(format string, v ...interface{}) {
	log.Printf("WARNING: "+format, v)
}
