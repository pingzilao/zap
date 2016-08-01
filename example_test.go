// Copyright (c) 2016 Uber Technologies, Inc.
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

package zap_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/uber-go/zap"
)

func Example() {
	// Log in JSON, using zap's reflection-free JSON encoder.
	// The default options will log any Info or higher logs to standard out.
	logger := zap.NewJSON()
	// For repeatable tests, pretend that it's always 1970.
	logger.StubTime()

	logger.Warn("Log without structured data...")
	logger.Warn(
		"Or use strongly-typed wrappers to add structured context.",
		zap.String("library", "zap"),
		zap.Duration("latency", time.Nanosecond),
	)

	// Avoid re-serializing the same data repeatedly by creating a child logger
	// with some attached context. That context is added to all the child's
	// log output, but doesn't affect the parent.
	child := logger.With(zap.String("user", "jane@test.com"), zap.Int("visits", 42))
	child.Error("Oh no!")

	// Output:
	// {"level":"warn","ts":0,"msg":"Log without structured data..."}
	// {"level":"warn","ts":0,"msg":"Or use strongly-typed wrappers to add structured context.","library":"zap","latency":1}
	// {"level":"error","ts":0,"msg":"Oh no!","user":"jane@test.com","visits":42}
}

func Example_fileOutput() {
	// Create a temporary file to output logs to.
	f, err := ioutil.TempFile("", "log")
	if err != nil {
		panic("failed to create temporary file")
	}
	defer os.Remove(f.Name())

	logger := zap.NewJSON(
		// Write the logging output to the specified file instead of stdout.
		// Any type implementing zap.WriteSyncer or zap.WriteFlusher can be used.
		zap.Output(f),
	)
	// Stub the current time in tests.
	logger.StubTime()

	logger.Info("This is an info log.", zap.Int("foo", 42))

	// Sync the file so logs are written to disk, and print the file contents.
	// zap will call Sync automatically when logging at FatalLevel or PanicLevel.
	f.Sync()
	contents, err := ioutil.ReadFile(f.Name())
	if err != nil {
		panic("failed to read temporary file")
	}

	fmt.Println(string(contents))
	// Output:
	// {"level":"info","ts":0,"msg":"This is an info log.","foo":42}
}

func ExampleNest() {
	logger := zap.NewJSON()
	// Stub the current time in tests.
	logger.StubTime()

	// We'd like the logging context to be {"outer":{"inner":42}}
	nest := zap.Nest("outer", zap.Int("inner", 42))
	logger.Info("Logging a nested field.", nest)

	// Output:
	// {"level":"info","ts":0,"msg":"Logging a nested field.","outer":{"inner":42}}
}

func ExampleNewJSON() {
	// The default logger outputs to standard out and only writes logs that are
	// Info level or higher.
	logger := zap.NewJSON()
	// Stub the current time in tests.
	logger.StubTime()

	// The default logger does not print Debug logs.
	logger.Debug("This won't be printed.")
	logger.Info("This is an info log.")

	// Output:
	// {"level":"info","ts":0,"msg":"This is an info log."}
}

func ExampleNewJSON_options() {
	// We can pass multiple options to the NewJSON method to configure
	// the logging level, output location, or even the initial context.
	logger := zap.NewJSON(
		zap.DebugLevel,
		zap.Fields(zap.Int("count", 1)),
	)
	// Stub the current time in tests.
	logger.StubTime()

	logger.Debug("This is a debug log.")
	logger.Info("This is an info log.")

	// Output:
	// {"level":"debug","ts":0,"msg":"This is a debug log.","count":1}
	// {"level":"info","ts":0,"msg":"This is an info log.","count":1}
}

func ExampleCheckedMessage() {
	logger := zap.NewJSON()
	// Stub the current time in tests.
	logger.StubTime()

	// By default, the debug logging level is disabled. However, calls to
	// logger.Debug will still allocate a slice to hold any passed fields.
	// Particularly performance-sensitive applications can avoid paying this
	// penalty by using checked messages.
	if cm := logger.Check(zap.DebugLevel, "This is a debug log."); cm.OK() {
		// Debug-level logging is disabled, so we won't get here.
		cm.Write(zap.Int("foo", 42), zap.Stack())
	}

	if cm := logger.Check(zap.InfoLevel, "This is an info log."); cm.OK() {
		// Since info-level logging is enabled, we expect to write out this message.
		cm.Write()
	}

	// Output:
	// {"level":"info","ts":0,"msg":"This is an info log."}
}

func ExampleLevel_MarshalText() {
	level := zap.ErrorLevel
	s := struct {
		Level *zap.Level `json:"level"`
	}{&level}
	bytes, _ := json.Marshal(s)
	fmt.Println(string(bytes))

	// Output:
	// {"level":"error"}
}

func ExampleLevel_UnmarshalText() {
	var s struct {
		Level zap.Level `json:"level"`
	}
	// The zero value for a zap.Level is zap.InfoLevel.
	fmt.Println(s.Level)

	json.Unmarshal([]byte(`{"level":"error"}`), &s)
	fmt.Println(s.Level)

	// Output:
	// info
	// error
}

func ExampleNewJSONEncoder() {
	// An encoder with the default settings.
	zap.NewJSONEncoder()

	// Dropping timestamps is often useful in tests.
	zap.NewJSONEncoder(zap.NoTime())

	// In production, customize the encoder to work with your log aggregation
	// system.
	zap.NewJSONEncoder(
		zap.RFC3339Formatter("@timestamp"), // human-readable timestamps
		zap.MessageKey("@message"),         // customize the message key
		zap.LevelString("@level"),          // stringify the log level
	)
}
