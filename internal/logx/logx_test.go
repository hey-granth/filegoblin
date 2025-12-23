package logx

import (
	"bytes"
	"strings"
	"testing"
)

// go convention method states that test methods start with Test and take a single argument of type *testing.T
// the "t" argument provides methods for reporting test failures and logging during tests.
// TestLoggerInfoAndError tests that Info and Error log messages are formatted correctly.
func TestLoggerInfoAndError(t *testing.T) {
	var buf bytes.Buffer // declares a bytes.Buffer to capture log output for inspection in memory instead of writing in stdout or a file.
	logger := New(&buf)  // creates a new Logger that writes to the buffer by passing a pointer to buf.

	logger.Info("hello %s", "world")
	logger.Error("failed: %d", 42)

	out := buf.String()
	if !strings.Contains(out, "INFO") {
		t.Fatalf("expected INFO in log output; got: %q", out)
	}
	if !strings.Contains(out, "ERROR") {
		t.Fatalf("expected ERROR in log output; got: %q", out)
	}
	if !strings.Contains(out, "hello world") {
		t.Fatalf("expected message body 'hello world'; got: %q", out)
	}
}

// t.Fatalf is used to log a formatted error message and stop the test immediately if a condition is not met.
