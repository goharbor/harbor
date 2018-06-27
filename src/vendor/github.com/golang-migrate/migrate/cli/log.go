package main

import (
	"fmt"
	logpkg "log"
	"os"
)

type Log struct {
	verbose bool
}

func (l *Log) Printf(format string, v ...interface{}) {
	if l.verbose {
		logpkg.Printf(format, v...)
	} else {
		fmt.Fprintf(os.Stderr, format, v...)
	}
}

func (l *Log) Println(args ...interface{}) {
	if l.verbose {
		logpkg.Println(args...)
	} else {
		fmt.Fprintln(os.Stderr, args...)
	}
}

func (l *Log) Verbose() bool {
	return l.verbose
}

func (l *Log) fatalf(format string, v ...interface{}) {
	l.Printf(format, v...)
	os.Exit(1)
}

func (l *Log) fatal(args ...interface{}) {
	l.Println(args...)
	os.Exit(1)
}

func (l *Log) fatalErr(err error) {
	l.fatal("error:", err)
}
