// Copyright 2014 Apcera Inc. All rights reserved.

package server

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/nats-io/gnatsd/logger"
)

func TestSetLogger(t *testing.T) {
	server := &Server{}
	defer server.SetLogger(nil, false, false)
	dl := &DummyLogger{}
	server.SetLogger(dl, true, true)

	// We assert that the logger has change to the DummyLogger
	_ = log.logger.(*DummyLogger)

	if debug != 1 {
		t.Fatalf("Expected debug 1, received value %d\n", debug)
	}

	if trace != 1 {
		t.Fatalf("Expected trace 1, received value %d\n", trace)
	}

	// Check traces
	expectedStr := "This is a Notice"
	Noticef(expectedStr)
	dl.checkContent(t, expectedStr)
	expectedStr = "This is an Error"
	Errorf(expectedStr)
	dl.checkContent(t, expectedStr)
	expectedStr = "This is a Fatal"
	Fatalf(expectedStr)
	dl.checkContent(t, expectedStr)
	expectedStr = "This is a Debug"
	Debugf(expectedStr)
	dl.checkContent(t, expectedStr)
	expectedStr = "This is a Trace"
	Tracef(expectedStr)
	dl.checkContent(t, expectedStr)

	// Make sure that we can reset to fal
	server.SetLogger(dl, false, false)
	if debug != 0 {
		t.Fatalf("Expected debug 0, got %v", debug)
	}
	if trace != 0 {
		t.Fatalf("Expected trace 0, got %v", trace)
	}
	// Now, Debug and Trace should not produce anything
	dl.msg = ""
	Debugf("This Debug should not be traced")
	dl.checkContent(t, "")
	Tracef("This Trace should not be traced")
	dl.checkContent(t, "")
}

type DummyLogger struct {
	msg string
}

func (dl *DummyLogger) checkContent(t *testing.T, expectedStr string) {
	if dl.msg != expectedStr {
		stackFatalf(t, "Expected log to be: %v, got %v", expectedStr, dl.msg)
	}
}

func (l *DummyLogger) Noticef(format string, v ...interface{}) {
	l.msg = fmt.Sprintf(format, v...)
}
func (l *DummyLogger) Errorf(format string, v ...interface{}) {
	l.msg = fmt.Sprintf(format, v...)
}
func (l *DummyLogger) Fatalf(format string, v ...interface{}) {
	l.msg = fmt.Sprintf(format, v...)
}
func (l *DummyLogger) Debugf(format string, v ...interface{}) {
	l.msg = fmt.Sprintf(format, v...)
}
func (l *DummyLogger) Tracef(format string, v ...interface{}) {
	l.msg = fmt.Sprintf(format, v...)
}

func TestReOpenLogFile(t *testing.T) {
	// We can't rename the file log when still opened on Windows, so skip
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}
	s := &Server{opts: &Options{}}
	defer s.SetLogger(nil, false, false)

	// First check with no logger
	s.SetLogger(nil, false, false)
	s.ReOpenLogFile()

	// Then when LogFile is not provided.
	dl := &DummyLogger{}
	s.SetLogger(dl, false, false)
	s.ReOpenLogFile()
	dl.checkContent(t, "File log re-open ignored, not a file logger")

	// Set a File log
	s.opts.LogFile = "test.log"
	defer os.Remove(s.opts.LogFile)
	defer os.Remove(s.opts.LogFile + ".bak")
	fileLog := logger.NewFileLogger(s.opts.LogFile, s.opts.Logtime, s.opts.Debug, s.opts.Trace, true)
	s.SetLogger(fileLog, false, false)
	// Add some log
	expectedStr := "This is a Notice"
	Noticef(expectedStr)
	// Check content of log
	buf, err := ioutil.ReadFile(s.opts.LogFile)
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}
	if !strings.Contains(string(buf), expectedStr) {
		t.Fatalf("Expected log to contain: %q, got %q", expectedStr, string(buf))
	}
	// Close the file and rename it
	if err := os.Rename(s.opts.LogFile, s.opts.LogFile+".bak"); err != nil {
		t.Fatalf("Unable to rename log file: %v", err)
	}
	// Now re-open LogFile
	s.ReOpenLogFile()
	// Content should indicate that we have re-opened the log
	buf, err = ioutil.ReadFile(s.opts.LogFile)
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}
	if strings.HasSuffix(string(buf), "File log-reopened") {
		t.Fatalf("File should indicate that file log was re-opened, got: %v", string(buf))
	}
	// Make sure we can append to the log
	Noticef("New message")
	buf, err = ioutil.ReadFile(s.opts.LogFile)
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}
	if strings.HasSuffix(string(buf), "New message") {
		t.Fatalf("New message was not appended after file was re-opened, got: %v", string(buf))
	}
}
