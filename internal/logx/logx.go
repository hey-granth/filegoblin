package logx

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// Logger is a tiny wrapper so tests can inspect output if needed.
type Logger struct {
	std *log.Logger // This holds the *log.Logger used to format and write messages. It's a pointer so methods and internal state are shared, not copied.
	mu  sync.Mutex  // This Mutex is locked around write operations (see Info/Error) so multiple goroutines don't interleave log output.
	// Mutex (mutual exclusion) is a synchronization primitive that ensures only one goroutine at a time can execute a "critical section" of code that accesses shared state
	out io.Writer // This stores the io.Writer (for example os.Stdout or a file) the logger writes to; it’s exposed by the Writer() method so callers can inspect or reuse it.
}

// this is a constructor for the Logger type. It creates and returns a new *Logger configured to write to the given io.Writer, defaulting to standard output when nil.
func New(w io.Writer) *Logger {
	if w == nil { // this sets up a default log writer, like if the value of w is passed to be null, the logs will be diplayed into the terminal
		w = os.Stdout
	}
	std := log.New(w, "", 0) // use of standard library logger
	// we are passing empty string as a prefix coz we are modifying it later on with [INFO] and [ERROR] tags in the Info and Error methods respectively later on.
	// the standard log library provides a lot of flags which can be displayed in the logs, such as date, time, etc with the logs.
	// This builds a *log.Logger that writes to w with no prefix and no flags — formatting (timestamp, level) is handled by your wrapper, not the standard logger.

	// This returns a heap-allocated *Logger containing the internal *log.Logger and the io.Writer used. Using a pointer means shared internal state (like the mutex) behaves correctly when the logger is used across goroutines.
	return &Logger{ // returns the pointer to the new logger object.
		std: std,
		out: w,
	}
}

// like we do self in python functions and methods, we do (l *Logger) in golang.
// we use pointer so we can later lock the actual mutex and ensure thread safety, instead of a copy.
// The ... makes this variadic (like Python's *args). interface{} is Go's "any type" - equivalent to Python's Any or just not type-hinting. So this accepts zero or more arguments of any type.
func (l *Logger) Info(format string, v ...interface{}) {
	l.mu.Lock()                      // this locks the mutex to ensure that only one goroutine can execute the following code block at a time, preventing interleaved log output.
	defer l.mu.Unlock()              // this schedules the unlock to happen when the function returns, ensuring the mutex is always released.
	msg := fmt.Sprintf(format, v...) // this formats the log message using the provided format string and arguments. v... unpacks the variadic arguments. for example, if format is "Hello %s" and v is ["World"], msg becomes "Hello World".
	// escape newlines and carriage returns to prevent log injection / header spoofing
	msg = strings.ReplaceAll(msg, "\n", "\\n")                           // this escapes newlines in the message to avoid log injection. for example, if msg is "Hello\nWorld", it becomes "Hello\\nWorld".
	msg = strings.ReplaceAll(msg, "\r", "\\r")                           // this escapes carriage returns in the message to avoid log injection. for example, if msg is "Hello\rWorld", it becomes "Hello\\rWorld".
	l.std.Printf("%s [INFO] %s\n", time.Now().Format(time.RFC3339), msg) // this prints the formatted log message to the logger's output, prefixed with the current time and the [INFO] tag.
}

// same as Info method but for error level logs.
func (l *Logger) Error(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	msg := fmt.Sprintf(format, v...)
	msg = strings.ReplaceAll(msg, "\n", "\\n")
	msg = strings.ReplaceAll(msg, "\r", "\\r")
	l.std.Printf("%s [ERROR] %s\n", time.Now().Format(time.RFC3339), msg)
}

// Writer returns the io.Writer the logger writes to. This lets callers inspect or reuse the underlying writer if needed.
// io.Writer is an interface which is written in a syntax to define the return type of the function.
func (l *Logger) Writer() io.Writer { return l.out } // this exposes the raw writer used by the logger.

//Why expose the raw writer:
//Some code needs to write to the same destination as the logger but without the timestamp/level formatting. Common scenarios:
//1. Interface compatibility: Many Go functions accept io.Writer as a parameter. If you want to redirect their output to your log file, you pass logger.Writer(). Example: json.NewEncoder(logger.Writer()).Encode(data) - writes JSON directly to the log without "[INFO]" prefixes.
//2. Performance-critical paths: The formatted logging (Info, Error) does string allocation, formatting, timestamp generation. If you're dumping binary data or high-volume output, direct writes skip that overhead.
//3. Third-party library integration: Libraries that expect an io.Writer for their output (HTTP response recorders, template engines, streaming parsers) can write to your log destination without modification.

// GOROUTINE A goroutine is a lightweight, user-space thread managed by the Go runtime. It lets you run functions concurrently using the go keyword. Goroutines are cheap to create, multiplexed onto OS threads by the Go scheduler, and can run in parallel on multiple CPU cores.
