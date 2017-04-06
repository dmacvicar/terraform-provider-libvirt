// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"bytes"
	"fmt"
	"log/syslog"
	"os/exec"
	"strings"
	"syscall"
)

type LoggerOps interface {
	Emerg(string) error
	Alert(string) error
	Crit(string) error
	Err(string) error
	Warning(string) error
	Notice(string) error
	Info(string) error
	Debug(string) error
	Close() error
}

// Logger implements a variadic flavor of log/syslog.Writer
type Logger struct {
	ops           LoggerOps
	prefixStack   []string
	opSequenceNum int
}

// New creates a new logger.
// syslog is tried first, if syslog fails Stdout is used.
func New() Logger {
	logger := Logger{}
	if slogger, err := syslog.New(syslog.LOG_DEBUG, "ignition"); err == nil {
		logger.ops = slogger
	} else {
		logger.ops = Stdout{}
		logger.Err("unable to open syslog: %v", err)
	}
	return logger
}

// Close closes the logger.
func (l Logger) Close() {
	l.ops.Close()
}

// Emerg logs a message at emergency priority.
func (l Logger) Emerg(format string, a ...interface{}) error {
	return l.log(l.ops.Emerg, format, a...)
}

// Alert logs a message at alert priority.
func (l Logger) Alert(format string, a ...interface{}) error {
	return l.log(l.ops.Alert, format, a...)
}

// Crit logs a message at critical priority.
func (l Logger) Crit(format string, a ...interface{}) error {
	return l.log(l.ops.Crit, format, a...)
}

// Err logs a message at error priority.
func (l Logger) Err(format string, a ...interface{}) error {
	return l.log(l.ops.Err, format, a...)
}

// Warning logs a message at warning priority.
func (l Logger) Warning(format string, a ...interface{}) error {
	return l.log(l.ops.Warning, format, a...)
}

// Notice logs a message at notice priority.
func (l Logger) Notice(format string, a ...interface{}) error {
	return l.log(l.ops.Notice, format, a...)
}

// Info logs a message at info priority.
func (l Logger) Info(format string, a ...interface{}) error {
	return l.log(l.ops.Info, format, a...)
}

// Debug logs a message at debug priority.
func (l Logger) Debug(format string, a ...interface{}) error {
	return l.log(l.ops.Debug, format, a...)
}

// PushPrefix pushes the supplied message onto the Logger's prefix stack.
// The prefix stack is concatenated in FIFO order and prefixed to the start of every message logged via Logger.
func (l *Logger) PushPrefix(format string, a ...interface{}) {
	l.prefixStack = append(l.prefixStack, fmt.Sprintf(format, a...))
}

// PopPrefix pops the top entry from the Logger's prefix stack.
// The prefix stack is concatenated in FIFO order and prefixed to the start of every message logged via Logger.
func (l *Logger) PopPrefix() {
	if len(l.prefixStack) == 0 {
		l.Debug("popped from empty stack")
		return
	}
	l.prefixStack = l.prefixStack[:len(l.prefixStack)-1]
}

// quotedCmd returns a concatenated, quoted form of cmd's cmdline
func quotedCmd(cmd *exec.Cmd) string {
	if len(cmd.Args) == 0 {
		return fmt.Sprintf("%q", cmd.Path)
	}

	var q []string
	for _, s := range cmd.Args {
		q = append(q, fmt.Sprintf("%q", s))
	}

	return strings.Join(q, ` `)
}

// LogCmd runs and logs the supplied cmd as an operation with distinct start/finish/fail log messages uniformly combined with the supplied format string.
// The exact command path and arguments being executed are also logged for debugging assistance.
func (l *Logger) LogCmd(cmd *exec.Cmd, format string, a ...interface{}) (int, error) {
	code := -1
	f := func() error {
		cmdLine := quotedCmd(cmd)
		l.Debug("executing: %s", cmdLine)

		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				code = exitErr.Sys().(syscall.WaitStatus).ExitStatus()
			}
			return fmt.Errorf("%v: Cmd: %s Stdout: %q Stderr: %q", err, cmdLine, stdout.Bytes(), stderr.Bytes())
		}
		return nil
	}
	err := l.LogOp(f, format, a...)
	return code, err
}

// LogOp calls and logs the supplied function as an operation with distinct start/finish/fail log messages uniformly combined with the supplied format string.
func (l *Logger) LogOp(op func() error, format string, a ...interface{}) error {
	l.opSequenceNum++
	l.PushPrefix("op(%x)", l.opSequenceNum)
	defer l.PopPrefix()

	l.logStart(format, a...)
	if err := op(); err != nil {
		l.logFail("%s: %v", fmt.Sprintf(format, a...), err)
		return err
	}
	l.logFinish(format, a...)
	return nil
}

// logStart logs the start of a multi-step/substantial/time-consuming operation.
func (l Logger) logStart(format string, a ...interface{}) {
	l.Info(fmt.Sprintf("[started]  %s", format), a...)
}

// logFail logs the failure of a multi-step/substantial/time-consuming operation.
func (l Logger) logFail(format string, a ...interface{}) {
	l.Crit(fmt.Sprintf("[failed]   %s", format), a...)
}

// logFinish logs the completion of a multi-step/substantial/time-consuming operation.
func (l Logger) logFinish(format string, a ...interface{}) {
	l.Info(fmt.Sprintf("[finished] %s", format), a...)
}

// log logs a formatted message using the supplied logFunc.
func (l Logger) log(logFunc func(string) error, format string, a ...interface{}) error {
	return logFunc(l.sprintf(format, a...))
}

// sprintf returns the current prefix stack, if any, concatenated with the supplied format string and args in expanded form.
func (l Logger) sprintf(format string, a ...interface{}) string {
	m := []string{}
	for _, pfx := range l.prefixStack {
		m = append(m, fmt.Sprintf("%s:", pfx))
	}
	m = append(m, fmt.Sprintf(format, a...))
	return strings.Join(m, " ")
}
