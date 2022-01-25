// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package goleak

import (
	"fmt"
	"io"
	"os"
)

// Variables for stubbing in unit tests.
var (
	_osExit             = os.Exit
	_osStderr io.Writer = os.Stderr
)

// TestingM is the minimal subset of testing.M that we use.
type TestingM interface {
	Run() int
}

// VerifyTestMain can be used in a TestMain function for package tests to
// verify that there were no goroutine leaks.
// To use it, your TestMain function should look like:
//
//  func TestMain(m *testing.M) {
//    goleak.VerifyTestMain(m)
//  }
//
// See https://golang.org/pkg/testing/#hdr-Main for more details.
//
// This will run all tests as per normal, and if they were successful, look
// for any goroutine leaks and fail the tests if any leaks were found.
func VerifyTestMain(m TestingM, options ...Option) {
	_osExit(verifyTestMain(m, options...))
}

func verifyTestMain(m TestingM, options ...Option) int {
	exitCode := m.Run()
	if exitCode != 0 {
		// If there are any failures, return without checking leaks (test has already failed).
		return exitCode
	}

	err := Find(options...)
	if err == nil {
		// No failures + no leaks.
		return 0
	}

	// Check if the Find error was caused by a leak (non-zero exit), or a stack parsing failure (exit code 0)
	if _, ok := err.(stackParseErr); ok {
		fmt.Fprintf(_osStderr, "goleak: skipped due to stack parsing failures: %v\n", err)
		return 0
	} else {
		fmt.Fprintf(_osStderr, "goleak: failed due to stack parsing failures: %v\n", err)
		return 1
	}
}
