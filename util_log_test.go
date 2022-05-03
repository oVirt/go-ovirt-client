package ovirtclient_test

import (
	"log"
	"testing"
)

func testLogger(t *testing.T) *log.Logger {
	return log.New(testLogWriter{t}, "", 0)
}

type testLogWriter struct {
	t *testing.T
}

func (t testLogWriter) Write(p []byte) (n int, err error) {
	t.t.Log(string(p))
	return len(p), nil
}
