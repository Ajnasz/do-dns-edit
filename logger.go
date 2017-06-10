package main

import "log"
import "os"

// Logger logs to std out/err
type Logger struct{}

// Log prints to stdout
func (l Logger) Log(args ...interface{}) {
	log.SetOutput(os.Stdout)
	log.Println(args...)
}

// Error prints to stderr
func (l Logger) Error(args ...interface{}) {
	log.SetOutput(os.Stderr)
	log.Println(args...)
}

// Fatal prints to stderr and exits
func (l Logger) Fatal(args ...interface{}) {
	log.SetOutput(os.Stderr)
	log.Fatal(args...)
}

// Fatalf prints to stderr and exits
func (l Logger) Fatalf(s string, args ...interface{}) {
	log.SetOutput(os.Stderr)
	log.Fatalf(s, args...)
}
